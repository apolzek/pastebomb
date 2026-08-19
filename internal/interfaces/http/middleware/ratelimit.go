package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/interfaces/http/reqctx"
	"golang.org/x/time/rate"

	"github.com/apolzek/pastebomb/internal/interfaces/http/dto"
)

// visitor is one client's bucket plus the last time it was used, which is what
// lets the janitor decide when the entry can go.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Defaults for the bookkeeping knobs.
const (
	defaultSweepInterval = time.Minute
	defaultMinIdleTTL    = 10 * time.Minute
	// defaultMaxTracked caps memory even if traffic outpaces the janitor.
	defaultMaxTracked = 50_000
)

// Options tunes a RateLimiter. Zero-valued fields fall back to the defaults.
type Options struct {
	// RequestsPerSecond is the sustained rate allowed per client.
	RequestsPerSecond float64
	// Burst is how many requests a client may make back to back.
	Burst int
	// IdleTTL is how long an unused bucket is kept. Forgetting a bucket that
	// has had time to refill is equivalent to keeping it, so this never
	// loosens the limit.
	IdleTTL time.Duration
	// SweepInterval is how often idle buckets are evicted.
	SweepInterval time.Duration
	// MaxTracked bounds the number of clients held in memory.
	MaxTracked int
}

func (o Options) withDefaults() Options {
	if o.SweepInterval <= 0 {
		o.SweepInterval = defaultSweepInterval
	}
	if o.MaxTracked <= 0 {
		o.MaxTracked = defaultMaxTracked
	}
	if o.IdleTTL <= 0 {
		o.IdleTTL = defaultMinIdleTTL
		if o.RequestsPerSecond > 0 {
			refill := time.Duration(float64(o.Burst)/o.RequestsPerSecond) * time.Second
			if refill > o.IdleTTL {
				o.IdleTTL = refill
			}
		}
	}
	return o
}

// RateLimiter is a per-client-IP token bucket with bounded memory.
//
// The previous implementation kept one entry per address forever, so a stream
// of spoofed source addresses grew the map until the process ran out of
// memory. Entries are now evicted once they have been idle long enough that
// their bucket would have refilled anyway.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	options  Options
	stop     chan struct{}
	stopOnce sync.Once
}

// NewRateLimiter allows rps requests per second per client, tolerating bursts
// of burst requests.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return NewRateLimiterWithOptions(Options{RequestsPerSecond: rps, Burst: burst})
}

// NewRateLimiterWithOptions is the full constructor.
func NewRateLimiterWithOptions(options Options) *RateLimiter {
	limiter := &RateLimiter{
		visitors: make(map[string]*visitor),
		options:  options.withDefaults(),
		stop:     make(chan struct{}),
	}
	go limiter.janitor()
	return limiter
}

// Handler returns the Gin middleware.
func (rl *RateLimiter) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !rl.allow(ctx.ClientIP()) {
			ctx.Header("Retry-After", "1")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Message:   "too many requests, slow down",
				Code:      "rate_limited",
				RequestID: reqctx.RequestID(ctx),
			})
			return
		}
		ctx.Next()
	}
}

// Stop ends the janitor goroutine. Safe to call more than once.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() { close(rl.stop) })
}

// Size reports how many clients are currently tracked.
func (rl *RateLimiter) Size() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.visitors)
}

func (rl *RateLimiter) allow(clientIP string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	known, exists := rl.visitors[clientIP]
	if !exists {
		// Once the cap is reached the table is dropped wholesale. Every client
		// then gets a fresh bucket, which is a brief relaxation of the limit
		// rather than an outage or an unbounded allocation.
		if len(rl.visitors) >= rl.options.MaxTracked {
			rl.visitors = make(map[string]*visitor, rl.options.MaxTracked/2)
		}
		known = &visitor{limiter: rate.NewLimiter(rate.Limit(rl.options.RequestsPerSecond), rl.options.Burst)}
		rl.visitors[clientIP] = known
	}
	known.lastSeen = now

	return known.limiter.AllowN(now, 1)
}

func (rl *RateLimiter) janitor() {
	ticker := time.NewTicker(rl.options.SweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stop:
			return
		case now := <-ticker.C:
			rl.evictIdle(now)
		}
	}
}

func (rl *RateLimiter) evictIdle(now time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for clientIP, known := range rl.visitors {
		if now.Sub(known.lastSeen) > rl.options.IdleTTL {
			delete(rl.visitors, clientIP)
		}
	}
}
