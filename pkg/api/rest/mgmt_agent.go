package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/cloud-barista/ai-ha/pkg/framework/agent"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// A registration token is good for one enrolment within this long unless the
// request says otherwise.
const (
	defaultTokenTTLSeconds = 30 * 60
	defaultTokenMaxUses    = 1
)

// RestGetAgents godoc
// @ID GetAgents
// @Summary List enrolled agents
// @Description List enrolled agents, most recently enrolled first, with when each was last heard from. An agent is offline when it has not been heard from for two minutes.
// @Tags [Protection] Agent Management
// @Produce json
// @Success 200 {object} model.AgentListResp
// @Failure 401 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /agents [get]
func (h *mgmtHandler) RestGetAgents(c echo.Context) error {
	agents, err := h.agents.ListAgents(c.Request().Context())
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Agents could not be listed")
	}
	items := make([]model.AgentInfo, 0, len(agents))
	for _, a := range agents {
		items = append(items, model.AgentInfo{
			AgentID:             a.AgentID,
			SystemID:            a.SystemID,
			Hostname:            a.Hostname,
			AgentVersion:        a.AgentVersion,
			Mode:                string(a.Mode),
			Status:              a.State(),
			RegistrationTokenID: a.RegistrationTokenID,
			RegisteredAt:        a.RegisteredAt,
			LastSeenAt:          a.LastSeenAt,
			RevokedAt:           a.RevokedAt,
		})
	}
	return c.JSON(http.StatusOK, model.AgentListResp{Items: items})
}

// RestPostAgentRevoke godoc
// @ID PostAgentRevoke
// @Summary Revoke an agent's token
// @Description Revoke an agent's token, which stops working on its next request. Repeating this is safe. The agent keeps its local credentials, so enrolling it again takes "aiha-agent reset" on the host and a new registration token.
// @Tags [Protection] Agent Management
// @Produce json
// @Param agentId path string true "Agent ID"
// @Success 200 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /agents/{agentId}/revoke [post]
func (h *mgmtHandler) RestPostAgentRevoke(c echo.Context) error {
	agentID := c.Param("agentId")

	err := h.agents.RevokeAgent(c.Request().Context(), agentID)
	if errors.Is(err, agent.ErrAgentNotFound) {
		return refuse(c, nil, http.StatusNotFound, model.CodeNotFound, "Agent not found")
	}
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Agent could not be revoked")
	}
	log.Info().Str("agentId", agentID).Str("requestedBy", principalOf(c).Subject).Msg("Agent token revoked")
	return c.JSON(http.StatusOK, model.SimpleMsg{Message: "Agent revoked"})
}

// RestGetRegistrationTokens godoc
// @ID GetRegistrationTokens
// @Summary List registration tokens
// @Description List issued registration tokens and what is left of them. The tokens themselves are not stored and cannot be listed.
// @Tags [Protection] Agent Management
// @Produce json
// @Success 200 {object} model.RegistrationTokenListResp
// @Failure 401 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /registration-tokens [get]
func (h *mgmtHandler) RestGetRegistrationTokens(c echo.Context) error {
	tokens, now, err := h.agents.ListRegistrationTokens(c.Request().Context())
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Registration tokens could not be listed")
	}
	items := make([]model.RegistrationTokenInfo, 0, len(tokens))
	for _, t := range tokens {
		items = append(items, model.RegistrationTokenInfo{
			TokenID:   t.TokenID,
			Status:    t.State(now),
			Uses:      t.Uses,
			MaxUses:   t.MaxUses,
			ExpiresAt: t.ExpiresAt,
			Note:      t.Note,
			CreatedAt: t.CreatedAt,
			RevokedAt: t.RevokedAt,
		})
	}
	return c.JSON(http.StatusOK, model.RegistrationTokenListResp{Items: items})
}

// RestPostRegistrationToken godoc
// @ID PostRegistrationToken
// @Summary Issue a registration token
// @Description Issue a token that lets an agent enrol, with the settings the agent needs to reach this framework. It authorizes enrolment and nothing else. The token is returned only here and cannot be recovered: if an agent never receives its credentials, revoke the incomplete agent and issue a new one. An empty request issues a token good for one enrolment within 30 minutes.
// @Tags [Protection] Agent Management
// @Accept json
// @Produce json
// @Param registrationTokenReq body model.RegistrationTokenReq false "Token limits"
// @Success 201 {object} model.RegistrationTokenIssuedResp
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /registration-tokens [post]
func (h *mgmtHandler) RestPostRegistrationToken(c echo.Context) error {
	var req model.RegistrationTokenReq
	if err := c.Bind(&req); err != nil {
		return refuseMalformed(c)
	}
	if err := c.Validate(&req); err != nil {
		return refuse(c, nil, http.StatusBadRequest, model.CodeInvalidRequest,
			"TTL must be at most a week, uses at most 10000 and the note at most 512 characters")
	}
	if req.TTLSeconds == 0 {
		req.TTLSeconds = defaultTokenTTLSeconds
	}
	if req.MaxUses == 0 {
		req.MaxUses = defaultTokenMaxUses
	}

	issued, err := h.agents.IssueRegistrationToken(c.Request().Context(),
		time.Duration(req.TTLSeconds)*time.Second, req.MaxUses, req.Note)
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Registration token could not be issued")
	}

	// The token itself is never logged, only the record of issuing one.
	log.Info().Str("tokenId", issued.TokenID).Time("expiresAt", issued.ExpiresAt).
		Int("maxUses", issued.MaxUses).Str("requestedBy", principalOf(c).Subject).
		Msg("Registration token issued")
	return c.JSON(http.StatusCreated, model.RegistrationTokenIssuedResp{
		TokenID:   issued.TokenID,
		Token:     issued.Token,
		ExpiresAt: issued.ExpiresAt,
		MaxUses:   issued.MaxUses,
		Framework: h.access,
	})
}

// RestPostRegistrationTokenRevoke godoc
// @ID PostRegistrationTokenRevoke
// @Summary Revoke a registration token
// @Description Revoke a registration token so that it can enrol nothing further. Agents it already enrolled are unaffected. Repeating this is safe.
// @Tags [Protection] Agent Management
// @Produce json
// @Param tokenId path string true "Registration token ID"
// @Success 200 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /registration-tokens/{tokenId}/revoke [post]
func (h *mgmtHandler) RestPostRegistrationTokenRevoke(c echo.Context) error {
	tokenID := c.Param("tokenId")

	err := h.agents.RevokeRegistrationToken(c.Request().Context(), tokenID)
	if errors.Is(err, agent.ErrTokenNotFound) {
		return refuse(c, nil, http.StatusNotFound, model.CodeNotFound, "Registration token not found")
	}
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Registration token could not be revoked")
	}
	log.Info().Str("tokenId", tokenID).Str("requestedBy", principalOf(c).Subject).Msg("Registration token revoked")
	return c.JSON(http.StatusOK, model.SimpleMsg{Message: "Registration token revoked"})
}
