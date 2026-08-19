// Package httperr is the single place where a domain error becomes an HTTP
// response. Handlers never choose a status code by hand, so a new error kind
// cannot accidentally surface as a 200 or leak an internal message.
package httperr

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/interfaces/http/dto"
	"github.com/apolzek/pastebomb/internal/interfaces/http/reqctx"
)

// statusByKind maps the domain vocabulary to HTTP semantics.
var statusByKind = map[shared.Kind]int{
	shared.KindInvalidInput: http.StatusBadRequest,
	shared.KindUnauthorized: http.StatusUnauthorized,
	shared.KindForbidden:    http.StatusForbidden,
	shared.KindNotFound:     http.StatusNotFound,
	shared.KindConflict:     http.StatusConflict,
	shared.KindInternal:     http.StatusInternalServerError,
}

// Write renders err as JSON. Internal failures are logged in full and reported
// to the client as a generic message, so a database error never reaches the
// caller as a stack trace or a SQL fragment.
func Write(ctx *gin.Context, err error) {
	status, body := translate(ctx, err)
	ctx.JSON(status, body)
}

// Abort renders err and stops the handler chain. Used by middleware.
func Abort(ctx *gin.Context, err error) {
	status, body := translate(ctx, err)
	ctx.AbortWithStatusJSON(status, body)
}

// WriteText renders err as plain text, for the raw paste endpoint.
func WriteText(ctx *gin.Context, err error) {
	status, body := translate(ctx, err)
	ctx.String(status, body.Message)
}

func translate(ctx *gin.Context, err error) (int, dto.ErrorResponse) {
	requestID := reqctx.RequestID(ctx)

	domainErr, ok := shared.AsError(err)
	if !ok {
		slog.Error("unhandled error", "path", ctx.FullPath(), "request_id", requestID, "err", err)
		return http.StatusInternalServerError, dto.ErrorResponse{
			Message:   "internal server error",
			Code:      "internal_error",
			RequestID: requestID,
		}
	}

	status, known := statusByKind[domainErr.Kind]
	if !known {
		status = http.StatusInternalServerError
	}

	if status >= http.StatusInternalServerError {
		slog.Error("internal error",
			"path", ctx.FullPath(),
			"request_id", requestID,
			"code", domainErr.Code,
			"err", domainErr,
		)
		return status, dto.ErrorResponse{
			Message:   "internal server error",
			Code:      "internal_error",
			RequestID: requestID,
		}
	}

	return status, dto.ErrorResponse{
		Message:   domainErr.Message,
		Code:      domainErr.Code,
		RequestID: requestID,
	}
}

// Invalid wraps a binding failure. The validator message describes the payload
// the client already sent, so echoing it back leaks nothing.
func Invalid(ctx *gin.Context, err error) {
	Write(ctx, shared.Invalid("request.malformed", "request payload is invalid: "+err.Error()))
}
