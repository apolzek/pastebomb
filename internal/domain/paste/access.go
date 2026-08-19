package paste

import "github.com/apolzek/pastebomb/internal/domain/shared"

// AccessGuard is the domain service that decides whether a caller may read a
// paste. Keeping expiration and secret checks together means no read path can
// accidentally honour one rule and forget the other.
type AccessGuard struct {
	hasher shared.Hasher
	clock  shared.Clock
}

func NewAccessGuard(hasher shared.Hasher, clock shared.Clock) *AccessGuard {
	return &AccessGuard{hasher: hasher, clock: clock}
}

// Authorize returns nil when the paste may be served. An expired paste is
// reported as not found so its former existence stays private.
func (g *AccessGuard) Authorize(p *Paste, suppliedSecret string) error {
	if p == nil || p.IsExpired(g.clock.Now()) {
		return ErrNotFound
	}
	return p.Unlock(suppliedSecret, g.hasher)
}
