package rest

import (
	// swag resolves the model names in the annotations through this file's imports.
	_ "github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// The incident archive the HA support agent records to (TM-AI-HA-04 chapter 6)
// and the integrated history (SFR-HA-07). Nothing behind it exists yet, and its
// fields are not settled, so these routes hold the paths and answer 501 with
// no body contract.

// RestGetIncidents godoc
// @ID GetIncidents
// @Summary List incidents
// @Description Not implemented yet: answers 501. Will list incidents by target, type and period.
// @Tags [Diagnosis] Incident Management
// @Produce json
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /incidents [get]
func RestGetIncidents(c echo.Context) error {
	return notImplemented(c)
}

// RestPostIncident godoc
// @ID PostIncident
// @Summary Record an incident
// @Description Not implemented yet: answers 501. Will record an incident on behalf of the HA support agent.
// @Tags [Diagnosis] Incident Management
// @Accept json
// @Produce json
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /incidents [post]
func RestPostIncident(c echo.Context) error {
	return notImplemented(c)
}

// RestGetIncident godoc
// @ID GetIncident
// @Summary Get an incident
// @Description Not implemented yet: answers 501. Will return an incident's events, diagnoses, alarms and actions.
// @Tags [Diagnosis] Incident Management
// @Produce json
// @Param incidentId path string true "Incident ID"
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /incidents/{incidentId} [get]
func RestGetIncident(c echo.Context) error {
	return notImplemented(c)
}

// RestPostIncidentEvent godoc
// @ID PostIncidentEvent
// @Summary Record an incident event
// @Description Not implemented yet: answers 501.
// @Tags [Diagnosis] Incident Management
// @Accept json
// @Produce json
// @Param incidentId path string true "Incident ID"
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /incidents/{incidentId}/events [post]
func RestPostIncidentEvent(c echo.Context) error {
	return notImplemented(c)
}

// RestPostIncidentDiagnosis godoc
// @ID PostIncidentDiagnosis
// @Summary Record an incident diagnosis
// @Description Not implemented yet: answers 501. Diagnoses will accumulate rather than replace one another, so it stays visible which one an action was approved on.
// @Tags [Diagnosis] Incident Management
// @Accept json
// @Produce json
// @Param incidentId path string true "Incident ID"
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /incidents/{incidentId}/diagnoses [post]
func RestPostIncidentDiagnosis(c echo.Context) error {
	return notImplemented(c)
}

// RestPostIncidentAlarm godoc
// @ID PostIncidentAlarm
// @Summary Record an incident alarm
// @Description Not implemented yet: answers 501.
// @Tags [Diagnosis] Incident Management
// @Accept json
// @Produce json
// @Param incidentId path string true "Incident ID"
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /incidents/{incidentId}/alarms [post]
func RestPostIncidentAlarm(c echo.Context) error {
	return notImplemented(c)
}

// RestPostIncidentAction godoc
// @ID PostIncidentAction
// @Summary Record an incident action
// @Description Not implemented yet: answers 501. Actions will refer to the switch or restore job they caused rather than copy its history.
// @Tags [Diagnosis] Incident Management
// @Accept json
// @Produce json
// @Param incidentId path string true "Incident ID"
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /incidents/{incidentId}/actions [post]
func RestPostIncidentAction(c echo.Context) error {
	return notImplemented(c)
}

// RestGetHistory godoc
// @ID GetHistory
// @Summary List the integrated history
// @Description Not implemented yet: answers 501. Will list redundancy, switch, backup and restore work and incidents in one timeline (SFR-HA-07).
// @Tags [Diagnosis] Incident Management
// @Produce json
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /history [get]
func RestGetHistory(c echo.Context) error {
	return notImplemented(c)
}
