package paste

import "github.com/apolzek/pastebomb/internal/domain/shared"

var (
	// ErrNotFound also covers expired pastes, so a caller cannot tell an
	// expired paste apart from one that never existed.
	ErrNotFound = shared.NotFound("paste.not_found", "paste not found")

	// ErrSecretRequired is returned both for a missing and for a wrong secret.
	ErrSecretRequired = shared.Unauthorized("paste.secret_required", "a valid secret is required to read this paste")

	ErrSlugTaken = shared.Conflict("paste.slug_taken", "could not allocate a unique paste handle")
	ErrForbidden = shared.Forbidden("paste.forbidden", "you may only act on your own pastes")
)
