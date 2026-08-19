package persistence

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// notExpired is the predicate every read path shares: a paste past its
// deadline is invisible even before the purge job removes the row.
const notExpired = "expiration_date IS NULL OR expiration_date > ?"

// PasteRepository is the GORM-backed implementation of paste.Repository.
type PasteRepository struct{ db *gorm.DB }

func NewPasteRepository(db *gorm.DB) *PasteRepository { return &PasteRepository{db: db} }

var _ paste.Repository = (*PasteRepository)(nil)

func (r *PasteRepository) Create(ctx context.Context, document *paste.Paste) error {
	record := pasteRecordFrom(document.ToSnapshot())

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return paste.ErrSlugTaken
		}
		return shared.Internal("paste.create_failed", "could not create paste").WithCause(err)
	}

	document.AssignID(paste.ID(record.ID))
	return nil
}

func (r *PasteRepository) BySlug(ctx context.Context, slug paste.Slug) (*paste.Paste, error) {
	var record pasteRecord
	err := r.db.WithContext(ctx).Where("url_id = ?", slug.String()).Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, paste.ErrNotFound
		}
		return nil, shared.Internal("paste.query_failed", "could not load paste").WithCause(err)
	}
	return record.toAggregate(), nil
}

func (r *PasteRepository) ListPublic(ctx context.Context, page shared.Page, now time.Time) ([]*paste.Paste, error) {
	query := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Where(notExpired, now)
	return r.findMany(query, page)
}

func (r *PasteRepository) ListByOwner(ctx context.Context, owner paste.OwnerID, page shared.Page, now time.Time) ([]*paste.Paste, error) {
	query := r.db.WithContext(ctx).
		Where("user_id = ?", int64(owner)).
		Where(notExpired, now)
	return r.findMany(query, page)
}

// DeleteExpired removes pastes whose deadline has passed. The where clause is
// served by the index on expiration_date, so the sweep stays cheap.
func (r *PasteRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expiration_date IS NOT NULL AND expiration_date <= ?", before).
		Delete(&pasteRecord{})
	if result.Error != nil {
		return 0, shared.Internal("paste.purge_failed", "could not purge expired pastes").WithCause(result.Error)
	}
	return result.RowsAffected, nil
}

func (r *PasteRepository) findMany(query *gorm.DB, page shared.Page) ([]*paste.Paste, error) {
	var records []pasteRecord
	if err := query.Order("id DESC").Limit(page.Limit()).Offset(page.Offset()).Find(&records).Error; err != nil {
		return nil, shared.Internal("paste.query_failed", "could not list pastes").WithCause(err)
	}

	documents := make([]*paste.Paste, 0, len(records))
	for _, record := range records {
		documents = append(documents, record.toAggregate())
	}
	return documents, nil
}
