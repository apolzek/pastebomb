package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/interfaces/http/reqctx"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/interfaces/http/httperr"
)

const bearerPrefix = "Bearer "

var (
	errMissingToken = shared.Unauthorized("auth.token_missing", "authentication token is required")
	errInvalidToken = shared.Unauthorized("auth.token_invalid", "authentication token is invalid or expired")
	errNotAdmin     = shared.Forbidden("auth.forbidden", "administrator rights are required")
)

// Authenticate rejects any request without a valid bearer token.
func Authenticate(verifier application.TokenVerifier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawToken, ok := bearerToken(ctx)
		if !ok {
			httperr.Abort(ctx, errMissingToken)
			return
		}

		identity, err := verifier.Verify(rawToken)
		if err != nil {
			httperr.Abort(ctx, errInvalidToken)
			return
		}

		reqctx.SetIdentity(ctx, identity)
		ctx.Next()
	}
}

// RequireAdmin gates administrative routes on the admin claim carried by the
// token. It replaces the previous check, which trusted a plain "owner" header
// that any client could set.
func RequireAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !reqctx.Identity(ctx).Admin {
			httperr.Abort(ctx, errNotAdmin)
			return
		}
		ctx.Next()
	}
}

// bearerToken extracts the credential from the Authorization header. The
// scheme is matched case-insensitively per RFC 7235, but nothing else about
// the header is trusted.
func bearerToken(ctx *gin.Context) (string, bool) {
	header := ctx.GetHeader("Authorization")
	if len(header) <= len(bearerPrefix) {
		return "", false
	}
	if !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(bearerPrefix):])
	return token, token != ""
}
