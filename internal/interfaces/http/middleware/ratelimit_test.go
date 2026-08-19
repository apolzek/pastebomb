package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/interfaces/http/middleware"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// newLimitedEngine wires the limiter in front of a trivial handler.
func newLimitedEngine(limiter *middleware.RateLimiter) *gin.Engine {
	engine := gin.New()
	engine.GET("/", limiter.Handler(), func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	return engine
}

func call(engine *gin.Engine, clientIP string) int {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = clientIP + ":1234"

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestTheBurstIsAllowedAndThenRequestsAreRejected(t *testing.T) {
	t.Parallel()

	limiter := middleware.NewRateLimiter(0.001, 3)
	t.Cleanup(limiter.Stop)
	engine := newLimitedEngine(limiter)

	for i := range 3 {
		assert.Equal(t, http.StatusOK, call(engine, "10.0.0.1"), "request %d should be inside the burst", i+1)
	}
	assert.Equal(t, http.StatusTooManyRequests, call(engine, "10.0.0.1"))
}

func TestRejectionsCarryARetryAfterHeader(t *testing.T) {
	t.Parallel()

	limiter := middleware.NewRateLimiter(0.001, 1)
	t.Cleanup(limiter.Stop)
	engine := newLimitedEngine(limiter)

	require.Equal(t, http.StatusOK, call(engine, "10.0.0.2"))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "10.0.0.2:1234"
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "1", recorder.Header().Get("Retry-After"))
	assert.Contains(t, recorder.Body.String(), "rate_limited")
}

// Buckets are per client, so one noisy address must not lock everyone out.
func TestClientsAreLimitedIndependently(t *testing.T) {
	t.Parallel()

	limiter := middleware.NewRateLimiter(0.001, 1)
	t.Cleanup(limiter.Stop)
	engine := newLimitedEngine(limiter)

	require.Equal(t, http.StatusOK, call(engine, "10.0.0.3"))
	require.Equal(t, http.StatusTooManyRequests, call(engine, "10.0.0.3"))

	assert.Equal(t, http.StatusOK, call(engine, "10.0.0.4"))
}

// The previous limiter kept one entry per address forever, so a stream of
// spoofed sources grew the table without bound until the process died.
func TestIdleClientsAreEvictedSoTheTableStaysBounded(t *testing.T) {
	t.Parallel()

	limiter := middleware.NewRateLimiterWithOptions(middleware.Options{
		RequestsPerSecond: 100,
		Burst:             100,
		IdleTTL:           20 * time.Millisecond,
		SweepInterval:     10 * time.Millisecond,
	})
	t.Cleanup(limiter.Stop)
	engine := newLimitedEngine(limiter)

	for _, ip := range []string{"10.1.0.1", "10.1.0.2", "10.1.0.3"} {
		require.Equal(t, http.StatusOK, call(engine, ip))
	}
	require.Equal(t, 3, limiter.Size())

	assert.Eventually(t, func() bool { return limiter.Size() == 0 }, time.Second, 10*time.Millisecond,
		"idle client buckets should have been evicted")
}

func TestTheTableIsCappedEvenUnderSustainedTraffic(t *testing.T) {
	t.Parallel()

	limiter := middleware.NewRateLimiterWithOptions(middleware.Options{
		RequestsPerSecond: 100,
		Burst:             100,
		MaxTracked:        8,
	})
	t.Cleanup(limiter.Stop)
	engine := newLimitedEngine(limiter)

	for i := range 100 {
		call(engine, "10.2."+itoa(i/256)+"."+itoa(i%256))
	}
	assert.LessOrEqual(t, limiter.Size(), 8)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
