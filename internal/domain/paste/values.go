package paste

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// ID is the internal identity of a paste. It is never exposed publicly: the
// Slug is the only handle clients get, so pastes cannot be enumerated.
type ID int64

func (id ID) IsZero() bool { return id == 0 }

// OwnerID references the User aggregate by identity only, which keeps the two
// aggregates decoupled. The zero value means an anonymous paste.
type OwnerID int64

func (o OwnerID) IsAnonymous() bool { return o == 0 }

const (
	maxTitleLength    = 255
	maxCategoryLength = 255
	// MaxContentBytes caps a single paste at 1 MiB of plain text.
	MaxContentBytes = 1 << 20
	// SlugLength gives 62^12 possibilities, far beyond brute force over HTTP.
	SlugLength     = 12
	maxSecretBytes = 72
	// maxExpirationHorizon keeps a paste from pinning a row forever through a
	// nonsensical expiration date.
	maxExpirationHorizon = 10 * 365 * 24 * time.Hour
)

// Title is a required, bounded label.
type Title string

func NewTitle(raw string) (Title, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", shared.Invalid("paste.title.empty", "title is required")
	}
	if utf8.RuneCountInString(trimmed) > maxTitleLength {
		return "", shared.Invalid("paste.title.too_long", "title must be at most 255 characters")
	}
	return Title(trimmed), nil
}

func (t Title) String() string { return string(t) }

// Content is the paste body. It is stored and compared as plain text; the
// base64 encoding used by the database is a storage detail of the adapter.
type Content string

func NewContent(raw string) (Content, error) {
	if raw == "" {
		return "", shared.Invalid("paste.content.empty", "content is required")
	}
	if len(raw) > MaxContentBytes {
		return "", shared.Invalid("paste.content.too_large", "content must be at most 1 MiB")
	}
	if !utf8.ValidString(raw) {
		return "", shared.Invalid("paste.content.encoding", "content must be valid UTF-8")
	}
	return Content(raw), nil
}

func (c Content) String() string { return string(c) }
func (c Content) Bytes() []byte  { return []byte(c) }

// Category is an optional free-form tag.
type Category string

func NewCategory(raw string) (Category, error) {
	trimmed := strings.TrimSpace(raw)
	if utf8.RuneCountInString(trimmed) > maxCategoryLength {
		return "", shared.Invalid("paste.category.too_long", "category must be at most 255 characters")
	}
	return Category(trimmed), nil
}

func (c Category) String() string { return string(c) }

// Slug is the unguessable public handle of a paste.
type Slug string

func NewSlug(raw string) (Slug, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", shared.Invalid("paste.slug.empty", "slug is required")
	}
	if utf8.RuneCountInString(trimmed) > 64 {
		return "", shared.Invalid("paste.slug.too_long", "slug is not a valid paste handle")
	}
	return Slug(trimmed), nil
}

func (s Slug) String() string { return string(s) }

// SlugGenerator produces unguessable handles. The domain requires the
// property; infrastructure supplies a crypto/rand implementation.
type SlugGenerator interface {
	Generate() Slug
}

// Visibility decides whether a paste shows up in the public index. An unlisted
// paste is still reachable by whoever holds its slug.
type Visibility uint8

const (
	VisibilityUnlisted Visibility = iota
	VisibilityPublic
)

func VisibilityFrom(isPublic bool) Visibility {
	if isPublic {
		return VisibilityPublic
	}
	return VisibilityUnlisted
}

func (v Visibility) IsPublic() bool { return v == VisibilityPublic }

// Expiration is an optional deletion deadline. The zero value never expires.
type Expiration struct{ at *time.Time }

// NoExpiration is the explicit "keep forever" value.
var NoExpiration = Expiration{}

// NewExpiration parses an RFC3339 timestamp. An empty string means no
// expiration; a past or absurdly distant date is rejected.
func NewExpiration(raw string, now time.Time) (Expiration, error) {
	if strings.TrimSpace(raw) == "" {
		return NoExpiration, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return NoExpiration, shared.Invalid("paste.expiration.format", "expiration date must be an RFC3339 timestamp, for example 2030-12-31T23:59:59Z")
	}
	parsed = parsed.UTC()
	if !parsed.After(now) {
		return NoExpiration, shared.Invalid("paste.expiration.past", "expiration date must be in the future")
	}
	if parsed.After(now.Add(maxExpirationHorizon)) {
		return NoExpiration, shared.Invalid("paste.expiration.too_far", "expiration date must be within 10 years")
	}
	return Expiration{at: &parsed}, nil
}

// RehydrateExpiration rebuilds the value object from a stored timestamp.
func RehydrateExpiration(at *time.Time) Expiration {
	if at == nil {
		return NoExpiration
	}
	utc := at.UTC()
	return Expiration{at: &utc}
}

func (e Expiration) IsSet() bool { return e.at != nil }

func (e Expiration) At() *time.Time {
	if e.at == nil {
		return nil
	}
	at := *e.at
	return &at
}

func (e Expiration) HasPassed(now time.Time) bool {
	return e.at != nil && !now.Before(*e.at)
}

// SecretHash guards a paste behind a shared secret. The empty value means the
// paste is readable by anyone holding the slug.
type SecretHash string

// NewSecretHash hashes the plaintext secret. An empty secret yields an empty
// hash, which the aggregate reads as "no secret required".
func NewSecretHash(plain string, hasher shared.Hasher) (SecretHash, error) {
	if plain == "" {
		return "", nil
	}
	if len(plain) > maxSecretBytes {
		return "", shared.Invalid("paste.secret.too_long", "secret must be at most 72 characters")
	}
	hashed, err := hasher.Hash(plain)
	if err != nil {
		return "", shared.Internal("paste.secret.hash_failed", "could not hash secret").WithCause(err)
	}
	return SecretHash(hashed), nil
}

func (s SecretHash) IsSet() bool    { return s != "" }
func (s SecretHash) String() string { return string(s) }
