package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_AllowsUpToBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(0.001, 3) // tiny refill, burst of 3

	r := gin.New()
	r.GET("/", rl.Handler(), func(c *gin.Context) { c.Status(200) })

	doReq := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:5678"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	// First `burst` requests succeed.
	assert.Equal(t, 200, doReq())
	assert.Equal(t, 200, doReq())
	assert.Equal(t, 200, doReq())

	// Bucket drained → next call must 429.
	assert.Equal(t, 429, doReq())
}

func TestRateLimiter_PerIPIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(0.001, 1)

	r := gin.New()
	r.GET("/", rl.Handler(), func(c *gin.Context) { c.Status(200) })

	hit := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip + ":1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, 200, hit("10.0.0.1"))
	assert.Equal(t, 429, hit("10.0.0.1"), "same IP should be throttled")
	assert.Equal(t, 200, hit("10.0.0.2"), "different IP has its own bucket")
}
