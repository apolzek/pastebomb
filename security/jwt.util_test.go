package security

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndDecodeToken(t *testing.T) {
	claims := &jwt.MapClaims{
		"user_id": float64(42),
		"role":    "admin",
	}

	tokenStr, err := GenerateToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	decoded, err := DecodeToken(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, float64(42), decoded["user_id"])
	assert.Equal(t, "admin", decoded["role"])
}

func TestVerifyToken_Valid(t *testing.T) {
	tokenStr, err := GenerateToken(&jwt.MapClaims{"sub": "alice"})
	assert.NoError(t, err)

	tok, err := VerifyToken(tokenStr)
	assert.NoError(t, err)
	assert.True(t, tok.Valid)
}

func TestVerifyToken_Malformed(t *testing.T) {
	_, err := VerifyToken("not-a-valid-token")
	assert.Error(t, err)
}

func TestDecodeToken_WrongSigningMethod(t *testing.T) {
	// Signed with `none`, but the verifier requires HMAC.
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "x"})
	signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	_, err = DecodeToken(signed)
	assert.Error(t, err)
}

// A token signed with a different HMAC key must not verify against ours.
func TestDecodeToken_WrongSecret(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "x"})
	signed, err := tok.SignedString([]byte("a-completely-different-secret"))
	assert.NoError(t, err)

	_, err = DecodeToken(signed)
	assert.Error(t, err)
}

// Confirms loadSecret returns a non-empty key (env-loaded or fallback).
// We don't assert the value because the fallback is intentional in dev/test.
func TestLoadSecret_NonEmpty(t *testing.T) {
	assert.NotEmpty(t, loadSecret())
}
