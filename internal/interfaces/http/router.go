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
	Read      *middleware.RateLimiter
}

// Stop releases every limiter.
func (l Limiters) Stop() {
	l.Login.Stop()
	l.Anonymous.Stop()
	l.Read.Stop()
}

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
		Login:     middleware.NewRateLimiter(deps.Config.RateLimit.LoginRPS, deps.Config.RateLimit.LoginBurst),
		Anonymous: middleware.NewRateLimiter(deps.Config.RateLimit.AnonRPS, deps.Config.RateLimit.AnonBurst),
		Read:      middleware.NewRateLimiter(deps.Config.RateLimit.ReadRPS, deps.Config.RateLimit.ReadBurst),
	}

	// Reading a secret-protected paste runs bcrypt, which costs about 250
	// times an ordinary read on a route that needs no credential. Capping the
	// public reads keeps that from becoming an amplification lever.
	readLimit := limiters.Read.Handler()

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
		users.GET("/anonymous/posts", readLimit, pasteHandler.ListPublic)
		users.POST("/anonymous/post", limiters.Anonymous.Handler(), pasteHandler.CreateAnonymous)
		users.GET("/anonymous/:slug", readLimit, pasteHandler.Read)
		users.GET("/anonymous/:slug/raw", readLimit, pasteHandler.ReadRaw)

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
		users.GET("/:handle", readLimit, userHandler.PublicProfile)
		users.DELETE("/:handle", authenticated, userHandler.DeactivateByID)
	}

	return engine, limiters
}
