package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// Round-trip: a bcrypt hash of a password verifies against that password
// and rejects a different one. This pins the behavior the login + signup
// controllers depend on.
func TestBcryptRoundTrip(t *testing.T) {
	pwd := []byte("correct-horse-battery-staple")
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	assert.NoError(t, err)
	assert.NotEqual(t, string(pwd), string(hash))

	assert.NoError(t, bcrypt.CompareHashAndPassword(hash, pwd))
	assert.Error(t, bcrypt.CompareHashAndPassword(hash, []byte("wrong-password")))
}
