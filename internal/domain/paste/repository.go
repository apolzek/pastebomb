package paste

import (
	"context"
	"time"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// Repository is the persistence port for the Paste aggregate.
//
// Listing methods exclude expired pastes at the query level so a delayed purge
// never widens the window in which stale content is served.
type Repository interface {
	Create(ctx context.Context, p *Paste) error

	// BySlug is the only single-paste lookup. There is deliberately no ByID:
	// exposing sequential ids would make private pastes enumerable.
	BySlug(ctx context.Context, slug Slug) (*Paste, error)

	ListPublic(ctx context.Context, page shared.Page, now time.Time) ([]*Paste, error)
	ListByOwner(ctx context.Context, owner OwnerID, page shared.Page, now time.Time) ([]*Paste, error)

	// DeleteExpired removes pastes whose deadline passed before the given
	// instant and reports how many rows went away.
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
