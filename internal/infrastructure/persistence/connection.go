// Package persistence contains the GORM adapters that implement the domain
// repository ports. It is the only package aware of table and column names.
package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/apolzek/pastebomb/internal/infrastructure/config"
)

// Open connects to the database, retrying with a fixed backoff so a container
// that starts before its database still comes up cleanly.
//
// Connection errors are logged without the DSN: it carries the password.
func Open(ctx context.Context, cfg config.Database) (*gorm.DB, error) {
	var lastErr error

	for attempt := 1; attempt <= cfg.ConnectAttempts; attempt++ {
		db, err := connect(cfg)
		if err == nil {
			slog.Info("database connected", "driver", cfg.Driver, "host", cfg.Host, "name", cfg.Name)
			return db, nil
		}

		lastErr = err
		slog.Warn("database connection failed",
			"attempt", attempt,
			"attempts", cfg.ConnectAttempts,
			"driver", cfg.Driver,
			"host", cfg.Host,
			"err", err,
		)

		if attempt == cfg.ConnectAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cfg.ConnectBackoff):
		}
	}

	return nil, fmt.Errorf("connect to %s database after %d attempts: %w", cfg.Driver, cfg.ConnectAttempts, lastErr)
}

func connect(cfg config.Database) (*gorm.DB, error) {
	dialector, err := dialector(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		// TranslateError turns driver-specific duplicate-key and
		// no-rows failures into the portable gorm sentinels the
		// repositories map to domain errors.
		TranslateError: true,
		Logger:         gormlogger.Default.LogMode(gormlogger.Warn),
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, err
	}

	pool, err := db.DB()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(cfg.MaxOpenConns)
	pool.SetMaxIdleConns(cfg.MaxIdleConns)
	pool.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, pool.Ping()
}

func dialector(cfg config.Database) (gorm.Dialector, error) {
	switch cfg.Driver {
	case config.DriverMySQL:
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
			cfg.User, url.QueryEscape(cfg.Password), cfg.Host, cfg.Port, cfg.Name,
		)
		return mysql.Open(dsn), nil

	case config.DriverPostgres:
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
		)
		return postgres.Open(dsn), nil

	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
}

// Close releases the pool. Called on shutdown.
func Close(db *gorm.DB) error {
	pool, err := db.DB()
	if err != nil {
		return err
	}
	return pool.Close()
}

// Ping verifies the connection is still usable, backing the readiness probe.
func Ping(ctx context.Context, db *gorm.DB) error {
	pool, err := db.DB()
	if err != nil {
		return err
	}
	return pool.PingContext(ctx)
}
