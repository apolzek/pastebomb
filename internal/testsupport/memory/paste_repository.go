package memory

import (
	"context"
	"sync"
	"time"

	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// PasteRepository stores pastes in a map keyed by slug.
type PasteRepository struct {
	mu       sync.RWMutex
	bySlug   map[string]paste.Snapshot
	order    []string
	sequence int64
}

func NewPasteRepository() *PasteRepository {
	return &PasteRepository{bySlug: make(map[string]paste.Snapshot)}
}

var _ paste.Repository = (*PasteRepository)(nil)

func (r *PasteRepository) Create(_ context.Context, document *paste.Paste) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	snapshot := document.ToSnapshot()
	if _, taken := r.bySlug[snapshot.Slug]; taken {
		return paste.ErrSlugTaken
	}

	r.sequence++
	snapshot.ID = r.sequence

	r.bySlug[snapshot.Slug] = snapshot
	r.order = append(r.order, snapshot.Slug)
	document.AssignID(paste.ID(snapshot.ID))
	return nil
}

func (r *PasteRepository) BySlug(_ context.Context, slug paste.Slug) (*paste.Paste, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot, exists := r.bySlug[slug.String()]
	if !exists {
		return nil, paste.ErrNotFound
	}
	return paste.Rehydrate(snapshot), nil
}

func (r *PasteRepository) ListPublic(_ context.Context, page shared.Page, now time.Time) ([]*paste.Paste, error) {
	return r.list(page, func(s paste.Snapshot) bool {
		return s.IsPublic && !expired(s, now)
	}), nil
}

func (r *PasteRepository) ListByOwner(_ context.Context, owner paste.OwnerID, page shared.Page, now time.Time) ([]*paste.Paste, error) {
	return r.list(page, func(s paste.Snapshot) bool {
		return s.OwnerID == int64(owner) && !expired(s, now)
	}), nil
}

func (r *PasteRepository) DeleteExpired(_ context.Context, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var deleted int64
	remaining := make([]string, 0, len(r.order))
	for _, slug := range r.order {
		if expired(r.bySlug[slug], before) {
			delete(r.bySlug, slug)
			deleted++
			continue
		}
		remaining = append(remaining, slug)
	}
	r.order = remaining
	return deleted, nil
}

func (r *PasteRepository) list(page shared.Page, match func(paste.Snapshot) bool) []*paste.Paste {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*paste.Paste
	// Newest first, matching the ORDER BY id DESC of the SQL adapter.
	for i := len(r.order) - 1; i >= 0; i-- {
		if snapshot := r.bySlug[r.order[i]]; match(snapshot) {
			matched = append(matched, paste.Rehydrate(snapshot))
		}
	}
	return paginate(matched, page)
}

func expired(snapshot paste.Snapshot, now time.Time) bool {
	return snapshot.ExpiresAt != nil && !now.Before(*snapshot.ExpiresAt)
}
