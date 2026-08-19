package audit_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/infrastructure/audit"
)

func recordAndRead(t *testing.T, events ...application.AuditEvent) []map[string]any {
	t.Helper()

	dir := t.TempDir()
	recorder := audit.NewFileRecorder(dir)
	for _, event := range events {
		recorder.Record(context.Background(), event)
	}
	require.NoError(t, recorder.Close())

	raw, err := os.ReadFile(filepath.Join(dir, "audit.log"))
	require.NoError(t, err)

	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		entries = append(entries, entry)
	}
	return entries
}

func TestRecordWritesOneJSONLinePerEvent(t *testing.T) {
	t.Parallel()

	entries := recordAndRead(t,
		application.AuditEvent{Level: application.AuditInfo, Actor: "ada", Action: "login.succeeded"},
		application.AuditEvent{Level: application.AuditWarning, Actor: "eve", Action: "login.failed"},
	)

	require.Len(t, entries, 2)
	assert.Equal(t, "login.succeeded", entries[0]["action"])
	assert.Equal(t, "INFO", entries[0]["level"])
	assert.Equal(t, "eve", entries[1]["actor"])
	assert.NotEmpty(t, entries[0]["when"])
}

// The audit trail is read by humans during incidents. A credential that lands
// in it turns a log file into a second copy of the password database.
func TestRecordRedactsCredentialShapedFields(t *testing.T) {
	t.Parallel()

	entries := recordAndRead(t, application.AuditEvent{
		Level:  application.AuditInfo,
		Actor:  "ada",
		Action: "login.attempt",
		Details: map[string]any{
			"password":      "hunter2",
			"API_Key":       "live_abc123",
			"authorization": "Bearer xyz",
			"user_secret":   "s3cret",
			"jwt":           "eyJhbGciOi",
			"user_id":       int64(7),
			"nested": map[string]any{
				"refresh_token": "rt_123",
				"attempt":       float64(2),
			},
		},
	})

	require.Len(t, entries, 1)
	details, ok := entries[0]["details"].(map[string]any)
	require.True(t, ok)

	for _, key := range []string{"password", "API_Key", "authorization", "user_secret", "jwt"} {
		assert.Equal(t, "[REDACTED]", details[key], "field %q was not redacted", key)
	}
	assert.EqualValues(t, 7, details["user_id"])

	nested, ok := details["nested"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "[REDACTED]", nested["refresh_token"])
	assert.EqualValues(t, 2, nested["attempt"])
}

func TestRecordHandlesEventsWithoutDetails(t *testing.T) {
	t.Parallel()

	entries := recordAndRead(t, application.AuditEvent{
		Level:  application.AuditError,
		Actor:  "system",
		Action: "purge.failed",
	})

	require.Len(t, entries, 1)
	_, present := entries[0]["details"]
	assert.False(t, present)
}
