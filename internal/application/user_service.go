package application

import (
	"context"

	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
)

// RegisterCommand carries a sign-up request.
type RegisterCommand struct {
	Name     string
	Username string
	Email    string
	BornDate string
	Password string
}

// UpdateProfileCommand carries a full profile replacement.
type UpdateProfileCommand struct {
	Name     string
	Username string
	Email    string
	BornDate string
}

// UserService implements the user-facing use cases.
type UserService struct {
	users  user.Repository
	hasher shared.Hasher
	clock  shared.Clock
	audit  AuditRecorder
}

func NewUserService(users user.Repository, hasher shared.Hasher, clock shared.Clock, audit AuditRecorder) *UserService {
	return &UserService{users: users, hasher: hasher, clock: clock, audit: audit}
}

// Register creates an account. Uniqueness is checked up front for a clear
// error message, and the unique indexes in the schema are what actually
// prevent a duplicate under a concurrent sign-up.
func (s *UserService) Register(ctx context.Context, cmd RegisterCommand) (*user.User, error) {
	account, err := user.Register(user.Registration(cmd), s.hasher, s.clock)
	if err != nil {
		return nil, err
	}

	if err := s.assertHandlesFree(ctx, account.Email(), account.Username(), 0); err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, account); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, AuditEvent{
		Level:   AuditInfo,
		Actor:   account.Username().String(),
		Action:  "user.registered",
		Details: map[string]any{"user_id": int64(account.ID())},
	})
	return account, nil
}

// Profile returns the full record of the authenticated user.
func (s *UserService) Profile(ctx context.Context, id user.ID) (*user.User, error) {
	return s.users.ByID(ctx, id)
}

// PublicProfile resolves a user by numeric id or username. The caller is
// responsible for rendering only the fields that are safe to publish.
func (s *UserService) PublicProfile(ctx context.Context, handle string) (*user.User, error) {
	account, err := s.users.ByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	if !account.IsActive() {
		return nil, user.ErrNotFound
	}
	return account, nil
}

// UpdateProfile replaces the mutable profile fields of the authenticated user.
func (s *UserService) UpdateProfile(ctx context.Context, id user.ID, cmd UpdateProfileCommand) (*user.User, error) {
	account, err := s.users.ByID(ctx, id)
	if err != nil {
		return nil, err
	}

	candidateEmail, err := user.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}
	candidateUsername, err := user.NewUsername(cmd.Username)
	if err != nil {
		return nil, err
	}

	var emailToCheck user.Email
	if account.EmailChanged(candidateEmail) {
		emailToCheck = candidateEmail
	}
	var usernameToCheck user.Username
	if account.UsernameChanged(candidateUsername) {
		usernameToCheck = candidateUsername
	}
	if err := s.assertHandlesFree(ctx, emailToCheck, usernameToCheck, account.ID()); err != nil {
		return nil, err
	}

	if err := account.UpdateProfile(user.ProfileUpdate(cmd), s.clock); err != nil {
		return nil, err
	}
	if err := s.users.Update(ctx, account); err != nil {
		return nil, err
	}

	s.audit.Record(ctx, AuditEvent{
		Level:   AuditInfo,
		Actor:   account.Username().String(),
		Action:  "user.profile_updated",
		Details: map[string]any{"user_id": int64(account.ID())},
	})
	return account, nil
}

// Deactivate soft-deletes an account. A user may only deactivate their own
// account unless the caller holds admin rights.
func (s *UserService) Deactivate(ctx context.Context, actor Identity, target user.ID) error {
	if actor.IsAnonymous() {
		return user.ErrInvalidCredentials
	}
	if actor.UserID != target && !actor.Admin {
		return user.ErrForbidden
	}

	account, err := s.users.ByID(ctx, target)
	if err != nil {
		return err
	}

	account.Deactivate(s.clock)
	if err := s.users.Update(ctx, account); err != nil {
		return err
	}

	s.audit.Record(ctx, AuditEvent{
		Level:   AuditWarning,
		Actor:   actor.Email,
		Action:  "user.deactivated",
		Details: map[string]any{"user_id": int64(target)},
	})
	return nil
}

// List returns every account, active or not. Reserved for administrators.
func (s *UserService) List(ctx context.Context, page shared.Page) ([]*user.User, error) {
	return s.users.List(ctx, page)
}

// ListActive returns only accounts that have not been deactivated.
func (s *UserService) ListActive(ctx context.Context, page shared.Page) ([]*user.User, error) {
	return s.users.ListActive(ctx, page)
}

// assertHandlesFree checks the unique fields that were actually supplied. The
// zero value of either argument means "unchanged, nothing to verify".
func (s *UserService) assertHandlesFree(ctx context.Context, email user.Email, username user.Username, excluding user.ID) error {
	if email != "" {
		taken, err := s.users.ExistsByEmail(ctx, email, excluding)
		if err != nil {
			return err
		}
		if taken {
			return user.ErrEmailTaken
		}
	}
	if username != "" {
		taken, err := s.users.ExistsByUsername(ctx, username, excluding)
		if err != nil {
			return err
		}
		if taken {
			return user.ErrUsernameTaken
		}
	}
	return nil
}
