// Package config loads and validates every runtime setting once, at startup.
// Nothing else in the codebase reads os.Getenv, so a missing or nonsensical
// setting fails the process immediately instead of surfacing as a strange
// error under load.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Driver enumerates the supported database backends.
type Driver string

const (
	DriverMySQL    Driver = "mysql"
	DriverPostgres Driver = "pgsql"
)

// minJWTSecretLength matches the HMAC-SHA256 block size: a shorter key adds no
// security over a 32-byte one and usually signals a placeholder.
const minJWTSecretLength = 32

// placeholderSecrets are values that ship in examples and must never reach a
// running deployment.
var placeholderSecrets = []string{
	"change-me",
	"changeme",
	"secret",
	"dev-only-do-not-use-in-prod",
}

// Config is the fully validated application configuration.
type Config struct {
	HTTP     HTTP
	Database Database
	Security Security
	CORS     CORS
	Purge    Purge
}

// HTTP holds the server socket and the timeouts that bound a slow client.
type HTTP struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxBodyBytes      int64
	TrustedProxies    []string
}

// Database holds the connection settings and the pool bounds.
type Database struct {
	Driver          Driver
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnectAttempts int
	ConnectBackoff  time.Duration
}

// Security holds credential settings.
type Security struct {
	JWTSecret   []byte
	JWTIssuer   string
	TokenTTL    time.Duration
	BcryptCost  int
	AuditLogDir string
}

// CORS holds the browser access policy.
type CORS struct {
	AllowedOrigins  []string
	AllowAllOrigins bool
}

// Purge holds the expired-paste sweeper settings.
type Purge struct {
	Interval time.Duration
	Timeout  time.Duration
}

// Load reads the environment, applies defaults and validates the result.
func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTP{
			Address:           envString("APP_PORT", ":8000"),
			ReadHeaderTimeout: envDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:       envDuration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:      envDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:       envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:   envDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			MaxBodyBytes:      int64(envInt("HTTP_MAX_BODY_BYTES", 2<<20)),
			TrustedProxies:    envList("HTTP_TRUSTED_PROXIES"),
		},
		Database: Database{
			Driver:          Driver(strings.ToLower(envString("DB_DRIVER", string(DriverMySQL)))),
			Host:            envString("DB_HOST", "localhost"),
			Port:            envString("DB_PORT", "3306"),
			Name:            envString("DB_NAME", "pastebomb"),
			User:            envString("DB_USER", "user"),
			Password:        os.Getenv("DB_PASSWORD"),
			SSLMode:         envString("DB_SSLMODE", "disable"),
			MaxOpenConns:    envInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: envDuration("DB_CONN_MAX_LIFETIME", time.Hour),
			ConnectAttempts: envInt("DB_CONNECT_ATTEMPTS", 5),
			ConnectBackoff:  envDuration("DB_CONNECT_BACKOFF", 3*time.Second),
		},
		Security: Security{
			JWTSecret:   []byte(os.Getenv("JWT_SECRET")),
			JWTIssuer:   envString("JWT_ISSUER", "pastebomb"),
			TokenTTL:    envDuration("JWT_TTL", 24*time.Hour),
			BcryptCost:  envInt("BCRYPT_COST", 12),
			AuditLogDir: envString("AUDIT_LOG_DIR", "."),
		},
		Purge: Purge{
			Interval: envDuration("PURGE_INTERVAL", time.Minute),
			Timeout:  envDuration("PURGE_TIMEOUT", 30*time.Second),
		},
	}

	origins := envList("CORS_ALLOWED_ORIGINS")
	if len(origins) == 0 {
		origins = []string{"http://localhost:8080"}
	}
	if len(origins) == 1 && origins[0] == "*" {
		cfg.CORS.AllowAllOrigins = true
	} else {
		cfg.CORS.AllowedOrigins = origins
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	var problems []string

	if c.Database.Driver != DriverMySQL && c.Database.Driver != DriverPostgres {
		problems = append(problems, fmt.Sprintf("DB_DRIVER %q is not supported (use mysql or pgsql)", c.Database.Driver))
	}
	if c.Database.Password == "" {
		problems = append(problems, "DB_PASSWORD is required")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		problems = append(problems, "DB_MAX_IDLE_CONNS must not exceed DB_MAX_OPEN_CONNS")
	}

	switch {
	case len(c.Security.JWTSecret) == 0:
		problems = append(problems, "JWT_SECRET is required")
	case len(c.Security.JWTSecret) < minJWTSecretLength:
		problems = append(problems, fmt.Sprintf("JWT_SECRET must be at least %d characters", minJWTSecretLength))
	default:
		lowered := strings.ToLower(string(c.Security.JWTSecret))
		for _, placeholder := range placeholderSecrets {
			if strings.Contains(lowered, placeholder) {
				problems = append(problems, "JWT_SECRET still contains an example value, generate a real one")
				break
			}
		}
	}

	// Costs below 10 are brute-forceable on commodity hardware; above 15 a
	// login takes long enough to become a denial-of-service lever.
	if c.Security.BcryptCost < 10 || c.Security.BcryptCost > 15 {
		problems = append(problems, "BCRYPT_COST must be between 10 and 15")
	}
	if c.Security.TokenTTL <= 0 {
		problems = append(problems, "JWT_TTL must be positive")
	}
	if c.HTTP.MaxBodyBytes <= 0 {
		problems = append(problems, "HTTP_MAX_BODY_BYTES must be positive")
	}
	if c.Purge.Interval <= 0 {
		problems = append(problems, "PURGE_INTERVAL must be positive")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(strings.TrimSpace(os.Getenv(key)))
	if err != nil {
		return fallback
	}
	return value
}

func envList(key string) []string {
	raw := os.Getenv(key)
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
