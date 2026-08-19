// Package reqctx owns the values middleware attaches to a request and the
// handlers read back. It exists so the middleware, the handlers and the error
// renderer can share those keys without importing one another, and without the
// mutable package-level state that a registration hook would need.
package reqctx

import (
	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/application"
)

// Keys are unexported constants rather than raw strings spread across the
// codebase, so a typo becomes a compile error instead of a silent miss.
const (
	identityKey  = "auth.identity"
	requestIDKey = "request.id"
)

// SetIdentity stores the authenticated principal on the request.
func SetIdentity(ctx *gin.Context, identity application.Identity) {
	ctx.Set(identityKey, identity)
}

// Identity returns the authenticated principal. The zero value means the
// request is anonymous, which callers must treat as unauthenticated.
func Identity(ctx *gin.Context) application.Identity {
	value, exists := ctx.Get(identityKey)
	if !exists {
		return application.Identity{}
	}
	identity, ok := value.(application.Identity)
	if !ok {
		return application.Identity{}
	}
	return identity
}

// SetRequestID stores the correlation id of the request.
func SetRequestID(ctx *gin.Context, id string) {
	ctx.Set(requestIDKey, id)
}

// RequestID returns the correlation id, or the empty string outside a request
// that went through the middleware.
func RequestID(ctx *gin.Context) string {
	value, exists := ctx.Get(requestIDKey)
	if !exists {
		return ""
	}
	id, _ := value.(string)
	return id
}
