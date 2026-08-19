package dto

import (
	"time"

	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/user"
)

// Envelope is the success shape of every JSON response.
type Envelope struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Meta carries pagination details alongside a list.
type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Count   int `json:"count"`
}

// ErrorResponse is the failure shape. Code is a stable machine-readable
// identifier; Message is safe to display and never contains internals.
type ErrorResponse struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

// UserView is the full profile, returned only to the account owner and to
// administrators.
type UserView struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	BornDate  string    `json:"born_date"`
	IsActive  bool      `json:"is_active"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PublicUserView is the profile anyone may read. It deliberately omits the
// email address and the date of birth, which the previous public endpoint
// exposed to unauthenticated callers.
type PublicUserView struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserView(account *user.User) UserView {
	return UserView{
		ID:        int64(account.ID()),
		Name:      account.Name().String(),
		Username:  account.Username().String(),
		Email:     account.Email().String(),
		BornDate:  account.BornDate().String(),
		IsActive:  account.IsActive(),
		IsAdmin:   account.IsAdmin(),
		CreatedAt: account.CreatedAt(),
		UpdatedAt: account.UpdatedAt(),
	}
}

func NewUserViews(accounts []*user.User) []UserView {
	views := make([]UserView, 0, len(accounts))
	for _, account := range accounts {
		views = append(views, NewUserView(account))
	}
	return views
}

func NewPublicUserView(account *user.User) PublicUserView {
	return PublicUserView{
		ID:        int64(account.ID()),
		Name:      account.Name().String(),
		Username:  account.Username().String(),
		CreatedAt: account.CreatedAt(),
	}
}

// PasteView is a paste including its body. Only returned once the access
// secret, if any, has been verified.
type PasteView struct {
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Category       string     `json:"category,omitempty"`
	UrlID          string     `json:"url_id"`
	IsPublic       bool       `json:"is_public"`
	HasSecret      bool       `json:"has_secret"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// PasteSummary is a paste without its body. Listings use it so a
// secret-protected paste never leaks its content through an index.
type PasteSummary struct {
	Title          string     `json:"title"`
	Category       string     `json:"category,omitempty"`
	UrlID          string     `json:"url_id"`
	IsPublic       bool       `json:"is_public"`
	HasSecret      bool       `json:"has_secret"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

func NewPasteView(document *paste.Paste) PasteView {
	return PasteView{
		Title:          document.Title().String(),
		Content:        document.Content().String(),
		Category:       document.Category().String(),
		UrlID:          document.Slug().String(),
		IsPublic:       document.Visibility().IsPublic(),
		HasSecret:      document.HasSecret(),
		CreatedAt:      document.CreatedAt(),
		ExpirationDate: document.Expiration().At(),
	}
}

func NewPasteSummary(document *paste.Paste) PasteSummary {
	return PasteSummary{
		Title:          document.Title().String(),
		Category:       document.Category().String(),
		UrlID:          document.Slug().String(),
		IsPublic:       document.Visibility().IsPublic(),
		HasSecret:      document.HasSecret(),
		CreatedAt:      document.CreatedAt(),
		ExpirationDate: document.Expiration().At(),
	}
}

func NewPasteSummaries(documents []*paste.Paste) []PasteSummary {
	summaries := make([]PasteSummary, 0, len(documents))
	for _, document := range documents {
		summaries = append(summaries, NewPasteSummary(document))
	}
	return summaries
}

// NewPasteViews renders full pastes, used by the owner-only endpoint that
// returns the caller's own bodies.
func NewPasteViews(documents []*paste.Paste) []PasteView {
	views := make([]PasteView, 0, len(documents))
	for _, document := range documents {
		views = append(views, NewPasteView(document))
	}
	return views
}

// TokenView is the login response.
type TokenView struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
}
