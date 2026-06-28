package cors_config

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const envAllowedOrigins = "CORS_ALLOWED_ORIGINS"

// allowedOrigins reads a comma-separated list from CORS_ALLOWED_ORIGINS and
// falls back to localhost:8080 if unset. Wildcard "*" is honored verbatim.
func allowedOrigins() []string {
	raw := os.Getenv(envAllowedOrigins)
	if raw == "" {
		return []string{"http://localhost:8080"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func CorsConfigContrib() gin.HandlerFunc {
	config := cors.DefaultConfig()
	origins := allowedOrigins()
	if len(origins) == 1 && origins[0] == "*" {
		config.AllowAllOrigins = true
	} else {
		config.AllowOrigins = origins
	}
	config.AllowHeaders = append(config.AllowHeaders, "Authorization", "X-Post-Secret", "owner")
	return cors.New(config)
}
