package security_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/apolzek/pastebomb/internal/infrastructure/security"
)

// A low cost keeps the suite fast. Production cost comes from configuration.
const testCost = bcrypt.MinCost

func TestHashIsSaltedAndNeverContainsThePlaintext(t *testing.T) {
	t.Parallel()

	hasher := security.NewBcryptHasher(testCost)

	first, err := hasher.Hash("correct horse battery")
	require.NoError(t, err)
	second, err := hasher.Hash("correct horse battery")
	require.NoError(t, err)

	// Distinct salts mean equal passwords produce different hashes, so a
	// leaked table does not reveal which users share a password.
	assert.NotEqual(t, first, second)
	assert.NotContains(t, first, "correct horse battery")
	assert.True(t, strings.HasPrefix(first, "$2"))
}

func TestMatchesAcceptsOnlyTheOriginalPlaintext(t *testing.T) {
	t.Parallel()

	hasher := security.NewBcryptHasher(testCost)

	hash, err := hasher.Hash("correct horse battery")
	require.NoError(t, err)

	assert.True(t, hasher.Matches(hash, "correct horse battery"))
	assert.False(t, hasher.Matches(hash, "Correct horse battery"))
	assert.False(t, hasher.Matches(hash, ""))
}

// An account row with no stored hash must never authenticate, whatever the
// candidate is.
func TestMatchesRejectsAnEmptyStoredHash(t *testing.T) {
	t.Parallel()

	hasher := security.NewBcryptHasher(testCost)

	assert.False(t, hasher.Matches("", ""))
	assert.False(t, hasher.Matches("", "anything"))
}

func TestOutOfRangeCostFallsBackToTheDefault(t *testing.T) {
	t.Parallel()

	for _, cost := range []int{-1, 0, 99} {
		hasher := security.NewBcryptHasher(cost)

		hash, err := hasher.Hash("some password")
		require.NoError(t, err)

		actual, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)
		assert.Equal(t, bcrypt.DefaultCost, actual)
	}
}
