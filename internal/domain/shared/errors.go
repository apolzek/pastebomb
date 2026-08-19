// Package shared holds the primitives every aggregate depends on: the error
// vocabulary, pagination and the clock. It must never import another layer.
package shared

import (
	"errors"
	"fmt"
)

// Kind classifies a domain error so outer layers (HTTP, CLI, workers) can map
// it to their own vocabulary without string matching.
type Kind uint8

const (
	KindInternal Kind = iota
	KindInvalidInput
	KindNotFound
	KindConflict
	KindUnauthorized
	KindForbidden
)

// Error is the single error type crossing layer boundaries. Message is safe to
// return to a client; cause is kept for logs and never serialized.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// WithCause attaches the underlying failure for logging purposes.
func (e *Error) WithCause(cause error) *Error {
	clone := *e
	clone.cause = cause
	return &clone
}

func newError(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

func Invalid(code, message string) *Error      { return newError(KindInvalidInput, code, message) }
func NotFound(code, message string) *Error     { return newError(KindNotFound, code, message) }
func Conflict(code, message string) *Error     { return newError(KindConflict, code, message) }
func Unauthorized(code, message string) *Error { return newError(KindUnauthorized, code, message) }
func Forbidden(code, message string) *Error    { return newError(KindForbidden, code, message) }
func Internal(code, message string) *Error     { return newError(KindInternal, code, message) }

// KindOf reports the classification of err, defaulting to KindInternal for
// errors that did not originate in the domain.
func KindOf(err error) Kind {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr.Kind
	}
	return KindInternal
}

// AsError returns the domain error carried by err, if any.
func AsError(err error) (*Error, bool) {
	var domainErr *Error
	ok := errors.As(err, &domainErr)
	return domainErr, ok
}
