package post_controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestParseExpiration_Empty(t *testing.T) {
	got, err := parseExpiration("")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestParseExpiration_Future(t *testing.T) {
	in := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	got, err := parseExpiration(in)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestParseExpiration_Past(t *testing.T) {
	in := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	_, err := parseExpiration(in)
	assert.Error(t, err)
}

func TestParseExpiration_BadFormat(t *testing.T) {
	_, err := parseExpiration("not-a-date")
	assert.Error(t, err)
}

func TestHashSecretAccess_Empty(t *testing.T) {
	got, err := hashSecretAccess("")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestHashSecretAccess_RoundTrip(t *testing.T) {
	hash, err := hashSecretAccess("opensesame")
	assert.NoError(t, err)
	assert.NotNil(t, hash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(*hash), []byte("opensesame")))
}

func TestCheckSecretAccess(t *testing.T) {
	// No secret → always allowed.
	assert.True(t, checkSecretAccess(nil, ""))
	assert.True(t, checkSecretAccess(nil, "anything"))

	empty := ""
	assert.True(t, checkSecretAccess(&empty, ""))

	hash, _ := hashSecretAccess("right")
	assert.True(t, checkSecretAccess(hash, "right"))
	assert.False(t, checkSecretAccess(hash, "wrong"))
	assert.False(t, checkSecretAccess(hash, ""))
}
