package http

import (
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/infrastructure/config"
)

// newCORS builds the browser access policy from configuration.
//
// Credentials are never allowed: the API authenticates with a bearer token,
// not a cookie, so enabling them would only widen the blast radius of a
// permissive origin list.
func newCORS(cfg config.CORS) gin.HandlerFunc {
	policy := cors.DefaultConfig()
	policy.AllowCredentials = false
	policy.AllowHeaders = append(policy.AllowHeaders, "Authorization", secretHeaderName)
	policy.ExposeHeaders = append(policy.ExposeHeaders, "X-Request-ID", "Retry-After")

	if cfg.AllowAllOrigins {
		slog.Warn("CORS is open to every origin, set CORS_ALLOWED_ORIGINS for a production deployment")
		policy.AllowAllOrigins = true
	} else {
		policy.AllowOrigins = cfg.AllowedOrigins
	}

	return cors.New(policy)
}

// secretHeaderName mirrors the header the paste handler reads. It is repeated
// here rather than imported so the CORS policy does not depend on a handler.
const secretHeaderName = "X-Post-Secret"
