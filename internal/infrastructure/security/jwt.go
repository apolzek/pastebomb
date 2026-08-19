package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/user"
)

const (
	signingAlgorithm = "HS256"
	claimEmail       = "email"
	claimName        = "name"
	claimAdmin       = "admin"
)

// JWTProvider issues and verifies HS256 tokens. It satisfies both
// application.TokenIssuer and application.TokenVerifier.
type JWTProvider struct {
	secret []byte
	issuer string
	ttl    time.Duration
	clock  func() time.Time
}

func NewJWTProvider(secret []byte, issuer string, ttl time.Duration) *JWTProvider {
	return &JWTProvider{secret: secret, issuer: issuer, ttl: ttl, clock: func() time.Time { return time.Now().UTC() }}
}

// Issue mints a token for the given identity.
func (p *JWTProvider) Issue(identity application.Identity) (application.Token, error) {
	now := p.clock()
	expiresAt := now.Add(p.ttl)

	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", int64(identity.UserID)),
		"iss":      p.issuer,
		"iat":      jwt.NewNumericDate(now),
		"nbf":      jwt.NewNumericDate(now),
		"exp":      jwt.NewNumericDate(expiresAt),
		claimEmail: identity.Email,
		claimName:  identity.Name,
		claimAdmin: identity.Admin,
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(p.secret)
	if err != nil {
		return application.Token{}, err
	}
	return application.Token{Value: signed, ExpiresAt: expiresAt}, nil
}

// ErrInvalidToken is returned for any token that fails verification. The
// specific reason stays internal so a caller cannot probe the signing setup.
var ErrInvalidToken = errors.New("invalid token")

// Verify checks the signature, the algorithm and the registered claims.
//
// Pinning the accepted algorithm is what closes the "alg: none" and
// "HMAC-signed with the RSA public key" confusion attacks: the token header
// is never allowed to choose how it gets verified.
func (p *JWTProvider) Verify(rawToken string) (application.Identity, error) {
	parsed, err := jwt.Parse(
		rawToken,
		func(*jwt.Token) (any, error) { return p.secret, nil },
		jwt.WithValidMethods([]string{signingAlgorithm}),
		jwt.WithIssuer(p.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(p.clock),
	)
	if err != nil || !parsed.Valid {
		return application.Identity{}, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return application.Identity{}, ErrInvalidToken
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return application.Identity{}, ErrInvalidToken
	}
	var id int64
	if _, err := fmt.Sscanf(subject, "%d", &id); err != nil || id <= 0 {
		return application.Identity{}, ErrInvalidToken
	}

	admin, _ := claims[claimAdmin].(bool)
	email, _ := claims[claimEmail].(string)
	name, _ := claims[claimName].(string)

	return application.Identity{
		UserID: user.ID(id),
		Email:  email,
		Name:   name,
		Admin:  admin,
	}, nil
}
