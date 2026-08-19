// Package memory holds in-memory implementations of the domain repository
// ports. They let the use cases and the HTTP layer be tested end to end
// without a database, which is what keeps `go test ./...` runnable on a fresh
// clone with no docker compose running.
package memory

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
)

// UserRepository stores users in a map, preserving insertion order for stable
// pagination.
type UserRepository struct {
	mu       sync.RWMutex
	byID     map[user.ID]user.Snapshot
	order    []user.ID
	sequence int64
}

func NewUserRepository() *UserRepository {
	return &UserRepository{byID: make(map[user.ID]user.Snapshot)}
}

var _ user.Repository = (*UserRepository)(nil)

func (r *UserRepository) Create(_ context.Context, account *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	snapshot := account.ToSnapshot()
	for _, existing := range r.byID {
		if existing.Email == snapshot.Email {
			return user.ErrEmailTaken
		}
		if existing.Username == snapshot.Username {
			return user.ErrUsernameTaken
		}
	}

	r.sequence++
	snapshot.ID = r.sequence
	id := user.ID(snapshot.ID)

	r.byID[id] = snapshot
	r.order = append(r.order, id)
	account.AssignID(id)
	return nil
}

func (r *UserRepository) Update(_ context.Context, account *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	snapshot := account.ToSnapshot()
	if _, exists := r.byID[user.ID(snapshot.ID)]; !exists {
		return user.ErrNotFound
	}
	r.byID[user.ID(snapshot.ID)] = snapshot
	return nil
}

func (r *UserRepository) ByID(_ context.Context, id user.ID) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot, exists := r.byID[id]
	if !exists {
		return nil, user.ErrNotFound
	}
	return user.Rehydrate(snapshot), nil
}

func (r *UserRepository) ByEmail(_ context.Context, email user.Email) (*user.User, error) {
	return r.first(func(s user.Snapshot) bool { return s.Email == email.String() })
}

func (r *UserRepository) ByHandle(_ context.Context, handle string) (*user.User, error) {
	if id, err := strconv.ParseInt(handle, 10, 64); err == nil {
		return r.first(func(s user.Snapshot) bool { return s.ID == id })
	}
	lowered := strings.ToLower(handle)
	return r.first(func(s user.Snapshot) bool { return s.Username == lowered })
}

func (r *UserRepository) ExistsByEmail(_ context.Context, email user.Email, excluding user.ID) (bool, error) {
	return r.exists(func(s user.Snapshot) bool {
		return s.Email == email.String() && user.ID(s.ID) != excluding
	}), nil
}

func (r *UserRepository) ExistsByUsername(_ context.Context, username user.Username, excluding user.ID) (bool, error) {
	return r.exists(func(s user.Snapshot) bool {
		return s.Username == username.String() && user.ID(s.ID) != excluding
	}), nil
}

func (r *UserRepository) List(_ context.Context, page shared.Page) ([]*user.User, error) {
	return r.list(page, func(user.Snapshot) bool { return true }), nil
}

func (r *UserRepository) ListActive(_ context.Context, page shared.Page) ([]*user.User, error) {
	return r.list(page, func(s user.Snapshot) bool { return s.Active }), nil
}

// PromoteToAdmin grants administrative rights, mirroring the out-of-band SQL
// update an operator would run in production.
func (r *UserRepository) PromoteToAdmin(id user.ID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if snapshot, exists := r.byID[id]; exists {
		snapshot.Admin = true
		r.byID[id] = snapshot
	}
}

func (r *UserRepository) first(match func(user.Snapshot) bool) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, id := range r.order {
		if snapshot := r.byID[id]; match(snapshot) {
			return user.Rehydrate(snapshot), nil
		}
	}
	return nil, user.ErrNotFound
}

func (r *UserRepository) exists(match func(user.Snapshot) bool) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, id := range r.order {
		if match(r.byID[id]) {
			return true
		}
	}
	return false
}

func (r *UserRepository) list(page shared.Page, match func(user.Snapshot) bool) []*user.User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*user.User
	for _, id := range r.order {
		if snapshot := r.byID[id]; match(snapshot) {
			matched = append(matched, user.Rehydrate(snapshot))
		}
	}
	return paginate(matched, page)
}
