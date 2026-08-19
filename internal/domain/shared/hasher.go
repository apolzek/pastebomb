package shared

// Hasher is the port for one-way hashing of secrets (user passwords and paste
// access secrets). The domain states the need; infrastructure picks bcrypt.
type Hasher interface {
	Hash(plain string) (string, error)
	// Matches must be resistant to timing analysis.
	Matches(hash, plain string) bool
}
