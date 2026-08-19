package user

import (
	"context"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// Repository is the persistence port for the User aggregate. Lookups return
// ErrNotFound rather than a nil aggregate, so a missing row can never be
// mistaken for an empty one.
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error

	ByID(ctx context.Context, id ID) (*User, error)
	ByEmail(ctx context.Context, email Email) (*User, error)
	// ByHandle resolves either a numeric id or a username, which is what the
	// public profile route accepts.
	ByHandle(ctx context.Context, handle string) (*User, error)

	ExistsByEmail(ctx context.Context, email Email, excluding ID) (bool, error)
	ExistsByUsername(ctx context.Context, username Username, excluding ID) (bool, error)

	List(ctx context.Context, page shared.Page) ([]*User, error)
	ListActive(ctx context.Context, page shared.Page) ([]*User, error)
}
