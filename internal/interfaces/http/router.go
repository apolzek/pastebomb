// Package http is the delivery layer: it turns the application use cases into
// an HTTP API. Nothing below this package knows that HTTP exists.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/infrastructure/config"
	"github.com/apolzek/pastebomb/internal/interfaces/http/handler"
	"github.com/apolzek/pastebomb/internal/interfaces/http/middleware"
)

// Dependencies is everything the router needs, supplied by the composition
// root. Passing it as one struct keeps the wiring readable as the API grows.
type Dependencies struct {
	Config        config.Config
	AuthService   *application.AuthService
	UserService   *application.UserService
	PasteService  *application.PasteService
	TokenVerifier application.TokenVerifier
	CheckDatabase handler.HealthChecker
}

// Limiters bundles the rate limiters so the caller can stop their janitor
// goroutines on shutdown.
type Limiters struct {
	Login     *middleware.RateLimiter
	Anonymous *middleware.RateLimiter
}

// Stop releases both limiters.
func (l Limiters) Stop() {
	l.Login.Stop()
	l.Anonymous.Stop()
}

// Rate limit budgets:
//   - login is deliberately harsh, since a brute force needs many attempts but
//     a real user needs only a handful.
//   - anonymous writes are looser: they carry no credential to guess, they
//     just must not become a spam firehose.
var (
	loginRPS, loginBurst = 1.0 / 5.0, 5
	anonRPS, anonBurst   = 1.0, 10
)

// NewRouter builds the Gin engine with the full middleware stack.
func NewRouter(deps Dependencies) (*gin.Engine, Limiters) {
	engine := gin.New()
	// Client IPs drive the rate limiter, so forwarded-for headers are only
	// honoured for proxies that were explicitly configured as trusted.
	_ = engine.SetTrustedProxies(deps.Config.HTTP.TrustedProxies)

	engine.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.RequestLogger(),
		middleware.SecureHeaders(),
		middleware.BodyLimit(deps.Config.HTTP.MaxBodyBytes),
		newCORS(deps.Config.CORS),
	)

	limiters := Limiters{
		Login:     middleware.NewRateLimiter(loginRPS, loginBurst),
		Anonymous: middleware.NewRateLimiter(anonRPS, anonBurst),
	}

	authHandler := handler.NewAuthHandler(deps.AuthService)
	userHandler := handler.NewUserHandler(deps.UserService)
	pasteHandler := handler.NewPasteHandler(deps.PasteService)
	healthHandler := handler.NewHealthHandler(deps.CheckDatabase)

	authenticated := middleware.Authenticate(deps.TokenVerifier)
	adminOnly := middleware.RequireAdmin()

	engine.GET("/health", healthHandler.Live)
	engine.GET("/ready", healthHandler.Ready)
	engine.POST("/login", limiters.Login.Handler(), authHandler.Login)

	users := engine.Group("/u")
	{
		// Public.
		users.POST("", limiters.Anonymous.Handler(), userHandler.Register)
		users.POST("/", limiters.Anonymous.Handler(), userHandler.Register)
		users.GET("/anonymous/posts", pasteHandler.ListPublic)
		users.POST("/anonymous/post", limiters.Anonymous.Handler(), pasteHandler.CreateAnonymous)
		users.GET("/anonymous/:slug", pasteHandler.Read)
		users.GET("/anonymous/:slug/raw", pasteHandler.ReadRaw)

		// Authenticated, acting on the caller's own account.
		self := users.Group("/me", authenticated)
		{
			self.GET("", userHandler.Me)
			self.PATCH("", userHandler.UpdateMe)
			self.DELETE("", userHandler.DeactivateMe)
			self.GET("/posts", pasteHandler.ListMine)
			self.GET("/posts/content", pasteHandler.ListMineWithContent)
			self.POST("/post", pasteHandler.CreateOwned)
		}

		// Administrative.
		users.GET("/all", authenticated, adminOnly, userHandler.List)
		users.GET("", authenticated, adminOnly, userHandler.ListActive)
		users.GET("/paginate", authenticated, adminOnly, userHandler.ListActive)

		// Public profile lookup. Declared last so the static routes above win.
		users.GET("/:handle", userHandler.PublicProfile)
		users.DELETE("/:handle", authenticated, userHandler.DeactivateByID)
	}

	return engine, limiters
}
