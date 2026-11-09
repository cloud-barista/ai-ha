package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/cloud-barista/ai-ha/pkg/framework/token"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// ctxPrincipal is the context key carrying the authenticated management caller.
const ctxPrincipal = "aiha.principal"

// adminSubject is who a caller holding the shared admin token is recorded as.
const adminSubject = "admin"

// ErrAdminTokenRejected reports a management credential that is not valid.
var ErrAdminTokenRejected = errors.New("admin token rejected")

// Principal is the authenticated caller of a management route. Handlers read
// who is calling from here and never from the credential, so that replacing
// the way callers authenticate does not reach them.
type Principal struct {
	// Subject names the caller in what it creates, e.g. a job's requester.
	Subject string
}

// AdminAuthenticator identifies the holder of a management credential. It
// returns ErrAdminTokenRejected for one that is not valid; any other error
// means it could not tell, which the caller may retry.
type AdminAuthenticator func(ctx context.Context, token string) (Principal, error)

// StaticAdminToken authenticates the holder of the single token whose digest
// is hash. It is the interim credential for the management API until integrated
// user authentication (SFR-API-03) takes the place of this function.
func StaticAdminToken(hash []byte) AdminAuthenticator {
	return func(_ context.Context, presented string) (Principal, error) {
		if !token.Matches(hash, presented) {
			return Principal{}, ErrAdminTokenRejected
		}
		return Principal{Subject: adminSubject}, nil
	}
}

// adminAuth authenticates a management caller. It is separate from agentAuth on
// purpose: an agent token must never open this API, nor the reverse.
func adminAuth(authenticate AdminAuthenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			presented, ok := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if !ok {
				return c.JSON(http.StatusUnauthorized,
					model.SimpleMsg{Message: "Admin token required", Code: model.CodeUnauthorized})
			}

			principal, err := authenticate(c.Request().Context(), presented)
			switch {
			case err == nil:
				c.Set(ctxPrincipal, principal)
				return next(c)
			case errors.Is(err, ErrAdminTokenRejected):
				return c.JSON(http.StatusUnauthorized,
					model.SimpleMsg{Message: "Admin token rejected", Code: model.CodeUnauthorized})
			default:
				log.Error().Err(err).Msg("Could not authenticate a management caller")
				c.Response().Header().Set(echo.HeaderRetryAfter, retryAfterSeconds)
				return c.JSON(http.StatusServiceUnavailable,
					model.SimpleMsg{Message: "Service temporarily unavailable, retry shortly", Code: model.CodeUnavailable})
			}
		}
	}
}

// principalOf returns the authenticated caller. It is only called from handlers
// behind adminAuth, which always sets it.
func principalOf(c echo.Context) Principal {
	p, _ := c.Get(ctxPrincipal).(Principal)
	return p
}
