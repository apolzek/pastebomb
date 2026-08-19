package fake

import (
	"fmt"
	"sync"
	"time"

	"github.com/apolzek/pastebomb/internal/application"
)

// TokenIssuer hands out predictable tokens and remembers the last identity it
// was asked to sign, so a test can assert which claims were minted.
type TokenIssuer struct {
	mu   sync.Mutex
	last application.Identity
}

func NewTokenIssuer() *TokenIssuer { return &TokenIssuer{} }

var (
	_ application.TokenIssuer   = (*TokenIssuer)(nil)
	_ application.TokenVerifier = (*TokenIssuer)(nil)
)

func (i *TokenIssuer) Issue(identity application.Identity) (application.Token, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.last = identity
	return application.Token{
		Value:     fmt.Sprintf("token-for-%d", int64(identity.UserID)),
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil
}

// Verify accepts only tokens this issuer produced.
func (i *TokenIssuer) Verify(rawToken string) (application.Identity, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if rawToken != fmt.Sprintf("token-for-%d", int64(i.last.UserID)) {
		return application.Identity{}, fmt.Errorf("unknown token")
	}
	return i.last, nil
}

// LastIdentity returns the identity of the most recently issued token.
func (i *TokenIssuer) LastIdentity() application.Identity {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.last
}
