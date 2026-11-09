package rest

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/cloud-barista/ai-ha/pkg/framework/agent"
	"github.com/cloud-barista/ai-ha/pkg/framework/inventory"
	"github.com/cloud-barista/ai-ha/pkg/framework/job"
	"github.com/cloud-barista/ai-ha/pkg/framework/repository"
	"github.com/cloud-barista/ai-ha/pkg/framework/restore"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Bounds of a list page. A history can grow without limit, so a caller gets a
// page rather than everything, whatever it asks for.
const (
	defaultListLimit = 50
	maxListLimit     = 200
)

// mgmtHandler serves the management API: what operators, the portal and the
// HA support agent use. Agents use the routes under /agent instead.
type mgmtHandler struct {
	agents    *agent.Service
	jobs      *job.Service
	inventory *inventory.Service
	repos     *repository.Service
	restores  *restore.Service
	// access is handed to whoever has to configure an agent.
	access model.FrameworkAccess
}

// registerManagementRoutes adds the management API. Without an authenticator
// it adds nothing: there is no unauthenticated mode to fall into.
//
// Routes whose work is not implemented are registered regardless, so that the
// whole contract is reachable and a client gets 501 rather than 404.
func registerManagementRoutes(g *echo.Group, cfg Config) {
	if cfg.AdminAuth == nil {
		return
	}
	// Attached route by route rather than to a group: a group's middleware
	// also runs for paths that match nothing, which would answer an unknown
	// /aiha path with 401 instead of 404.
	auth := adminAuth(cfg.AdminAuth)
	h := &mgmtHandler{
		agents:    cfg.Agent,
		jobs:      cfg.Jobs,
		inventory: cfg.Inventory,
		repos:     cfg.Repositories,
		restores:  cfg.Restore,
		access:    cfg.FrameworkAccess,
	}

	if cfg.Agent != nil {
		g.GET("/systems", h.RestGetSystems, auth)
		g.GET("/systems/:systemId", h.RestGetSystem, auth)
		g.GET("/agents", h.RestGetAgents, auth)
		g.POST("/agents/:agentId/revoke", h.RestPostAgentRevoke, auth)
		g.GET("/registration-tokens", h.RestGetRegistrationTokens, auth)
		g.POST("/registration-tokens", h.RestPostRegistrationToken, auth)
		g.POST("/registration-tokens/:tokenId/revoke", h.RestPostRegistrationTokenRevoke, auth)
	}
	if cfg.Inventory != nil {
		g.GET("/systems/:systemId/inventory", h.RestGetSystemInventory, auth)
	}
	if cfg.Agent != nil && cfg.Inventory != nil && cfg.Jobs != nil {
		g.POST("/systems/:systemId/backups", h.RestPostSystemBackup, auth)
	}
	if cfg.Jobs != nil {
		g.GET("/backups", h.RestGetBackups, auth)
		g.GET("/backups/:backupSetId", h.RestGetBackup, auth)
		g.GET("/jobs", h.RestGetJobs, auth)
		g.GET("/jobs/:jobId", h.RestGetJob, auth)
		g.POST("/jobs/:jobId/cancel", h.RestPostJobCancel, auth)
	}
	if cfg.Restore != nil {
		g.POST("/restores", h.RestPostRestore, auth)
		g.GET("/restores", h.RestGetRestores, auth)
		g.GET("/restores/:restoreId", h.RestGetRestore, auth)
	}
	if cfg.Repositories != nil {
		g.GET("/repositories", h.RestGetRepositories, auth)
		g.POST("/repositories", h.RestPostRepository, auth)
		g.GET("/repositories/:repoId", h.RestGetRepository, auth)
		g.PUT("/repositories/:repoId/credentials", h.RestPutRepositoryCredentials, auth)
	}

	g.POST("/systems/:systemId/inventory/refresh", RestPostSystemInventoryRefresh, auth)
	g.PUT("/systems/:systemId/targets", RestPutSystemTargets, auth)
	g.DELETE("/backups/:backupSetId", RestDeleteBackup, auth)
	g.POST("/restores/:restoreId/cancel", RestPostRestoreCancel, auth)
	g.POST("/repositories/:repoId/verify", RestPostRepositoryVerify, auth)

	g.GET("/ha/groups", RestGetHAGroups, auth)
	g.POST("/ha/groups", RestPostHAGroup, auth)
	g.GET("/ha/groups/:groupId", RestGetHAGroup, auth)
	g.POST("/ha/groups/:groupId/setup", RestPostHAGroupSetup, auth)
	g.POST("/ha/groups/:groupId/teardown", RestPostHAGroupTeardown, auth)
	g.POST("/ha/groups/:groupId/switchover", RestPostHAGroupSwitchover, auth)
	g.POST("/ha/groups/:groupId/failover", RestPostHAGroupFailover, auth)
	g.POST("/ha/groups/:groupId/resume", RestPostHAGroupResume, auth)
	g.POST("/ha/groups/:groupId/rejoin", RestPostHAGroupRejoin, auth)
	g.POST("/ha/groups/:groupId/acknowledge", RestPostHAGroupAcknowledge, auth)
	g.GET("/ha/groups/:groupId/transitions", RestGetHAGroupTransitions, auth)
	g.GET("/ha/events", RestGetHAEvents, auth)

	g.GET("/incidents", RestGetIncidents, auth)
	g.POST("/incidents", RestPostIncident, auth)
	g.GET("/incidents/:incidentId", RestGetIncident, auth)
	g.POST("/incidents/:incidentId/events", RestPostIncidentEvent, auth)
	g.POST("/incidents/:incidentId/diagnoses", RestPostIncidentDiagnosis, auth)
	g.POST("/incidents/:incidentId/alarms", RestPostIncidentAlarm, auth)
	g.POST("/incidents/:incidentId/actions", RestPostIncidentAction, auth)
	g.GET("/history", RestGetHistory, auth)
}

// notImplemented answers a route in the contract whose work does not exist yet.
func notImplemented(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented,
		model.SimpleMsg{Message: "Not implemented yet", Code: model.CodeNotImplemented})
}

// refuse logs why a request was turned down and answers it.
//
// msg is written for the caller and err is for the log: a library's error text
// never reaches a response. err may be nil when msg says all there is to say.
func refuse(c echo.Context, err error, status int, code, msg string) error {
	evt := log.Warn()
	if status >= http.StatusInternalServerError {
		evt = log.Error()
	}
	evt.Err(err).Str("route", c.Path()).Int("status", status).
		Str("requestId", c.Response().Header().Get(echo.HeaderXRequestID)).
		Msg("Management request refused")
	return c.JSON(status, model.SimpleMsg{Message: msg, Code: code})
}

// refuseMalformed answers a body that could not be read. The parser's error is
// not logged: it can quote the request, and a request can carry credentials.
func refuseMalformed(c echo.Context) error {
	return refuse(c, nil, http.StatusBadRequest, model.CodeInvalidRequest,
		"Malformed request body: check JSON syntax")
}

// listParams are the query parameters shared by the list routes.
type listParams struct {
	limit  int
	offset int
	from   time.Time
	to     time.Time
}

// listParamsError carries a message written for the caller.
type listParamsError string

func (e listParamsError) Error() string { return string(e) }

// parseListParams reads limit, offset, from and to, refusing rather than
// ignoring a value it cannot use: a history silently left unfiltered reads as
// an answer to the question that was asked.
func parseListParams(c echo.Context) (listParams, error) {
	p := listParams{limit: defaultListLimit}

	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > maxListLimit {
			return listParams{}, listParamsError("Limit must be between 1 and " + strconv.Itoa(maxListLimit))
		}
		p.limit = n
	}
	if raw := c.QueryParam("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return listParams{}, listParamsError("Offset must not be negative")
		}
		p.offset = n
	}

	var err error
	if p.from, err = parseTimeParam(c, "from"); err != nil {
		return listParams{}, err
	}
	if p.to, err = parseTimeParam(c, "to"); err != nil {
		return listParams{}, err
	}
	if !p.from.IsZero() && !p.to.IsZero() && !p.from.Before(p.to) {
		return listParams{}, listParamsError("The period must start before it ends")
	}
	return p, nil
}

// parseTimeParam reads an optional RFC 3339 time.
func parseTimeParam(c echo.Context, name string) (time.Time, error) {
	raw := c.QueryParam(name)
	if raw == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, listParamsError("Time must be RFC 3339, e.g. 2026-11-03T00:00:00Z")
	}
	return t, nil
}

// refuseListParams answers a query parseListParams turned down.
func refuseListParams(c echo.Context, err error) error {
	var params listParamsError
	if errors.As(err, &params) {
		return refuse(c, nil, http.StatusBadRequest, model.CodeInvalidRequest, string(params))
	}
	return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Request failed")
}
