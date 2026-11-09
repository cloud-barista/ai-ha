package rest

import (
	"errors"
	"net/http"

	"github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/cloud-barista/ai-ha/pkg/framework/job"
	"github.com/cloud-barista/ai-ha/pkg/framework/restore"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// bootVerificationPending is what a restored disk's boot confirmation reads
// until the framework hears from the restored system.
const bootVerificationPending = "pending"

// RestPostRestore godoc
// @ID PostRestore
// @Summary Request a restore
// @Description Register a restore job for a backup set and return at once: follow it at the address in the Location header. The set must be one a backup job reported as completed; that job says which system it captured and which repository holds it. Target instances are still created by hand, so the response carries, this once, what the instance's init script needs, including its claim token. "Completed" is what the agent reported: whether the files are still in the repository is found out by the rescue agent. Requests are not deduplicated yet: repeating one registers another job.
// @Tags [Protection] Restore Management
// @Accept json
// @Produce json
// @Param restoreReq body model.RestoreReq true "Restore request"
// @Success 202 {object} model.RestoreAcceptedResp
// @Header 202 {string} Location "Where to follow the restore"
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /restores [post]
func (h *mgmtHandler) RestPostRestore(c echo.Context) error {
	var req model.RestoreReq
	if err := c.Bind(&req); err != nil {
		return refuseMalformed(c)
	}
	if err := c.Validate(&req); err != nil {
		return refuse(c, nil, http.StatusBadRequest, model.CodeInvalidRequest, "Backup set ID required")
	}

	created, err := h.restores.Create(c.Request().Context(), restore.NewJob{BackupSetID: req.BackupSetID})
	switch {
	case errors.Is(err, job.ErrBackupSetNotFound):
		return refuse(c, nil, http.StatusBadRequest, model.CodeBackupSetNotFound,
			"Backup set not found among completed backups")
	case errors.Is(err, job.ErrBackupSetAmbiguous):
		return refuse(c, err, http.StatusConflict, model.CodeBackupSetAmbiguous, "More than one backup reported this set")
	case errors.Is(err, job.ErrLegacyJob), errors.Is(err, restore.ErrInvalidJob):
		return refuse(c, err, http.StatusConflict, model.CodeBackupSetUnusable, "Backup set cannot be restored: take a new backup")
	case err != nil:
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Restore could not be registered")
	}

	// The claim token is returned here and logged nowhere.
	log.Info().Str("restoreId", created.RestoreID).Str("recoveryId", created.RecoveryID).
		Str("requestedBy", principalOf(c).Subject).Msg("Restore job created")
	c.Response().Header().Set(echo.HeaderLocation, "/aiha/restores/"+created.RestoreID)
	return c.JSON(http.StatusAccepted, model.RestoreAcceptedResp{
		RestoreID:      created.RestoreID,
		Status:         string(restore.StatusAwaitingAgent),
		ClaimExpiresAt: created.ClaimExpiresAt,
		Rescue: model.RescueAccess{
			RecoveryID: created.RecoveryID,
			ClaimToken: created.ClaimToken,
			Framework:  h.access,
		},
	})
}

// RestGetRestores godoc
// @ID GetRestores
// @Summary List the restore history
// @Description List restore jobs, newest first.
// @Tags [Protection] Restore Management
// @Produce json
// @Param systemId query string false "Only restores of this source system"
// @Param backupSetId query string false "Only restores of this backup set"
// @Param status query string false "Only restores in this status" Enums(awaiting_agent, claimed, restoring, awaiting_boot, failed)
// @Param from query string false "Only restores requested at or after this time (RFC 3339)"
// @Param to query string false "Only restores requested before this time (RFC 3339)"
// @Param limit query int false "How many to return, 1-200" default(50)
// @Param offset query int false "How many to skip" default(0)
// @Success 200 {object} model.RestoreListResp
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /restores [get]
func (h *mgmtHandler) RestGetRestores(c echo.Context) error {
	p, err := parseListParams(c)
	if err != nil {
		return refuseListParams(c, err)
	}
	status := restore.Status(c.QueryParam("status"))
	if status != "" && !status.Known() {
		return refuse(c, nil, http.StatusBadRequest, model.CodeInvalidRequest, "Status is not a restore status")
	}

	jobs, err := h.restores.List(c.Request().Context(), restore.ListFilter{
		Status:      status,
		SystemID:    c.QueryParam("systemId"),
		BackupSetID: c.QueryParam("backupSetId"),
		From:        p.from,
		To:          p.to,
		Limit:       p.limit,
		Offset:      p.offset,
	})
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Restores could not be listed")
	}
	items := make([]model.RestoreInfo, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, restoreInfo(j))
	}
	return c.JSON(http.StatusOK, model.RestoreListResp{Items: items})
}

// RestGetRestore godoc
// @ID GetRestore
// @Summary Get a restore
// @Description Return a restore's status, the step its rescue agent is on and how much of the backup files it has read. A status of awaiting_boot means the disk was restored; the restored system's boot has not been confirmed.
// @Tags [Protection] Restore Management
// @Produce json
// @Param restoreId path string true "Restore ID"
// @Success 200 {object} model.RestoreInfo
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /restores/{restoreId} [get]
func (h *mgmtHandler) RestGetRestore(c echo.Context) error {
	id := c.Param("restoreId")

	j, err := h.restores.Get(c.Request().Context(), id)
	if errors.Is(err, restore.ErrNotFound) {
		return refuse(c, nil, http.StatusNotFound, model.CodeNotFound, "Restore not found")
	}
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Restore could not be read")
	}
	// The service also finds a job by its recovery ID, for an operator reading
	// a rescue log. That ID is the rescue instance's and is exposed to whatever
	// runs there, so this API must not answer to it.
	if j.RestoreID != id {
		return refuse(c, nil, http.StatusNotFound, model.CodeNotFound, "Restore not found")
	}
	return c.JSON(http.StatusOK, restoreInfo(j))
}

// RestPostRestoreCancel godoc
// @ID PostRestoreCancel
// @Summary Ask a restore to stop
// @Description Not implemented yet: answers 501. A rescue agent has no path to receive a cancel, and past formatting a stopped restore does not return the disk to its earlier state.
// @Tags [Protection] Restore Management
// @Produce json
// @Param restoreId path string true "Restore ID"
// @Success 202 {object} model.RestoreInfo
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /restores/{restoreId}/cancel [post]
func RestPostRestoreCancel(c echo.Context) error {
	return notImplemented(c)
}

// restoreInfo describes a restore job. It carries no recovery ID or claim
// details: those belong to the rescue instance, not to whoever follows the job.
func restoreInfo(j restore.Job) model.RestoreInfo {
	info := model.RestoreInfo{
		RestoreID:      j.RestoreID,
		Status:         string(j.Status),
		RepositoryID:   j.RepositoryID,
		RestoreOutcome: string(j.Outcome),
		LastProgressAt: j.LastProgressAt,
		Message:        j.Message,
		CreatedAt:      j.CreatedAt,
		ClaimExpiresAt: j.ClaimExpiresAt,
		ClaimedAt:      j.ClaimedAt,
		FinishedAt:     j.FinishedAt,
	}
	if plan, err := model.DecodeRestorePlan(j.Plan); err != nil {
		// History: one row in an unexpected shape must not cost the whole list.
		log.Warn().Err(err).Str("restoreId", j.RestoreID).Msg("Restore job has an unreadable plan")
	} else {
		info.SourceSystemID = plan.SourceSystemID
		info.SourceBackupSetID = plan.SourceBackupSetID
	}
	if j.Status == restore.StatusAwaitingBoot {
		info.BootVerification = bootVerificationPending
	}
	if p := j.Progress; p.Seq > 0 || p.Phase != "" {
		info.Progress = &model.RestoreProgressInfo{
			Phase:          p.Phase,
			Step:           p.Step,
			TotalSteps:     p.TotalSteps,
			Item:           p.Item,
			ProcessedBytes: p.ProcessedBytes,
			TotalBytes:     p.TotalBytes,
			TotalUnknown:   p.TotalBytes == 0,
		}
	}
	return info
}
