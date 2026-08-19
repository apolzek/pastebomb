package shared

import "time"

// Clock abstracts "now" so expiration rules stay testable without sleeping.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production implementation, backed by the wall clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// FixedClock always reports the same instant. Intended for tests.
type FixedClock struct{ Instant time.Time }

func (c FixedClock) Now() time.Time { return c.Instant }
