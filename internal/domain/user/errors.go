package user

import "github.com/apolzek/pastebomb/internal/domain/shared"

var (
	// ErrNotFound is returned by repositories when no row matches.
	ErrNotFound = shared.NotFound("user.not_found", "user not found")

	// ErrInvalidCredentials is deliberately vague: it covers a missing user,
	// a deactivated user and a wrong password alike.
	ErrInvalidCredentials = shared.Unauthorized("user.invalid_credentials", "invalid email or password")

	ErrEmailTaken    = shared.Conflict("user.email_taken", "email already in use")
	ErrUsernameTaken = shared.Conflict("user.username_taken", "username already in use")
	ErrForbidden     = shared.Forbidden("user.forbidden", "you may only act on your own account")
)
