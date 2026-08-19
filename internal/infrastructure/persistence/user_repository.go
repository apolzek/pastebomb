package persistence

import (
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
)

// UserRepository is the GORM-backed implementation of user.Repository.
type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

var _ user.Repository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, account *user.User) error {
	record := userRecordFrom(account.ToSnapshot())

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		// The unique indexes are the real guard against a duplicate slipping
		// through between the availability check and the insert.
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return user.ErrEmailTaken
		}
		return shared.Internal("user.create_failed", "could not create user").WithCause(err)
	}

	account.AssignID(user.ID(record.ID))
	return nil
}

func (r *UserRepository) Update(ctx context.Context, account *user.User) error {
	record := userRecordFrom(account.ToSnapshot())

	result := r.db.WithContext(ctx).
		Model(&userRecord{}).
		Where("id = ?", record.ID).
		Updates(map[string]any{
			"name":       record.Name,
			"username":   record.Username,
			"email":      record.Email,
			"born_date":  record.BornDate,
			"password":   record.Password,
			"is_active":  record.IsActive,
			"updated_at": record.UpdatedAt,
		})

	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return user.ErrEmailTaken
		}
		return shared.Internal("user.update_failed", "could not update user").WithCause(err)
	}
	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) ByID(ctx context.Context, id user.ID) (*user.User, error) {
	return r.findOne(ctx, r.db.WithContext(ctx).Where("id = ?", int64(id)))
}

func (r *UserRepository) ByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	return r.findOne(ctx, r.db.WithContext(ctx).Where("email = ?", email.String()))
}

// ByHandle accepts either the numeric id or the username. The two are matched
// separately so a username made of digits cannot collide with an id.
func (r *UserRepository) ByHandle(ctx context.Context, handle string) (*user.User, error) {
	query := r.db.WithContext(ctx)
	if id, err := strconv.ParseInt(handle, 10, 64); err == nil {
		return r.findOne(ctx, query.Where("id = ?", id))
	}

	username, err := user.NewUsername(handle)
	if err != nil {
		return nil, user.ErrNotFound
	}
	return r.findOne(ctx, query.Where("username = ?", username.String()))
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email user.Email, excluding user.ID) (bool, error) {
	return r.exists(ctx, "email = ?", email.String(), excluding)
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username user.Username, excluding user.ID) (bool, error) {
	return r.exists(ctx, "username = ?", username.String(), excluding)
}

func (r *UserRepository) List(ctx context.Context, page shared.Page) ([]*user.User, error) {
	return r.findMany(ctx, r.db.WithContext(ctx), page)
}

func (r *UserRepository) ListActive(ctx context.Context, page shared.Page) ([]*user.User, error) {
	return r.findMany(ctx, r.db.WithContext(ctx).Where("is_active = ?", true), page)
}

func (r *UserRepository) findOne(_ context.Context, query *gorm.DB) (*user.User, error) {
	var record userRecord
	if err := query.Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, shared.Internal("user.query_failed", "could not load user").WithCause(err)
	}
	return record.toAggregate(), nil
}

func (r *UserRepository) findMany(_ context.Context, query *gorm.DB, page shared.Page) ([]*user.User, error) {
	var records []userRecord
	if err := query.Order("id ASC").Limit(page.Limit()).Offset(page.Offset()).Find(&records).Error; err != nil {
		return nil, shared.Internal("user.query_failed", "could not list users").WithCause(err)
	}

	accounts := make([]*user.User, 0, len(records))
	for _, record := range records {
		accounts = append(accounts, record.toAggregate())
	}
	return accounts, nil
}

func (r *UserRepository) exists(ctx context.Context, condition string, value any, excluding user.ID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&userRecord{}).Where(condition, value)
	if !excluding.IsZero() {
		query = query.Where("id <> ?", int64(excluding))
	}

	var count int64
	if err := query.Limit(1).Count(&count).Error; err != nil {
		return false, shared.Internal("user.query_failed", "could not check user uniqueness").WithCause(err)
	}
	return count > 0, nil
}
