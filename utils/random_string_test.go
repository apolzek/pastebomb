package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString_Length(t *testing.T) {
	for _, n := range []int{1, 5, 16, 64} {
		assert.Len(t, GenerateRandomString(n), n)
	}
}

func TestGenerateRandomString_Charset(t *testing.T) {
	const expected = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	s := GenerateRandomString(200)
	for _, c := range s {
		assert.True(t, strings.ContainsRune(expected, c), "unexpected char %q", c)
	}
}

func TestGenerateRandomString_Zero(t *testing.T) {
	assert.Equal(t, "", GenerateRandomString(0))
}

func TestGenerateRandomString_Negative(t *testing.T) {
	assert.Equal(t, "", GenerateRandomString(-5))
}

// Two back-to-back calls should not collide. With 62^16 ≈ 4.7e28 possible
// outputs the chance of a false positive is negligible.
func TestGenerateRandomString_Unique(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		s := GenerateRandomString(16)
		_, dup := seen[s]
		assert.False(t, dup, "duplicate string from crypto/rand generator")
		seen[s] = struct{}{}
	}
}
