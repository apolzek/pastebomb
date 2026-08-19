// Package audit writes the security audit trail to a rotating file.
package audit

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/apolzek/pastebomb/internal/application"
)

// sensitiveKeys name fields that must never be written verbatim. The match is
// a substring check, so "user_password" and "apiKey" are both caught.
var sensitiveKeys = []string{
	"password", "passwd", "pass", "secret", "token", "jwt",
	"authorization", "api_key", "apikey", "credential",
}

const redacted = "[REDACTED]"

// entry is the on-disk shape of an audit record.
type entry struct {
	Level   application.AuditLevel `json:"level"`
	Actor   string                 `json:"actor"`
	Action  string                 `json:"action"`
	When    time.Time              `json:"when"`
	Details map[string]any         `json:"details,omitempty"`
}

// FileRecorder appends JSON lines to a rotating log file.
type FileRecorder struct {
	mu     sync.Mutex
	writer io.WriteCloser
}

// NewFileRecorder opens audit.log inside dir, rotating at 5 MiB and keeping
// three compressed generations for a month.
func NewFileRecorder(dir string) *FileRecorder {
	return &FileRecorder{
		writer: &lumberjack.Logger{
			Filename:   filepath.Join(dir, "audit.log"),
			MaxSize:    5,
			MaxBackups: 3,
			MaxAge:     30,
			Compress:   true,
		},
	}
}

// Record writes one entry. Audit failures are logged but never propagated:
// losing an audit line must not turn a successful request into a 500.
func (r *FileRecorder) Record(_ context.Context, event application.AuditEvent) {
	payload, err := json.Marshal(entry{
		Level:   event.Level,
		Actor:   event.Actor,
		Action:  event.Action,
		When:    time.Now().UTC(),
		Details: scrub(event.Details),
	})
	if err != nil {
		slog.Error("audit entry could not be encoded", "action", event.Action, "err", err)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.writer.Write(append(payload, '\n')); err != nil {
		slog.Error("audit entry could not be written", "action", event.Action, "err", err)
	}
}

// Close flushes and releases the underlying file.
func (r *FileRecorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.writer.Close()
}

// scrub replaces credential-shaped values, walking nested maps so a secret
// buried one level down is caught too.
func scrub(details map[string]any) map[string]any {
	if details == nil {
		return nil
	}
	out := make(map[string]any, len(details))
	for key, value := range details {
		switch {
		case isSensitive(key):
			out[key] = redacted
		default:
			if nested, ok := value.(map[string]any); ok {
				out[key] = scrub(nested)
				continue
			}
			out[key] = value
		}
	}
	return out
}

func isSensitive(key string) bool {
	lowered := strings.ToLower(key)
	for _, candidate := range sensitiveKeys {
		if strings.Contains(lowered, candidate) {
			return true
		}
	}
	return false
}
