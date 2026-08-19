package shared_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

func TestKindOfDefaultsToInternalForForeignErrors(t *testing.T) {
	t.Parallel()

	assert.Equal(t, shared.KindInternal, shared.KindOf(errors.New("boom")))
	assert.Equal(t, shared.KindNotFound, shared.KindOf(shared.NotFound("x.missing", "missing")))
}

func TestKindOfSeesThroughWrapping(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("while loading: %w", shared.Conflict("x.taken", "taken"))
	assert.Equal(t, shared.KindConflict, shared.KindOf(wrapped))
}

func TestWithCauseKeepsTheOriginalUntouched(t *testing.T) {
	t.Parallel()

	base := shared.Internal("x.failed", "failed")
	cause := errors.New("driver exploded")

	derived := base.WithCause(cause)

	require.NotSame(t, base, derived)
	assert.ErrorIs(t, derived, cause)
	assert.NotErrorIs(t, base, cause)
	// The public message must not carry the cause, which may hold internals.
	assert.Equal(t, "failed", derived.Message)
}
