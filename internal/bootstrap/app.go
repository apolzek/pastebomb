// Package bootstrap is the composition root. It is the only place that knows
// how every layer fits together, which is what keeps the layers themselves
// free of import cycles and of hidden global state.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/infrastructure/audit"
	"github.com/apolzek/pastebomb/internal/infrastructure/config"
	"github.com/apolzek/pastebomb/internal/infrastructure/persistence"
	"github.com/apolzek/pastebomb/internal/infrastructure/security"
	httpiface "github.com/apolzek/pastebomb/internal/interfaces/http"
)

// Run wires the application and serves until ctx is cancelled.
func Run(ctx context.Context) error {
	// Loading .env is best effort: containers get their environment injected.
	if err := godotenv.Load(); err != nil {
		slog.Debug("no .env file loaded, relying on the process environment")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	configureLogging()

	db, err := persistence.Open(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer func() {
		if err := persistence.Close(db); err != nil {
			slog.Error("closing database failed", "err", err)
		}
	}()

	auditRecorder := audit.NewFileRecorder(cfg.Security.AuditLogDir)
	defer func() {
		if err := auditRecorder.Close(); err != nil {
			slog.Error("closing audit log failed", "err", err)
		}
	}()

	clock := shared.SystemClock{}
	hasher := security.NewBcryptHasher(cfg.Security.BcryptCost)
	tokens := security.NewJWTProvider(cfg.Security.JWTSecret, cfg.Security.JWTIssuer, cfg.Security.TokenTTL)
	slugs := security.NewRandomSlugGenerator(paste.SlugLength)

	userRepository := persistence.NewUserRepository(db)
	pasteRepository := persistence.NewPasteRepository(db)

	pasteService := application.NewPasteService(
		pasteRepository,
		paste.NewFactory(slugs, hasher, clock),
		paste.NewAccessGuard(hasher, clock),
		clock,
		auditRecorder,
	)

	engine, limiters := httpiface.NewRouter(httpiface.Dependencies{
		Config:        cfg,
		AuthService:   application.NewAuthService(userRepository, hasher, tokens, auditRecorder),
		UserService:   application.NewUserService(userRepository, hasher, clock, auditRecorder),
		PasteService:  pasteService,
		TokenVerifier: tokens,
		CheckDatabase: func(probeCtx context.Context) error { return persistence.Ping(probeCtx, db) },
	})

	purger := NewPurgeWorker(pasteService, cfg.Purge)
	purger.Start(ctx)
	defer purger.Wait()

	return httpiface.NewServer(engine, limiters, cfg.HTTP).Run(ctx)
}

// configureLogging installs a JSON slog handler and puts Gin in release mode
// unless GIN_MODE says otherwise, so the process never prints debug routing
// tables in production.
func configureLogging() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
}
