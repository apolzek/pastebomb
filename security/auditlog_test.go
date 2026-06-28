package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScrubSensitive_RedactsCommonKeys(t *testing.T) {
	in := map[string]interface{}{
		"username":      "alice",
		"password":      "hunter2",
		"jwt_token":     "eyJ...",
		"api_key":       "abc",
		"Authorization": "Bearer x",
		"safe":          "ok",
	}
	out := scrubSensitive(in).(map[string]interface{})

	assert.Equal(t, "alice", out["username"])
	assert.Equal(t, "ok", out["safe"])
	assert.Equal(t, "[REDACTED]", out["password"])
	assert.Equal(t, "[REDACTED]", out["jwt_token"])
	assert.Equal(t, "[REDACTED]", out["api_key"])
	assert.Equal(t, "[REDACTED]", out["Authorization"])
}

func TestScrubSensitive_NestedMap(t *testing.T) {
	in := map[string]interface{}{
		"user": map[string]interface{}{
			"name":     "bob",
			"password": "leak-me",
		},
	}
	out := scrubSensitive(in).(map[string]interface{})
	nested := out["user"].(map[string]interface{})
	assert.Equal(t, "bob", nested["name"])
	assert.Equal(t, "[REDACTED]", nested["password"])
}

func TestScrubSensitive_PassesNonMapsThrough(t *testing.T) {
	assert.Equal(t, "hello", scrubSensitive("hello"))
	assert.Equal(t, 42, scrubSensitive(42))
	assert.Nil(t, scrubSensitive(nil))
}
