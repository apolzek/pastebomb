// Package paste is the Paste aggregate: the pasted document, its public
// handle, its lifetime and its access rules.
package paste

import (
	"time"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// Paste is the aggregate root.
type Paste struct {
	id         ID
	slug       Slug
	owner      OwnerID
	title      Title
	content    Content
	category   Category
	visibility Visibility
	expiration Expiration
	secret     SecretHash
	createdAt  time.Time
	updatedAt  time.Time
}

// Draft is the unvalidated input of a paste creation.
type Draft struct {
	Owner     OwnerID
	Title     string
	Content   string
	Category  string
	IsPublic  bool
	ExpiresAt string
	Secret    string
}

// Factory creates Pastes. It exists because building one needs collaborators
// (slug source, hasher, clock) that the aggregate must not reach for itself.
type Factory struct {
	slugs  SlugGenerator
	hasher shared.Hasher
	clock  shared.Clock
}

func NewFactory(slugs SlugGenerator, hasher shared.Hasher, clock shared.Clock) *Factory {
	return &Factory{slugs: slugs, hasher: hasher, clock: clock}
}

// Create validates a draft and returns a Paste that satisfies every invariant.
func (f *Factory) Create(draft Draft) (*Paste, error) {
	now := f.clock.Now()

	title, err := NewTitle(draft.Title)
	if err != nil {
		return nil, err
	}
	content, err := NewContent(draft.Content)
	if err != nil {
		return nil, err
	}
	category, err := NewCategory(draft.Category)
	if err != nil {
		return nil, err
	}
	expiration, err := NewExpiration(draft.ExpiresAt, now)
	if err != nil {
		return nil, err
	}
	secret, err := NewSecretHash(draft.Secret, f.hasher)
	if err != nil {
		return nil, err
	}

	return &Paste{
		slug:       f.slugs.Generate(),
		owner:      draft.Owner,
		title:      title,
		content:    content,
		category:   category,
		visibility: VisibilityFrom(draft.IsPublic),
		expiration: expiration,
		secret:     secret,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// Snapshot is the flat representation exchanged with the persistence layer.
type Snapshot struct {
	ID         int64
	Slug       string
	OwnerID    int64
	Title      string
	Content    string
	Category   string
	IsPublic   bool
	ExpiresAt  *time.Time
	SecretHash string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Rehydrate rebuilds a stored paste without re-running creation invariants.
func Rehydrate(snapshot Snapshot) *Paste {
	return &Paste{
		id:         ID(snapshot.ID),
		slug:       Slug(snapshot.Slug),
		owner:      OwnerID(snapshot.OwnerID),
		title:      Title(snapshot.Title),
		content:    Content(snapshot.Content),
		category:   Category(snapshot.Category),
		visibility: VisibilityFrom(snapshot.IsPublic),
		expiration: RehydrateExpiration(snapshot.ExpiresAt),
		secret:     SecretHash(snapshot.SecretHash),
		createdAt:  snapshot.CreatedAt,
		updatedAt:  snapshot.UpdatedAt,
	}
}

func (p *Paste) ToSnapshot() Snapshot {
	return Snapshot{
		ID:         int64(p.id),
		Slug:       p.slug.String(),
		OwnerID:    int64(p.owner),
		Title:      p.title.String(),
		Content:    p.content.String(),
		Category:   p.category.String(),
		IsPublic:   p.visibility.IsPublic(),
		ExpiresAt:  p.expiration.At(),
		SecretHash: p.secret.String(),
		CreatedAt:  p.createdAt,
		UpdatedAt:  p.updatedAt,
	}
}

func (p *Paste) ID() ID                 { return p.id }
func (p *Paste) Slug() Slug             { return p.slug }
func (p *Paste) Owner() OwnerID         { return p.owner }
func (p *Paste) Title() Title           { return p.title }
func (p *Paste) Content() Content       { return p.content }
func (p *Paste) Category() Category     { return p.category }
func (p *Paste) Visibility() Visibility { return p.visibility }
func (p *Paste) Expiration() Expiration { return p.expiration }
func (p *Paste) HasSecret() bool        { return p.secret.IsSet() }
func (p *Paste) CreatedAt() time.Time   { return p.createdAt }
func (p *Paste) UpdatedAt() time.Time   { return p.updatedAt }

// AssignID stamps the identity produced by the store. It is only valid once.
func (p *Paste) AssignID(id ID) {
	if p.id.IsZero() {
		p.id = id
	}
}

// IsExpired reports whether the paste should no longer be served. The purge
// job removes such rows eventually; reads must not wait for it.
func (p *Paste) IsExpired(now time.Time) bool { return p.expiration.HasPassed(now) }

// BelongsTo reports whether owner is the author of this paste.
func (p *Paste) BelongsTo(owner OwnerID) bool {
	return !p.owner.IsAnonymous() && p.owner == owner
}

// Unlock authorizes a read. A paste without a secret is readable by whoever
// knows the slug; otherwise the supplied secret must match the stored hash.
func (p *Paste) Unlock(suppliedSecret string, hasher shared.Hasher) error {
	if !p.secret.IsSet() {
		return nil
	}
	if !hasher.Matches(p.secret.String(), suppliedSecret) {
		return ErrSecretRequired
	}
	return nil
}
