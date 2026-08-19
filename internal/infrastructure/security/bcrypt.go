// Package security provides the cryptographic adapters: password hashing,
// token signing and unguessable handle generation.
package security

import (
	"crypto/subtle"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements shared.Hasher on top of bcrypt.
type BcryptHasher struct{ cost int }

// NewBcryptHasher clamps the cost into the range bcrypt actually supports so a
// misconfigured value degrades to the library default instead of failing every
// registration at runtime.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Matches compares a candidate against a stored hash. bcrypt's own comparison
// is already constant time; the extra guard only rejects an empty stored hash,
// which would otherwise make every candidate fail with a confusing error.
func (h *BcryptHasher) Matches(hash, plain string) bool {
	if subtle.ConstantTimeEq(int32(len(hash)), 0) == 1 {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
