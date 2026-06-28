package cors_config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowedOrigins_Default(t *testing.T) {
	t.Setenv(envAllowedOrigins, "")
	assert.Equal(t, []string{"http://localhost:8080"}, allowedOrigins())
}

func TestAllowedOrigins_FromEnv(t *testing.T) {
	t.Setenv(envAllowedOrigins, "https://a.com, https://b.com ,https://c.com")
	got := allowedOrigins()
	assert.Equal(t, []string{"https://a.com", "https://b.com", "https://c.com"}, got)
}

func TestAllowedOrigins_Wildcard(t *testing.T) {
	t.Setenv(envAllowedOrigins, "*")
	assert.Equal(t, []string{"*"}, allowedOrigins())
}

func TestAllowedOrigins_EmptyEntriesIgnored(t *testing.T) {
	t.Setenv(envAllowedOrigins, ",,https://only.example.com,, ,")
	assert.Equal(t, []string{"https://only.example.com"}, allowedOrigins())
}
