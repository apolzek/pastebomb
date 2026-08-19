package paste_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/testsupport/fake"
)

var now = time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

func newFactory() *paste.Factory {
	return paste.NewFactory(fake.NewSlugGenerator("slug"), fake.NewHasher(), shared.FixedClock{Instant: now})
}

func validDraft() paste.Draft {
	return paste.Draft{
		Title:    "deploy notes",
		Content:  "kubectl rollout status deploy/api",
		Category: "ops",
		IsPublic: true,
	}
}

func TestCreateProducesAnUnexpiredPublicPaste(t *testing.T) {
	t.Parallel()

	document, err := newFactory().Create(validDraft())
	require.NoError(t, err)

	assert.Equal(t, "slug-1", document.Slug().String())
	assert.True(t, document.Visibility().IsPublic())
	assert.False(t, document.HasSecret())
	assert.False(t, document.IsExpired(now))
	assert.True(t, document.Owner().IsAnonymous())
}

func TestCreateRejectsInvalidDrafts(t *testing.T) {
	t.Parallel()

	cases := map[string]func(*paste.Draft){
		"empty title":            func(d *paste.Draft) { d.Title = "  " },
		"empty content":          func(d *paste.Draft) { d.Content = "" },
		"content over the cap":   func(d *paste.Draft) { d.Content = strings.Repeat("x", paste.MaxContentBytes+1) },
		"invalid utf8 content":   func(d *paste.Draft) { d.Content = string([]byte{0xff, 0xfe}) },
		"expiration not rfc3339": func(d *paste.Draft) { d.ExpiresAt = "tomorrow" },
		"expiration in the past": func(d *paste.Draft) { d.ExpiresAt = "2020-01-01T00:00:00Z" },
		"expiration too distant": func(d *paste.Draft) { d.ExpiresAt = "2999-01-01T00:00:00Z" },
		"secret over the cap":    func(d *paste.Draft) { d.Secret = strings.Repeat("s", 73) },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			mutate(&draft)

			_, err := newFactory().Create(draft)
			require.Error(t, err)
			assert.Equal(t, shared.KindInvalidInput, shared.KindOf(err))
		})
	}
}

func TestCreateStoresTheSecretHashedNeverInPlaintext(t *testing.T) {
	t.Parallel()

	draft := validDraft()
	draft.Secret = "open sesame"

	document, err := newFactory().Create(draft)
	require.NoError(t, err)

	snapshot := document.ToSnapshot()
	assert.True(t, document.HasSecret())
	assert.True(t, fake.IsHashed(snapshot.SecretHash))
	// The stored value must be the hasher's output, never the plaintext. That
	// the fake is reversible is irrelevant here; bcrypt is covered separately.
	assert.NotEqual(t, "open sesame", snapshot.SecretHash)
}

func TestExpirationIsInclusiveOfTheDeadline(t *testing.T) {
	t.Parallel()

	draft := validDraft()
	draft.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)

	document, err := newFactory().Create(draft)
	require.NoError(t, err)

	assert.False(t, document.IsExpired(now.Add(59*time.Minute)))
	// The deadline itself already counts as expired.
	assert.True(t, document.IsExpired(now.Add(time.Hour)))
	assert.True(t, document.IsExpired(now.Add(2*time.Hour)))
}

func TestUnlockRequiresTheCorrectSecret(t *testing.T) {
	t.Parallel()

	hasher := fake.NewHasher()
	draft := validDraft()
	draft.Secret = "open sesame"

	document, err := newFactory().Create(draft)
	require.NoError(t, err)

	assert.ErrorIs(t, document.Unlock("", hasher), paste.ErrSecretRequired)
	assert.ErrorIs(t, document.Unlock("wrong", hasher), paste.ErrSecretRequired)
	assert.NoError(t, document.Unlock("open sesame", hasher))
}

func TestUnlockIsOpenWhenNoSecretWasSet(t *testing.T) {
	t.Parallel()

	document, err := newFactory().Create(validDraft())
	require.NoError(t, err)

	assert.NoError(t, document.Unlock("", fake.NewHasher()))
}

func TestAccessGuardReportsExpiredPastesAsNotFound(t *testing.T) {
	t.Parallel()

	clock := shared.FixedClock{Instant: now}
	guard := paste.NewAccessGuard(fake.NewHasher(), clock)

	draft := validDraft()
	draft.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)
	document, err := newFactory().Create(draft)
	require.NoError(t, err)

	assert.NoError(t, guard.Authorize(document, ""))

	// Once past the deadline the paste must look like it never existed, rather
	// than confirming a valid handle with a 410 or a 403.
	lateGuard := paste.NewAccessGuard(fake.NewHasher(), shared.FixedClock{Instant: now.Add(2 * time.Hour)})
	assert.ErrorIs(t, lateGuard.Authorize(document, ""), paste.ErrNotFound)
}

func TestAccessGuardTreatsNilAsNotFound(t *testing.T) {
	t.Parallel()

	guard := paste.NewAccessGuard(fake.NewHasher(), shared.FixedClock{Instant: now})
	assert.ErrorIs(t, guard.Authorize(nil, "anything"), paste.ErrNotFound)
}

func TestBelongsToRejectsAnonymousOwnership(t *testing.T) {
	t.Parallel()

	draft := validDraft()
	draft.Owner = 42
	owned, err := newFactory().Create(draft)
	require.NoError(t, err)

	anonymous, err := newFactory().Create(validDraft())
	require.NoError(t, err)

	assert.True(t, owned.BelongsTo(42))
	assert.False(t, owned.BelongsTo(43))
	// Anonymous pastes have no owner, so nobody may claim them.
	assert.False(t, anonymous.BelongsTo(0))
}

func TestSnapshotRoundTripPreservesEveryField(t *testing.T) {
	t.Parallel()

	draft := validDraft()
	draft.Owner = 9
	draft.Secret = "s3cret"
	draft.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)
	draft.IsPublic = false

	original, err := newFactory().Create(draft)
	require.NoError(t, err)

	restored := paste.Rehydrate(original.ToSnapshot())
	assert.Equal(t, original.ToSnapshot(), restored.ToSnapshot())
}
