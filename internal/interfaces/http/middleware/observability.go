package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/interfaces/http/reqctx"

	"github.com/apolzek/pastebomb/internal/infrastructure/security"
	"github.com/apolzek/pastebomb/internal/interfaces/http/dto"
)

const requestIDHeader = "X-Request-ID"

// requestIDLength is short: the id correlates log lines, it is not a secret.
const requestIDLength = 12

// RequestID assigns a correlation id to every request and echoes it back.
//
// An inbound X-Request-ID is deliberately ignored: honouring a client-supplied
// value would let a caller forge or collide log entries.
func RequestID() gin.HandlerFunc {
	generator := security.NewRandomSlugGenerator(requestIDLength)
	return func(ctx *gin.Context) {
		id := generator.Generate().String()
		reqctx.SetRequestID(ctx, id)
		ctx.Header(requestIDHeader, id)
		ctx.Next()
	}
}

// RequestLogger emits one structured line per request. The query string is
// omitted on purpose, since secrets are sometimes pasted into URLs.
func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		started := time.Now()
		ctx.Next()

		attrs := []any{
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"status", ctx.Writer.Status(),
			"duration_ms", time.Since(started).Milliseconds(),
			"client_ip", ctx.ClientIP(),
			"request_id", reqctx.RequestID(ctx),
		}

		switch {
		case ctx.Writer.Status() >= http.StatusInternalServerError:
			slog.Error("request failed", attrs...)
		case ctx.Writer.Status() >= http.StatusBadRequest:
			slog.Warn("request rejected", attrs...)
		default:
			slog.Info("request served", attrs...)
		}
	}
}

// Recovery converts a panic into a 500 without leaking the stack to the
// client, while logging enough to debug it.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(ctx *gin.Context, recovered any) {
		slog.Error("panic recovered",
			"path", ctx.Request.URL.Path,
			"request_id", reqctx.RequestID(ctx),
			"panic", recovered,
		)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message:   "internal server error",
			Code:      "internal_error",
			RequestID: reqctx.RequestID(ctx),
		})
	})
}

// SecureHeaders sets the response headers that cost nothing and close off a
// class of browser-side attacks.
//
//   - nosniff stops a text/plain paste from being re-interpreted as HTML or
//     JavaScript, which is the main cross-site scripting vector for a service
//     that serves user-supplied content verbatim.
//   - The frame and referrer headers block clickjacking and referrer leakage
//     of unlisted paste handles.
func SecureHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; sandbox")
		header.Set("Cross-Origin-Resource-Policy", "same-origin")
		ctx.Next()
	}
}

// BodyLimit caps how much a client may send. Without it a single request can
// stream gigabytes into memory before any handler gets a say.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBytes)
		ctx.Next()
	}
}
