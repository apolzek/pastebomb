package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/testsupport/fake"
	"github.com/apolzek/pastebomb/internal/testsupport/memory"
)

func newPasteService(clock shared.Clock, slugs paste.SlugGenerator) (*application.PasteService, *memory.PasteRepository) {
	hasher := fake.NewHasher()
	repository := memory.NewPasteRepository()

	service := application.NewPasteService(
		repository,
		paste.NewFactory(slugs, hasher, clock),
		paste.NewAccessGuard(hasher, clock),
		clock,
		memory.NewAuditRecorder(),
	)
	return service, repository
}

func createCommand() application.CreatePasteCommand {
	return application.CreatePasteCommand{
		Title:    "deploy notes",
		Content:  "kubectl rollout status deploy/api",
		IsPublic: true,
	}
}

func TestCreateAndReadRoundTrip(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.NewSlugGenerator("s"))

	created, err := service.Create(context.Background(), createCommand())
	require.NoError(t, err)

	found, err := service.Read(context.Background(), created.Slug().String(), "")
	require.NoError(t, err)
	assert.Equal(t, "kubectl rollout status deploy/api", found.Content().String())
}

func TestReadRequiresTheSecretWhenOneIsSet(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.NewSlugGenerator("s"))

	command := createCommand()
	command.Secret = "open sesame"
	created, err := service.Create(context.Background(), command)
	require.NoError(t, err)

	slug := created.Slug().String()

	_, err = service.Read(context.Background(), slug, "")
	assert.ErrorIs(t, err, paste.ErrSecretRequired)

	_, err = service.Read(context.Background(), slug, "guess")
	assert.ErrorIs(t, err, paste.ErrSecretRequired)

	found, err := service.Read(context.Background(), slug, "open sesame")
	require.NoError(t, err)
	assert.Equal(t, slug, found.Slug().String())
}

func TestReadReportsAnExpiredPasteAsNotFound(t *testing.T) {
	t.Parallel()

	service, repository := newPasteService(testClock, fake.NewSlugGenerator("s"))

	command := createCommand()
	command.ExpiresAt = testClock.Instant.Add(time.Hour).Format(time.RFC3339)
	created, err := service.Create(context.Background(), command)
	require.NoError(t, err)

	// The row is still in the table: reads must hide it without waiting for
	// the purge job to catch up.
	later := shared.FixedClock{Instant: testClock.Instant.Add(2 * time.Hour)}
	lateService := application.NewPasteService(
		repository,
		paste.NewFactory(fake.NewSlugGenerator("s"), fake.NewHasher(), later),
		paste.NewAccessGuard(fake.NewHasher(), later),
		later,
		memory.NewAuditRecorder(),
	)

	_, err = lateService.Read(context.Background(), created.Slug().String(), "")
	assert.ErrorIs(t, err, paste.ErrNotFound)
}

func TestReadRejectsAnUnknownOrMalformedHandle(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.NewSlugGenerator("s"))

	for _, handle := range []string{"", "   ", "does-not-exist"} {
		_, err := service.Read(context.Background(), handle, "")
		assert.ErrorIs(t, err, paste.ErrNotFound)
	}
}

// Reads are keyed on the random slug only. A sequential id must never resolve,
// or every unlisted paste becomes enumerable by counting.
func TestReadNeverResolvesANumericID(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.NewSlugGenerator("s"))

	created, err := service.Create(context.Background(), createCommand())
	require.NoError(t, err)
	require.Equal(t, paste.ID(1), created.ID())

	_, err = service.Read(context.Background(), "1", "")
	assert.ErrorIs(t, err, paste.ErrNotFound)
}

func TestListPublicExcludesUnlistedAndExpiredPastes(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.NewSlugGenerator("s"))

	public := createCommand()
	_, err := service.Create(context.Background(), public)
	require.NoError(t, err)

	unlisted := createCommand()
	unlisted.IsPublic = false
	unlisted.Owner = 7
	_, err = service.Create(context.Background(), unlisted)
	require.NoError(t, err)

	expired := createCommand()
	expired.ExpiresAt = testClock.Instant.Add(-time.Hour).Format(time.RFC3339)
	_, err = service.Create(context.Background(), expired)
	// A past deadline is rejected at creation, so it cannot enter the index.
	assert.Equal(t, shared.KindInvalidInput, shared.KindOf(err))

	listed, err := service.ListPublic(context.Background(), shared.NewPage(1, 10))
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.True(t, listed[0].Visibility().IsPublic())
}

func TestListOwnedReturnsOnlyTheCallersPastes(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.NewSlugGenerator("s"))

	mine := createCommand()
	mine.Owner = 7
	_, err := service.Create(context.Background(), mine)
	require.NoError(t, err)

	theirs := createCommand()
	theirs.Owner = 8
	_, err = service.Create(context.Background(), theirs)
	require.NoError(t, err)

	listed, err := service.ListOwned(context.Background(), 7, shared.NewPage(1, 10))
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.True(t, listed[0].BelongsTo(7))
}

func TestListOwnedRefusesAnAnonymousOwner(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.NewSlugGenerator("s"))

	_, err := service.ListOwned(context.Background(), 0, shared.NewPage(1, 10))
	assert.ErrorIs(t, err, paste.ErrForbidden)
}

// A slug generator that keeps repeating itself must surface as an error rather
// than as a second paste overwriting the first.
func TestCreateGivesUpAfterRepeatedSlugCollisions(t *testing.T) {
	t.Parallel()

	service, _ := newPasteService(testClock, fake.StaticSlugGenerator{Value: "always-the-same"})

	_, err := service.Create(context.Background(), createCommand())
	require.NoError(t, err)

	_, err = service.Create(context.Background(), createCommand())
	require.Error(t, err)
	assert.Equal(t, shared.KindInternal, shared.KindOf(err))
}

func TestPurgeExpiredRemovesOnlyPastDeadlines(t *testing.T) {
	t.Parallel()

	service, repository := newPasteService(testClock, fake.NewSlugGenerator("s"))

	expiring := createCommand()
	expiring.ExpiresAt = testClock.Instant.Add(time.Hour).Format(time.RFC3339)
	_, err := service.Create(context.Background(), expiring)
	require.NoError(t, err)

	_, err = service.Create(context.Background(), createCommand())
	require.NoError(t, err)

	deleted, err := service.PurgeExpired(context.Background())
	require.NoError(t, err)
	assert.Zero(t, deleted)

	later := shared.FixedClock{Instant: testClock.Instant.Add(2 * time.Hour)}
	lateService := application.NewPasteService(
		repository,
		paste.NewFactory(fake.NewSlugGenerator("s"), fake.NewHasher(), later),
		paste.NewAccessGuard(fake.NewHasher(), later),
		later,
		memory.NewAuditRecorder(),
	)

	deleted, err = lateService.PurgeExpired(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted)

	remaining, err := lateService.ListPublic(context.Background(), shared.NewPage(1, 10))
	require.NoError(t, err)
	assert.Len(t, remaining, 1)
}
