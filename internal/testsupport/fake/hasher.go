// Package fake holds deterministic stand-ins for infrastructure adapters.
package fake

import "strings"

// hashPrefix marks a value produced by Hasher so a test can spot an unhashed
// value being stored by mistake.
const hashPrefix = "fake-hash:"

// Hasher is a reversible stand-in for bcrypt. It is deterministic and
// instantaneous, which keeps domain tests fast, and it is obviously unsuitable
// for anything outside tests.
type Hasher struct{}

func NewHasher() *Hasher { return &Hasher{} }

func (Hasher) Hash(plain string) (string, error) { return hashPrefix + plain, nil }

func (Hasher) Matches(hash, plain string) bool {
	return hash != "" && hash == hashPrefix+plain
}

// IsHashed reports whether value went through Hash.
func IsHashed(value string) bool { return strings.HasPrefix(value, hashPrefix) }
