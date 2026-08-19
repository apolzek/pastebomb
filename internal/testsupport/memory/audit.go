package memory

import (
	"context"
	"sync"

	"github.com/apolzek/pastebomb/internal/application"
)

// AuditRecorder collects audit events in memory so tests can assert that a
// security-relevant action was recorded.
type AuditRecorder struct {
	mu     sync.Mutex
	events []application.AuditEvent
}

func NewAuditRecorder() *AuditRecorder { return &AuditRecorder{} }

var _ application.AuditRecorder = (*AuditRecorder)(nil)

func (r *AuditRecorder) Record(_ context.Context, event application.AuditEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
}

// Actions returns the recorded action names in order.
func (r *AuditRecorder) Actions() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	actions := make([]string, 0, len(r.events))
	for _, event := range r.events {
		actions = append(actions, event.Action)
	}
	return actions
}
