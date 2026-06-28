package security

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel define os tipos de log válidos
type LogLevel string

const (
	LogLevelInfo      LogLevel = "INFO"
	LogLevelDebug     LogLevel = "DEBUG"
	LogLevelError     LogLevel = "ERROR"
	LogLevelException LogLevel = "EXCEPTION"
)

type AuditLogEntry struct {
	Level   LogLevel    `json:"level"`
	Actor   string      `json:"actor"`
	Action  string      `json:"action"`
	When    time.Time   `json:"when"`
	Details interface{} `json:"details"`
}

// sensitiveKeys are scrubbed from any map[string]interface{} passed as
// Details so we never accidentally land a password/token/JWT in audit.log.
var sensitiveKeys = []string{
	"password", "pass", "secret", "token", "jwt", "authorization", "api_key", "apikey",
}

var (
	writerOnce sync.Once
	writer     io.Writer
)

func auditWriter() io.Writer {
	writerOnce.Do(func() {
		// 5MB per file, keep 3 rotated copies, compress old ones.
		writer = &lumberjack.Logger{
			Filename:   "audit.log",
			MaxSize:    5,
			MaxBackups: 3,
			MaxAge:     30,
			Compress:   true,
		}
	})
	return writer
}

// LogAuditEvent appends a JSON line to audit.log with sensitive fields scrubbed.
func LogAuditEvent(level LogLevel, actor string, action string, details interface{}) {
	entry := AuditLogEntry{
		Level:   level,
		Actor:   actor,
		Action:  action,
		When:    time.Now(),
		Details: scrubSensitive(details),
	}

	logEntry, err := json.Marshal(entry)
	if err != nil {
		return
	}

	_, _ = auditWriter().Write(append(logEntry, '\n'))
}

// scrubSensitive walks string-keyed maps and replaces values whose key looks
// like a credential with "[REDACTED]". Other types pass through untouched.
func scrubSensitive(v interface{}) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok {
		return v
	}
	out := make(map[string]interface{}, len(m))
	for k, val := range m {
		if isSensitiveKey(k) {
			out[k] = "[REDACTED]"
		} else {
			out[k] = scrubSensitive(val)
		}
	}
	return out
}

func isSensitiveKey(k string) bool {
	lower := strings.ToLower(k)
	for _, s := range sensitiveKeys {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
