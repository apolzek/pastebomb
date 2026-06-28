package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter is a per-client-IP token bucket. The map grows unbounded with
// unique IPs — fine for a small internal pastebin; swap for an LRU if this
// gets exposed to the broader internet.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

// NewRateLimiter caps each IP to `rps` requests per second with a burst of
// `burst` (use burst=rps for strict, or higher to absorb spikes).
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

func (rl *RateLimiter) limiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if l, ok := rl.visitors[ip]; ok {
		return l
	}
	l := rate.NewLimiter(rl.rps, rl.burst)
	rl.visitors[ip] = l
	return l
}

// Handler returns a Gin middleware that 429s when the bucket is exhausted.
// Set Retry-After to a conservative second so well-behaved clients back off.
func (rl *RateLimiter) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !rl.limiter(ctx.ClientIP()).Allow() {
			ctx.Header("Retry-After", "1")
			ctx.AbortWithStatusJSON(429, gin.H{"message": "too many requests"})
			return
		}
		ctx.Next()
	}
}

// Tunings used by the routes:
//   - LoginLimiter: harsh enough that brute force is impractical, loose
//     enough that a typo plus a retry doesn't lock a real user out.
//   - AnonLimiter:  protects the public paste creation endpoint from spam.
var (
	LoginLimiter = NewRateLimiter(1.0/5.0, 5) // 1 req / 5s, burst 5
	AnonLimiter  = NewRateLimiter(1.0, 10)    // 1 req / s,  burst 10
)

