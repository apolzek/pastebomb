package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
	"github.com/apolzek/pastebomb/internal/testsupport/fake"
	"github.com/apolzek/pastebomb/internal/testsupport/memory"
)

var testClock = shared.FixedClock{Instant: time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)}

type userFixture struct {
	service *application.UserService
	users   *memory.UserRepository
	audit   *memory.AuditRecorder
}

func newUserFixture() userFixture {
	users := memory.NewUserRepository()
	recorder := memory.NewAuditRecorder()
	return userFixture{
		service: application.NewUserService(users, fake.NewHasher(), testClock, recorder),
		users:   users,
		audit:   recorder,
	}
}

func registerCommand() application.RegisterCommand {
	return application.RegisterCommand{
		Name:     "Ada Lovelace",
		Username: "ada",
		Email:    "ada@example.com",
		BornDate: "1815-12-10",
		Password: "correct horse battery",
	}
}

func TestRegisterAssignsAnIDAndAuditsTheEvent(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()

	account, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	assert.False(t, account.ID().IsZero())
	assert.Contains(t, fixture.audit.Actions(), "user.registered")
}

func TestRegisterRejectsADuplicateEmailOrUsername(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	_, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	duplicateEmail := registerCommand()
	duplicateEmail.Username = "ada2"
	_, err = fixture.service.Register(context.Background(), duplicateEmail)
	assert.ErrorIs(t, err, user.ErrEmailTaken)

	duplicateUsername := registerCommand()
	duplicateUsername.Email = "other@example.com"
	_, err = fixture.service.Register(context.Background(), duplicateUsername)
	assert.ErrorIs(t, err, user.ErrUsernameTaken)
}

// Case and surrounding whitespace must not open a second account on the same
// address, which is what the normalization in the value object is for.
func TestRegisterTreatsEmailCaseInsensitively(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	_, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	shouted := registerCommand()
	shouted.Username = "ada2"
	shouted.Email = "ADA@EXAMPLE.COM"

	_, err = fixture.service.Register(context.Background(), shouted)
	assert.ErrorIs(t, err, user.ErrEmailTaken)
}

func TestUpdateProfileAllowsKeepingYourOwnEmail(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	account, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	updated, err := fixture.service.UpdateProfile(context.Background(), account.ID(), application.UpdateProfileCommand{
		Name:     "Ada King",
		Username: "ada",
		Email:    "ada@example.com",
		BornDate: "1815-12-10",
	})
	require.NoError(t, err)
	assert.Equal(t, "Ada King", updated.Name().String())
}

func TestUpdateProfileRejectsAnEmailOwnedBySomeoneElse(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	first, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	second := registerCommand()
	second.Username = "grace"
	second.Email = "grace@example.com"
	_, err = fixture.service.Register(context.Background(), second)
	require.NoError(t, err)

	_, err = fixture.service.UpdateProfile(context.Background(), first.ID(), application.UpdateProfileCommand{
		Name:     "Ada",
		Username: "ada",
		Email:    "grace@example.com",
		BornDate: "1815-12-10",
	})
	assert.ErrorIs(t, err, user.ErrEmailTaken)
}

func TestDeactivateRefusesToTouchAnotherAccount(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	victim, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	attacker := application.Identity{UserID: victim.ID() + 1, Email: "eve@example.com"}

	err = fixture.service.Deactivate(context.Background(), attacker, victim.ID())
	assert.ErrorIs(t, err, user.ErrForbidden)

	stillThere, err := fixture.service.Profile(context.Background(), victim.ID())
	require.NoError(t, err)
	assert.True(t, stillThere.IsActive())
}

func TestDeactivateAllowsTheOwnerAndAnAdmin(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	account, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	owner := application.Identity{UserID: account.ID()}
	require.NoError(t, fixture.service.Deactivate(context.Background(), owner, account.ID()))

	admin := application.Identity{UserID: account.ID() + 99, Admin: true}
	assert.NoError(t, fixture.service.Deactivate(context.Background(), admin, account.ID()))
}

func TestDeactivateRejectsAnAnonymousCaller(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	account, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	err = fixture.service.Deactivate(context.Background(), application.Identity{}, account.ID())
	assert.ErrorIs(t, err, user.ErrInvalidCredentials)
}

func TestPublicProfileHidesDeactivatedAccounts(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	account, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	found, err := fixture.service.PublicProfile(context.Background(), "ada")
	require.NoError(t, err)
	assert.Equal(t, account.ID(), found.ID())

	require.NoError(t, fixture.service.Deactivate(context.Background(), application.Identity{UserID: account.ID()}, account.ID()))

	_, err = fixture.service.PublicProfile(context.Background(), "ada")
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestListActiveExcludesDeactivatedAccounts(t *testing.T) {
	t.Parallel()

	fixture := newUserFixture()
	first, err := fixture.service.Register(context.Background(), registerCommand())
	require.NoError(t, err)

	second := registerCommand()
	second.Username = "grace"
	second.Email = "grace@example.com"
	_, err = fixture.service.Register(context.Background(), second)
	require.NoError(t, err)

	require.NoError(t, fixture.service.Deactivate(context.Background(), application.Identity{UserID: first.ID()}, first.ID()))

	active, err := fixture.service.ListActive(context.Background(), shared.NewPage(1, 10))
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, "grace", active[0].Username().String())

	all, err := fixture.service.List(context.Background(), shared.NewPage(1, 10))
	require.NoError(t, err)
	assert.Len(t, all, 2)
}
