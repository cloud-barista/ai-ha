package rest

import (
	"errors"
	"net/http"

	"github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/cloud-barista/ai-ha/pkg/framework/agent"
	"github.com/cloud-barista/ai-ha/pkg/framework/inventory"
	"github.com/labstack/echo/v4"
)

// RestGetSystems godoc
// @ID GetSystems
// @Summary List protected systems
// @Description List the systems an agent has enrolled for, most recently enrolled first, with the state of the agent speaking for each. An online agent is reachable; that does not say its work is healthy.
// @Tags [Protection] System Management
// @Produce json
// @Success 200 {object} model.ProtectedSystemListResp
// @Failure 401 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /systems [get]
func (h *mgmtHandler) RestGetSystems(c echo.Context) error {
	systems, err := h.agents.ListSystems(c.Request().Context())
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Systems could not be listed")
	}
	items := make([]model.ProtectedSystem, 0, len(systems))
	for _, s := range systems {
		items = append(items, protectedSystem(s))
	}
	return c.JSON(http.StatusOK, model.ProtectedSystemListResp{Items: items})
}

// RestGetSystem godoc
// @ID GetSystem
// @Summary Get a protected system
// @Description Return one protected system and how current its inventory is. The inventory is absent until the agent first reports.
// @Tags [Protection] System Management
// @Produce json
// @Param systemId path string true "System ID"
// @Success 200 {object} model.ProtectedSystemDetail
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /systems/{systemId} [get]
func (h *mgmtHandler) RestGetSystem(c echo.Context) error {
	ctx := c.Request().Context()

	system, err := h.agents.GetSystem(ctx, c.Param("systemId"))
	if errors.Is(err, agent.ErrSystemNotFound) {
		return refuse(c, nil, http.StatusNotFound, model.CodeNotFound, "System not found")
	}
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "System could not be read")
	}

	detail := model.ProtectedSystemDetail{ProtectedSystem: protectedSystem(system)}
	if h.inventory != nil {
		stored, err := h.inventory.Get(ctx, system.SystemID)
		switch {
		case err == nil:
			summary := h.inventorySummary(stored)
			detail.Inventory = &summary
		case errors.Is(err, inventory.ErrNoInventory):
		default:
			return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "System could not be read")
		}
	}
	return c.JSON(http.StatusOK, detail)
}

// RestGetSystemInventory godoc
// @ID GetSystemInventory
// @Summary Get what a system is made of
// @Description Return what the system's agent last reported. The framework cannot read a system itself, so this is always a past reading: check when it was collected before relying on it.
// @Tags [Protection] System Management
// @Produce json
// @Param systemId path string true "System ID"
// @Success 200 {object} model.SystemInventoryInfo
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /systems/{systemId}/inventory [get]
func (h *mgmtHandler) RestGetSystemInventory(c echo.Context) error {
	stored, err := h.inventory.Get(c.Request().Context(), c.Param("systemId"))
	if errors.Is(err, inventory.ErrNoInventory) {
		return refuse(c, nil, http.StatusNotFound, model.CodeNoInventory, "No inventory reported for this system")
	}
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Inventory could not be read")
	}
	return c.JSON(http.StatusOK, model.SystemInventoryInfo{
		SystemID:         stored.SystemID,
		AgentID:          stored.AgentID,
		InventorySummary: h.inventorySummary(stored),
		AgeSeconds:       int64(stored.Age().Seconds()),
		Report:           stored.Inventory,
	})
}

// RestPostSystemInventoryRefresh godoc
// @ID PostSystemInventoryRefresh
// @Summary Ask a system's agent to report its inventory again
// @Description Not implemented yet: answers 501. Agents report on startup and periodically, and no job kind carries this request to them.
// @Tags [Protection] System Management
// @Produce json
// @Param systemId path string true "System ID"
// @Success 202 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /systems/{systemId}/inventory/refresh [post]
func RestPostSystemInventoryRefresh(c echo.Context) error {
	return notImplemented(c)
}

// RestPutSystemTargets godoc
// @ID PutSystemTargets
// @Summary Set what a system's backups capture
// @Description Not implemented yet: answers 501. Targets are chosen per backup request, from the inventory or by name.
// @Tags [Protection] System Management
// @Accept json
// @Produce json
// @Param systemId path string true "System ID"
// @Param backupTargetsReq body model.BackupTargetsReq true "Backup targets"
// @Success 200 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /systems/{systemId}/targets [put]
func RestPutSystemTargets(c echo.Context) error {
	return notImplemented(c)
}

func protectedSystem(s agent.System) model.ProtectedSystem {
	a := s.Agent
	return model.ProtectedSystem{
		SystemID:     s.SystemID,
		Hostname:     a.Hostname,
		AgentID:      a.AgentID,
		AgentVersion: a.AgentVersion,
		AgentStatus:  a.State(),
		RegisteredAt: a.RegisteredAt,
		LastSeenAt:   a.LastSeenAt,
	}
}

func (h *mgmtHandler) inventorySummary(s inventory.Stored) model.InventorySummary {
	return model.InventorySummary{
		CollectedAt: s.CollectedAt,
		ReportedAt:  s.ReportedAt,
		Complete:    s.Complete,
		Stale:       s.Age() >= h.inventory.StaleAfter(),
	}
}
