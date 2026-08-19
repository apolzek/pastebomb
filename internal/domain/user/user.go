// Package user is the User aggregate: identity, profile data, credentials and
// the rules that keep them consistent. It knows nothing about HTTP or SQL.
package user

import (
	"time"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// User is the aggregate root. Every field is private so the only way to reach
// an invalid state is through a method that refuses to produce one.
type User struct {
	id           ID
	name         Name
	username     Username
	email        Email
	bornDate     BornDate
	passwordHash PasswordHash
	active       bool
	admin        bool
	createdAt    time.Time
	updatedAt    time.Time
}

// Registration carries the raw input of a sign-up. Validation happens inside
// Register, so callers cannot skip it.
type Registration struct {
	Name     string
	Username string
	Email    string
	BornDate string
	Password string
}

// Register builds a new active, non-admin user with all invariants enforced.
// Admin rights are never self-granted: they are set out of band.
func Register(reg Registration, hasher shared.Hasher, clock shared.Clock) (*User, error) {
	now := clock.Now()

	name, err := NewName(reg.Name)
	if err != nil {
		return nil, err
	}
	username, err := NewUsername(reg.Username)
	if err != nil {
		return nil, err
	}
	email, err := NewEmail(reg.Email)
	if err != nil {
		return nil, err
	}
	bornDate, err := NewBornDate(reg.BornDate, now)
	if err != nil {
		return nil, err
	}
	passwordHash, err := NewPasswordHash(reg.Password, hasher)
	if err != nil {
		return nil, err
	}

	return &User{
		name:         name,
		username:     username,
		email:        email,
		bornDate:     bornDate,
		passwordHash: passwordHash,
		active:       true,
		admin:        false,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// Snapshot is the flat view of a User used to cross the persistence boundary
// in both directions. Repositories map rows to it; Rehydrate turns it back
// into an aggregate without re-running creation invariants.
type Snapshot struct {
	ID           int64
	Name         string
	Username     string
	Email        string
	BornDate     time.Time
	PasswordHash string
	Active       bool
	Admin        bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Rehydrate reconstructs a persisted aggregate. It is intentionally total: a
// row that is already stored must always load, even if today's rules are
// stricter than the ones in force when it was written.
func Rehydrate(snapshot Snapshot) *User {
	return &User{
		id:           ID(snapshot.ID),
		name:         Name(snapshot.Name),
		username:     Username(snapshot.Username),
		email:        Email(snapshot.Email),
		bornDate:     RehydrateBornDate(snapshot.BornDate),
		passwordHash: PasswordHash(snapshot.PasswordHash),
		active:       snapshot.Active,
		admin:        snapshot.Admin,
		createdAt:    snapshot.CreatedAt,
		updatedAt:    snapshot.UpdatedAt,
	}
}

// ToSnapshot flattens the aggregate for storage.
func (u *User) ToSnapshot() Snapshot {
	return Snapshot{
		ID:           int64(u.id),
		Name:         u.name.String(),
		Username:     u.username.String(),
		Email:        u.email.String(),
		BornDate:     u.bornDate.Time(),
		PasswordHash: u.passwordHash.String(),
		Active:       u.active,
		Admin:        u.admin,
		CreatedAt:    u.createdAt,
		UpdatedAt:    u.updatedAt,
	}
}

func (u *User) ID() ID               { return u.id }
func (u *User) Name() Name           { return u.name }
func (u *User) Username() Username   { return u.username }
func (u *User) Email() Email         { return u.email }
func (u *User) BornDate() BornDate   { return u.bornDate }
func (u *User) IsActive() bool       { return u.active }
func (u *User) IsAdmin() bool        { return u.admin }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// AssignID stamps the identity produced by the store. It is only valid once.
func (u *User) AssignID(id ID) {
	if u.id.IsZero() {
		u.id = id
	}
}

// Authenticate verifies a login attempt. It returns the same error for an
// inactive account and a wrong password so the response cannot be used to
// probe which accounts exist.
func (u *User) Authenticate(plainPassword string, hasher shared.Hasher) error {
	if !u.active || !hasher.Matches(u.passwordHash.String(), plainPassword) {
		return ErrInvalidCredentials
	}
	return nil
}

// ProfileUpdate is the subset of the profile a user may change. Password,
// active flag and admin flag deliberately have their own operations.
type ProfileUpdate struct {
	Name     string
	Username string
	Email    string
	BornDate string
}

// UpdateProfile applies a full profile replacement, validating every field.
func (u *User) UpdateProfile(update ProfileUpdate, clock shared.Clock) error {
	now := clock.Now()

	name, err := NewName(update.Name)
	if err != nil {
		return err
	}
	username, err := NewUsername(update.Username)
	if err != nil {
		return err
	}
	email, err := NewEmail(update.Email)
	if err != nil {
		return err
	}
	bornDate, err := NewBornDate(update.BornDate, now)
	if err != nil {
		return err
	}

	u.name, u.username, u.email, u.bornDate = name, username, email, bornDate
	u.updatedAt = now
	return nil
}

// ChangePassword replaces the credential after re-validating the policy.
func (u *User) ChangePassword(plainPassword string, hasher shared.Hasher, clock shared.Clock) error {
	hash, err := NewPasswordHash(plainPassword, hasher)
	if err != nil {
		return err
	}
	u.passwordHash = hash
	u.updatedAt = clock.Now()
	return nil
}

// Deactivate performs the soft delete. Repeating it is a no-op rather than an
// error, which keeps the operation idempotent for retrying clients.
func (u *User) Deactivate(clock shared.Clock) {
	if !u.active {
		return
	}
	u.active = false
	u.updatedAt = clock.Now()
}

// EmailChanged reports whether candidate differs from the current address,
// letting the application layer skip a uniqueness query when nothing moved.
func (u *User) EmailChanged(candidate Email) bool { return u.email != candidate }

// UsernameChanged reports whether candidate differs from the current handle.
func (u *User) UsernameChanged(candidate Username) bool { return u.username != candidate }
