// Package application holds the use cases. It orchestrates domain aggregates
// and outbound ports, and it is the only layer both the HTTP adapter and the
// background jobs are allowed to call.
package application

import (
	"context"
	"time"

	"github.com/apolzek/pastebomb/internal/domain/user"
)

// Identity is the authenticated principal. It is what a token carries and
// what handlers pass back into the use cases.
type Identity struct {
	UserID user.ID
	Email  string
	Name   string
	Admin  bool
}

// IsAnonymous reports whether no user is attached to the request.
func (i Identity) IsAnonymous() bool { return i.UserID.IsZero() }

// Token is a freshly issued credential plus its expiry, so the client can
// refresh before it lapses instead of discovering the failure on a 401.
type Token struct {
	Value     string
	ExpiresAt time.Time
}

// TokenIssuer mints credentials for an authenticated identity.
type TokenIssuer interface {
	Issue(identity Identity) (Token, error)
}

// TokenVerifier validates an incoming credential. Implementations must reject
// unexpected signing algorithms rather than trusting the token header.
type TokenVerifier interface {
	Verify(rawToken string) (Identity, error)
}

// AuditLevel classifies an audit entry.
type AuditLevel string

const (
	AuditInfo    AuditLevel = "INFO"
	AuditWarning AuditLevel = "WARNING"
	AuditError   AuditLevel = "ERROR"
)

// AuditEvent is one entry in the security audit trail.
type AuditEvent struct {
	Level   AuditLevel
	Actor   string
	Action  string
	Details map[string]any
}

// AuditRecorder appends to the audit trail. Implementations must never block
// the request path and must scrub credential-shaped fields.
type AuditRecorder interface {
	Record(ctx context.Context, event AuditEvent)
}
