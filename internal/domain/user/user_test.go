package user_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
	"github.com/apolzek/pastebomb/internal/testsupport/fake"
)

func fixedClock() shared.Clock {
	return shared.FixedClock{Instant: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)}
}

func validRegistration() user.Registration {
	return user.Registration{
		Name:     "Ada Lovelace",
		Username: "ada",
		Email:    "ada@example.com",
		BornDate: "1815-12-10",
		Password: "correct horse battery",
	}
}

func TestRegisterProducesAnActiveNonAdminUser(t *testing.T) {
	t.Parallel()

	account, err := user.Register(validRegistration(), fake.NewHasher(), fixedClock())
	require.NoError(t, err)

	assert.True(t, account.IsActive())
	// Admin rights must never be self-granted through the sign-up payload.
	assert.False(t, account.IsAdmin())
	assert.Equal(t, "ada", account.Username().String())
	assert.Equal(t, "ada@example.com", account.Email().String())
}

func TestRegisterNormalizesEmailAndUsername(t *testing.T) {
	t.Parallel()

	registration := validRegistration()
	registration.Email = "  ADA@Example.COM "
	registration.Username = "  AdA  "

	account, err := user.Register(registration, fake.NewHasher(), fixedClock())
	require.NoError(t, err)

	assert.Equal(t, "ada@example.com", account.Email().String())
	assert.Equal(t, "ada", account.Username().String())
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	cases := map[string]func(*user.Registration){
		"empty name":            func(r *user.Registration) { r.Name = "   " },
		"short username":        func(r *user.Registration) { r.Username = "ab" },
		"username with spaces":  func(r *user.Registration) { r.Username = "ada lovelace" },
		"malformed email":       func(r *user.Registration) { r.Email = "not-an-email" },
		"email with display":    func(r *user.Registration) { r.Email = "Ada <ada@example.com>" },
		"born date wrong shape": func(r *user.Registration) { r.BornDate = "10/12/1815" },
		"born date in future":   func(r *user.Registration) { r.BornDate = "2999-01-01" },
		"short password":        func(r *user.Registration) { r.Password = "1234567" },
		"password over bcrypt limit": func(r *user.Registration) {
			r.Password = strings.Repeat("a", 73)
		},
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			registration := validRegistration()
			mutate(&registration)

			_, err := user.Register(registration, fake.NewHasher(), fixedClock())
			require.Error(t, err)
			assert.Equal(t, shared.KindInvalidInput, shared.KindOf(err))
		})
	}
}

func TestAuthenticateAcceptsTheCorrectPassword(t *testing.T) {
	t.Parallel()

	account, err := user.Register(validRegistration(), fake.NewHasher(), fixedClock())
	require.NoError(t, err)

	assert.NoError(t, account.Authenticate("correct horse battery", fake.NewHasher()))
}

func TestAuthenticateRejectsWrongPasswordAndInactiveAccount(t *testing.T) {
	t.Parallel()

	hasher := fake.NewHasher()
	clock := fixedClock()

	account, err := user.Register(validRegistration(), hasher, clock)
	require.NoError(t, err)

	assert.ErrorIs(t, account.Authenticate("wrong password", hasher), user.ErrInvalidCredentials)

	account.Deactivate(clock)
	// A deactivated account must fail with the same error as a wrong password,
	// so the response cannot be used to tell the two apart.
	assert.ErrorIs(t, account.Authenticate("correct horse battery", hasher), user.ErrInvalidCredentials)
}

func TestDeactivateIsIdempotent(t *testing.T) {
	t.Parallel()

	clock := fixedClock()
	account, err := user.Register(validRegistration(), fake.NewHasher(), clock)
	require.NoError(t, err)

	account.Deactivate(clock)
	updatedAt := account.UpdatedAt()
	account.Deactivate(clock)

	assert.False(t, account.IsActive())
	assert.Equal(t, updatedAt, account.UpdatedAt())
}

func TestUpdateProfileValidatesEveryField(t *testing.T) {
	t.Parallel()

	clock := fixedClock()
	account, err := user.Register(validRegistration(), fake.NewHasher(), clock)
	require.NoError(t, err)

	err = account.UpdateProfile(user.ProfileUpdate{
		Name:     "Ada King",
		Username: "ada.king",
		Email:    "ada.king@example.com",
		BornDate: "1815-12-10",
	}, clock)
	require.NoError(t, err)
	assert.Equal(t, "ada.king", account.Username().String())

	err = account.UpdateProfile(user.ProfileUpdate{
		Name:     "Ada King",
		Username: "ada.king",
		Email:    "broken",
		BornDate: "1815-12-10",
	}, clock)
	require.Error(t, err)
	// A rejected update must leave the aggregate exactly as it was.
	assert.Equal(t, "ada.king@example.com", account.Email().String())
}

func TestRehydrateRestoresStoredStateWithoutRevalidating(t *testing.T) {
	t.Parallel()

	// A username that today's rules would reject must still load, otherwise a
	// tightened rule would lock existing users out of their own accounts.
	account := user.Rehydrate(user.Snapshot{
		ID:       7,
		Username: "legacy user!",
		Email:    "legacy@example.com",
		Active:   true,
		Admin:    true,
	})

	assert.Equal(t, user.ID(7), account.ID())
	assert.Equal(t, "legacy user!", account.Username().String())
	assert.True(t, account.IsAdmin())
}

func TestAssignIDOnlyAppliesOnce(t *testing.T) {
	t.Parallel()

	account, err := user.Register(validRegistration(), fake.NewHasher(), fixedClock())
	require.NoError(t, err)

	account.AssignID(3)
	account.AssignID(9)
	assert.Equal(t, user.ID(3), account.ID())
}
