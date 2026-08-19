package application

import (
	"context"
	"errors"

	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// CreatePasteCommand carries a paste submission. Owner is zero for anonymous
// pastes, which the public endpoint always forces.
type CreatePasteCommand struct {
	Owner     paste.OwnerID
	Title     string
	Content   string
	Category  string
	IsPublic  bool
	ExpiresAt string
	Secret    string
}

// slugCollisionRetries covers the astronomically unlikely case of two pastes
// drawing the same random handle.
const slugCollisionRetries = 3

// PasteService implements the paste use cases.
type PasteService struct {
	pastes  paste.Repository
	factory *paste.Factory
	guard   *paste.AccessGuard
	clock   shared.Clock
	audit   AuditRecorder
}

func NewPasteService(
	pastes paste.Repository,
	factory *paste.Factory,
	guard *paste.AccessGuard,
	clock shared.Clock,
	audit AuditRecorder,
) *PasteService {
	return &PasteService{pastes: pastes, factory: factory, guard: guard, clock: clock, audit: audit}
}

// Create stores a new paste and returns it with its generated slug.
func (s *PasteService) Create(ctx context.Context, cmd CreatePasteCommand) (*paste.Paste, error) {
	var lastErr error
	for attempt := 0; attempt < slugCollisionRetries; attempt++ {
		document, err := s.factory.Create(paste.Draft(cmd))
		if err != nil {
			return nil, err
		}

		err = s.pastes.Create(ctx, document)
		switch {
		case err == nil:
			s.audit.Record(ctx, AuditEvent{
				Level:  AuditInfo,
				Actor:  actorLabel(cmd.Owner),
				Action: "paste.created",
				Details: map[string]any{
					"slug":      document.Slug().String(),
					"is_public": document.Visibility().IsPublic(),
					// Named "protected" rather than "has_secret" so the audit
					// scrubber, which matches key substrings, does not redact
					// a boolean that carries nothing sensitive.
					"protected": document.HasSecret(),
				},
			})
			return document, nil
		case errors.Is(err, paste.ErrSlugTaken):
			lastErr = err
			continue
		default:
			return nil, err
		}
	}
	return nil, shared.Internal("paste.slug_exhausted", "could not allocate a unique paste handle").WithCause(lastErr)
}

// Read returns a paste by its public handle, enforcing expiration and the
// optional access secret.
func (s *PasteService) Read(ctx context.Context, rawSlug, suppliedSecret string) (*paste.Paste, error) {
	slug, err := paste.NewSlug(rawSlug)
	if err != nil {
		// A malformed handle simply cannot match anything that exists.
		return nil, paste.ErrNotFound
	}

	document, err := s.pastes.BySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if err := s.guard.Authorize(document, suppliedSecret); err != nil {
		return nil, err
	}
	return document, nil
}

// ListPublic returns the public index. Bodies are intentionally not part of
// this projection: the handler renders metadata only, so a secret-protected
// paste never leaks its content through the index.
func (s *PasteService) ListPublic(ctx context.Context, page shared.Page) ([]*paste.Paste, error) {
	return s.pastes.ListPublic(ctx, page, s.clock.Now())
}

// ListOwned returns the pastes belonging to one user.
func (s *PasteService) ListOwned(ctx context.Context, owner paste.OwnerID, page shared.Page) ([]*paste.Paste, error) {
	if owner.IsAnonymous() {
		return nil, paste.ErrForbidden
	}
	return s.pastes.ListByOwner(ctx, owner, page, s.clock.Now())
}

// PurgeExpired deletes pastes whose deadline has passed. Driven by a
// background ticker, it is safe to call concurrently.
func (s *PasteService) PurgeExpired(ctx context.Context) (int64, error) {
	return s.pastes.DeleteExpired(ctx, s.clock.Now())
}

func actorLabel(owner paste.OwnerID) string {
	if owner.IsAnonymous() {
		return "anonymous"
	}
	return "user"
}
