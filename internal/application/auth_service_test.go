package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/user"
	"github.com/apolzek/pastebomb/internal/testsupport/fake"
	"github.com/apolzek/pastebomb/internal/testsupport/memory"
)

type authFixture struct {
	auth  *application.AuthService
	users *memory.UserRepository
	audit *memory.AuditRecorder
}

func newAuthFixture(t *testing.T) authFixture {
	t.Helper()

	users := memory.NewUserRepository()
	recorder := memory.NewAuditRecorder()
	hasher := fake.NewHasher()

	userService := application.NewUserService(users, hasher, testClock, recorder)
	_, err := userService.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	return authFixture{
		auth:  application.NewAuthService(users, hasher, fake.NewTokenIssuer(), recorder),
		users: users,
		audit: recorder,
	}
}

func TestLoginIssuesATokenForValidCredentials(t *testing.T) {
	t.Parallel()

	fixture := newAuthFixture(t)

	token, err := fixture.auth.Login(context.Background(), application.LoginCommand{
		Email:    "ada@example.com",
		Password: "correct horse battery",
	})
	require.NoError(t, err)

	assert.NotEmpty(t, token.Value)
	assert.False(t, token.ExpiresAt.IsZero())
	assert.Contains(t, fixture.audit.Actions(), "login.succeeded")
}

// The three failure modes must be indistinguishable to the caller, otherwise
// the endpoint becomes an oracle for which addresses are registered.
func TestLoginReturnsTheSameErrorForEveryFailureMode(t *testing.T) {
	t.Parallel()

	cases := map[string]application.LoginCommand{
		"unknown account": {Email: "nobody@example.com", Password: "correct horse battery"},
		"wrong password":  {Email: "ada@example.com", Password: "wrong password"},
		"malformed email": {Email: "not-an-email", Password: "correct horse battery"},
		"empty password":  {Email: "ada@example.com", Password: ""},
	}

	for name, command := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fixture := newAuthFixture(t)

			_, err := fixture.auth.Login(context.Background(), command)
			assert.ErrorIs(t, err, user.ErrInvalidCredentials)
			assert.Contains(t, fixture.audit.Actions(), "login.failed")
		})
	}
}

func TestLoginRejectsADeactivatedAccount(t *testing.T) {
	t.Parallel()

	fixture := newAuthFixture(t)
	users := application.NewUserService(fixture.users, fake.NewHasher(), testClock, fixture.audit)

	account, err := users.PublicProfile(context.Background(), "ada")
	require.NoError(t, err)
	require.NoError(t, users.Deactivate(context.Background(), application.Identity{UserID: account.ID()}, account.ID()))

	_, err = fixture.auth.Login(context.Background(), application.LoginCommand{
		Email:    "ada@example.com",
		Password: "correct horse battery",
	})
	assert.ErrorIs(t, err, user.ErrInvalidCredentials)
}

func TestLoginCarriesTheAdminClaimIntoTheToken(t *testing.T) {
	t.Parallel()

	fixture := newAuthFixture(t)
	issuer := fake.NewTokenIssuer()
	auth := application.NewAuthService(fixture.users, fake.NewHasher(), issuer, fixture.audit)

	account, err := fixture.users.ByEmail(context.Background(), "ada@example.com")
	require.NoError(t, err)
	fixture.users.PromoteToAdmin(account.ID())

	_, err = auth.Login(context.Background(), application.LoginCommand{
		Email:    "ada@example.com",
		Password: "correct horse battery",
	})
	require.NoError(t, err)

	assert.True(t, issuer.LastIdentity().Admin)
}
