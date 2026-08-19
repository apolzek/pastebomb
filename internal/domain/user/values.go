package user

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// ID is the identity of the User aggregate. Zero means "not persisted yet".
type ID int64

func (id ID) IsZero() bool { return id == 0 }

const (
	maxNameLength     = 255
	minUsernameLength = 3
	maxUsernameLength = 100
	maxEmailLength    = 255
	minPasswordLength = 8
	// bcrypt silently truncates beyond 72 bytes, so reject longer inputs
	// instead of accepting a password that is only partly checked.
	maxPasswordLength = 72
	bornDateLayout    = "2006-01-02"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Name is the user's display name.
type Name string

func NewName(raw string) (Name, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", shared.Invalid("user.name.empty", "name is required")
	}
	if utf8.RuneCountInString(trimmed) > maxNameLength {
		return "", shared.Invalid("user.name.too_long", "name must be at most 255 characters")
	}
	return Name(trimmed), nil
}

func (n Name) String() string { return string(n) }

// Username is the public handle, unique across the system.
type Username string

func NewUsername(raw string) (Username, error) {
	trimmed := strings.TrimSpace(raw)
	length := utf8.RuneCountInString(trimmed)
	if length < minUsernameLength || length > maxUsernameLength {
		return "", shared.Invalid("user.username.length", "username must be between 3 and 100 characters")
	}
	if !usernamePattern.MatchString(trimmed) {
		return "", shared.Invalid("user.username.charset", "username may only contain letters, digits, dot, underscore and hyphen")
	}
	return Username(strings.ToLower(trimmed)), nil
}

func (u Username) String() string { return string(u) }

// Email is a normalized, syntactically valid address.
type Email string

func NewEmail(raw string) (Email, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return "", shared.Invalid("user.email.empty", "email is required")
	}
	if len(trimmed) > maxEmailLength {
		return "", shared.Invalid("user.email.too_long", "email must be at most 255 characters")
	}
	address, err := mail.ParseAddress(trimmed)
	if err != nil || address.Address != trimmed {
		return "", shared.Invalid("user.email.invalid", "email is not a valid address")
	}
	return Email(trimmed), nil
}

func (e Email) String() string { return string(e) }

// BornDate is a calendar date that must not sit in the future.
type BornDate struct{ value time.Time }

func NewBornDate(raw string, now time.Time) (BornDate, error) {
	parsed, err := time.Parse(bornDateLayout, strings.TrimSpace(raw))
	if err != nil {
		return BornDate{}, shared.Invalid("user.born_date.format", "born_date must be formatted as YYYY-MM-DD")
	}
	if parsed.After(now) {
		return BornDate{}, shared.Invalid("user.born_date.future", "born_date must not be in the future")
	}
	return BornDate{value: parsed}, nil
}

// RehydrateBornDate rebuilds the value object from storage without re-running
// the future check, since stored data was already validated on the way in.
func RehydrateBornDate(value time.Time) BornDate { return BornDate{value: value} }

func (b BornDate) Time() time.Time { return b.value }
func (b BornDate) String() string  { return b.value.Format(bornDateLayout) }

// PasswordHash is a one-way hash. The plaintext never becomes part of the
// aggregate, so it cannot leak through a getter or a log statement.
type PasswordHash string

// NewPasswordHash validates the plaintext policy and immediately hashes it.
func NewPasswordHash(plain string, hasher shared.Hasher) (PasswordHash, error) {
	if len(plain) < minPasswordLength {
		return "", shared.Invalid("user.password.too_short", "password must be at least 8 characters")
	}
	if len(plain) > maxPasswordLength {
		return "", shared.Invalid("user.password.too_long", "password must be at most 72 characters")
	}
	hashed, err := hasher.Hash(plain)
	if err != nil {
		return "", shared.Internal("user.password.hash_failed", "could not hash password").WithCause(err)
	}
	return PasswordHash(hashed), nil
}

func (p PasswordHash) String() string { return string(p) }
