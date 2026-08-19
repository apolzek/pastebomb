package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker reports whether a downstream dependency is usable.
type HealthChecker func(ctx context.Context) error

// HealthHandler exposes the liveness and readiness probes.
type HealthHandler struct{ checkDatabase HealthChecker }

func NewHealthHandler(checkDatabase HealthChecker) *HealthHandler {
	return &HealthHandler{checkDatabase: checkDatabase}
}

// Live handles GET /health. It answers as long as the process is running, so
// an orchestrator does not restart a healthy pod because its database blipped.
func (h *HealthHandler) Live(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// Ready handles GET /ready and does check the database, so a pod that cannot
// serve traffic is pulled out of the load balancer.
func (h *HealthHandler) Ready(ctx *gin.Context) {
	probeCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.checkDatabase(probeCtx); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "dependency": "database"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
}
