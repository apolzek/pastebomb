package security

import (
	"encoding/json"
	"os"
	"time"
)

// LogLevel define os tipos de log válidos
type LogLevel string

const (
	// LogLevelInfo indica um log de informação
	LogLevelInfo LogLevel = "INFO"
	// LogLevelDebug indica um log de depuração
	LogLevelDebug LogLevel = "DEBUG"
	// LogLevelError indica um log de erro
	LogLevelError LogLevel = "ERROR"
	// LogLevelException indica um log de exceção
	LogLevelException LogLevel = "EXCEPTION"
)

// AuditLogEntry representa uma entrada no log de auditoria
type AuditLogEntry struct {
	Level   LogLevel    `json:"level"`
	Actor   string      `json:"actor"`
	Action  string      `json:"action"`
	When    time.Time   `json:"when"`
	Details interface{} `json:"details"`
}

// LogAuditEvent registra um evento no log de auditoria
func LogAuditEvent(level LogLevel, actor string, action string, details interface{}) {
	entry := AuditLogEntry{
		Level:   level,
		Actor:   actor,
		Action:  action,
		When:    time.Now(),
		Details: details,
	}

	logEntry, _ := json.Marshal(entry)
	file, _ := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()

	file.WriteString(string(logEntry) + "\n")
}
