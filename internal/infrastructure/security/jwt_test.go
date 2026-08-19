package security_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/user"
	"github.com/apolzek/pastebomb/internal/infrastructure/security"
)

var testSecret = []byte("a-test-secret-that-is-long-enough-32")

func newProvider() *security.JWTProvider {
	return security.NewJWTProvider(testSecret, "pastebomb", time.Hour)
}

func TestIssueThenVerifyRoundTripsTheIdentity(t *testing.T) {
	t.Parallel()

	provider := newProvider()
	issued := application.Identity{UserID: 42, Email: "ada@example.com", Name: "Ada", Admin: true}

	token, err := provider.Issue(issued)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), token.ExpiresAt, time.Minute)

	verified, err := provider.Verify(token.Value)
	require.NoError(t, err)
	assert.Equal(t, issued, verified)
}

func TestVerifyRejectsGarbage(t *testing.T) {
	t.Parallel()

	provider := newProvider()

	for name, raw := range map[string]string{
		"empty":         "",
		"not a jwt":     "hello",
		"two segments":  "aaa.bbb",
		"bad signature": "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.wrongsignature",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := provider.Verify(raw)
			assert.ErrorIs(t, err, security.ErrInvalidToken)
		})
	}
}

// A token signed with the "none" algorithm carries no proof of origin. The
// verifier must reject it on the algorithm alone, never on the claims.
func TestVerifyRejectsTheNoneAlgorithm(t *testing.T) {
	t.Parallel()

	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "1",
		"iss": "pastebomb",
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = newProvider().Verify(unsigned)
	assert.ErrorIs(t, err, security.ErrInvalidToken)
}

func TestVerifyRejectsATokenSignedWithAnotherSecret(t *testing.T) {
	t.Parallel()

	foreign := security.NewJWTProvider([]byte("another-secret-long-enough-for-hs256"), "pastebomb", time.Hour)

	token, err := foreign.Issue(application.Identity{UserID: 1})
	require.NoError(t, err)

	_, err = newProvider().Verify(token.Value)
	assert.ErrorIs(t, err, security.ErrInvalidToken)
}

func TestVerifyRejectsAnotherIssuer(t *testing.T) {
	t.Parallel()

	foreign := security.NewJWTProvider(testSecret, "someone-else", time.Hour)

	token, err := foreign.Issue(application.Identity{UserID: 1})
	require.NoError(t, err)

	_, err = newProvider().Verify(token.Value)
	assert.ErrorIs(t, err, security.ErrInvalidToken)
}

func TestVerifyRejectsAnExpiredToken(t *testing.T) {
	t.Parallel()

	expired := security.NewJWTProvider(testSecret, "pastebomb", -time.Minute)

	token, err := expired.Issue(application.Identity{UserID: 1})
	require.NoError(t, err)

	_, err = newProvider().Verify(token.Value)
	assert.ErrorIs(t, err, security.ErrInvalidToken)
}

// A token without a usable subject must not authenticate as user zero, which
// the application layer treats as "anonymous".
func TestVerifyRejectsMissingOrUnusableSubject(t *testing.T) {
	t.Parallel()

	for name, subject := range map[string]any{
		"absent":   nil,
		"zero":     "0",
		"negative": "-3",
		"words":    "admin",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			claims := jwt.MapClaims{
				"iss": "pastebomb",
				"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
			}
			if subject != nil {
				claims["sub"] = subject
			}

			raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
			require.NoError(t, err)

			identity, err := newProvider().Verify(raw)
			assert.ErrorIs(t, err, security.ErrInvalidToken)
			assert.Equal(t, user.ID(0), identity.UserID)
		})
	}
}

// The admin claim decides access to the administrative routes, so a token that
// omits it or sets it to a non-boolean must not be read as an admin.
func TestVerifyDefaultsAdminToFalse(t *testing.T) {
	t.Parallel()

	for name, admin := range map[string]any{
		"absent":        nil,
		"string true":   "true",
		"number one":    1,
		"boolean false": false,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			claims := jwt.MapClaims{
				"sub": "5",
				"iss": "pastebomb",
				"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
			}
			if admin != nil {
				claims["admin"] = admin
			}

			raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
			require.NoError(t, err)

			identity, err := newProvider().Verify(raw)
			require.NoError(t, err)
			assert.False(t, identity.Admin)
		})
	}
}
