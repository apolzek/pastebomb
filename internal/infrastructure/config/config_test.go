package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/infrastructure/config"
)

// validEnvironment is the smallest environment that loads cleanly.
func validEnvironment() map[string]string {
	return map[string]string{
		"DB_PASSWORD": "a-database-password",
		"JWT_SECRET":  "3f8c1d9b7a2e4c6f0b5d8e1a4c7f2b9d",
	}
}

// overridable lists every variable a test asserts a default for. Clearing them
// first keeps the suite independent of whatever the developer's shell exports.
var overridable = []string{
	"APP_PORT", "DB_DRIVER", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
	"BCRYPT_COST", "JWT_TTL", "CORS_ALLOWED_ORIGINS", "HTTP_MAX_BODY_BYTES",
	"HTTP_READ_HEADER_TIMEOUT", "HTTP_WRITE_TIMEOUT", "PURGE_INTERVAL",
}

func loadWith(t *testing.T, environment map[string]string) (config.Config, error) {
	t.Helper()
	for _, key := range overridable {
		t.Setenv(key, "")
	}
	for key, value := range environment {
		t.Setenv(key, value)
	}
	return config.Load()
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfg, err := loadWith(t, validEnvironment())
	require.NoError(t, err)

	assert.Equal(t, ":8000", cfg.HTTP.Address)
	assert.Equal(t, config.DriverMySQL, cfg.Database.Driver)
	assert.Equal(t, 12, cfg.Security.BcryptCost)
	assert.Equal(t, 24*time.Hour, cfg.Security.TokenTTL)
	assert.Equal(t, []string{"http://localhost:8080"}, cfg.CORS.AllowedOrigins)
	assert.False(t, cfg.CORS.AllowAllOrigins)
	// Timeouts must never default to zero, which http.Server reads as "never".
	assert.Positive(t, cfg.HTTP.ReadHeaderTimeout)
	assert.Positive(t, cfg.HTTP.WriteTimeout)
	assert.Positive(t, cfg.HTTP.MaxBodyBytes)
}

// Starting with a weak or placeholder signing key would let anyone mint a
// token for any account, so the process must refuse to boot instead.
func TestLoadRejectsAWeakOrMissingJWTSecret(t *testing.T) {
	cases := map[string]string{
		"missing":     "",
		"too short":   "short-secret",
		"placeholder": "change-me-to-a-long-random-string",
		"dev sample":  "dev-only-do-not-use-in-prod-padding",
	}

	for name, secret := range cases {
		t.Run(name, func(t *testing.T) {
			environment := validEnvironment()
			environment["JWT_SECRET"] = secret

			_, err := loadWith(t, environment)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "JWT_SECRET")
		})
	}
}

func TestLoadRejectsAMissingDatabasePassword(t *testing.T) {
	environment := validEnvironment()
	environment["DB_PASSWORD"] = ""

	_, err := loadWith(t, environment)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_PASSWORD")
}

func TestLoadRejectsAnUnknownDriver(t *testing.T) {
	environment := validEnvironment()
	environment["DB_DRIVER"] = "oracle"

	_, err := loadWith(t, environment)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_DRIVER")
}

func TestLoadRejectsAnUnsafeBcryptCost(t *testing.T) {
	for _, cost := range []string{"4", "20"} {
		environment := validEnvironment()
		environment["BCRYPT_COST"] = cost

		_, err := loadWith(t, environment)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BCRYPT_COST")
	}
}

func TestLoadReportsEveryProblemAtOnce(t *testing.T) {
	_, err := loadWith(t, map[string]string{
		"DB_PASSWORD": "",
		"JWT_SECRET":  "",
		"DB_DRIVER":   "oracle",
	})
	require.Error(t, err)

	message := err.Error()
	assert.Contains(t, message, "DB_PASSWORD")
	assert.Contains(t, message, "JWT_SECRET")
	assert.Contains(t, message, "DB_DRIVER")
	assert.Equal(t, 3, strings.Count(message, "\n  - "))
}

func TestLoadParsesTheOriginList(t *testing.T) {
	environment := validEnvironment()
	environment["CORS_ALLOWED_ORIGINS"] = " https://a.example , ,https://b.example "

	cfg, err := loadWith(t, environment)
	require.NoError(t, err)

	assert.Equal(t, []string{"https://a.example", "https://b.example"}, cfg.CORS.AllowedOrigins)
	assert.False(t, cfg.CORS.AllowAllOrigins)
}

func TestLoadRecognizesTheWildcardOrigin(t *testing.T) {
	environment := validEnvironment()
	environment["CORS_ALLOWED_ORIGINS"] = "*"

	cfg, err := loadWith(t, environment)
	require.NoError(t, err)

	assert.True(t, cfg.CORS.AllowAllOrigins)
	assert.Empty(t, cfg.CORS.AllowedOrigins)
}

func TestUnparsableNumericSettingsFallBackToDefaults(t *testing.T) {
	environment := validEnvironment()
	environment["DB_MAX_OPEN_CONNS"] = "plenty"
	environment["JWT_TTL"] = "a while"

	cfg, err := loadWith(t, environment)
	require.NoError(t, err)

	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 24*time.Hour, cfg.Security.TokenTTL)
}
