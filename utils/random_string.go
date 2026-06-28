package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString returns a cryptographically-random string of the given
// length, drawn from [a-zA-Z0-9]. Used for unguessable URL slugs and file
// names — do not switch back to math/rand.
func GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}
	max := big.NewInt(int64(len(charset)))
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// crypto/rand failure means the host RNG is broken — refuse to
			// return a predictable string by panicking.
			panic("crypto/rand failed: " + err.Error())
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}
