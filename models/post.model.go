package models

type Post struct {
	ID      *int8   `json:"id"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
	// DateTime *time.Time `json:"datetime"`
	UserId   *int8   `json:"user_id"`
	Author   *string `json:"author,omitempty"`
	UrlID    *string `json:"url_id"`
	IsPublic *int8   `json:"is_public"`
	// PrivacyStatus string    `json:"privacy_status"` // Pode ser "public", "private" ou "unlisted".
	// ExpiryDate    time.Time `json:"expiry_date,omitempty"`
	// Comments string   `json:"comments,omitempty"`
	// Tags   []string `json:"tags,omitempty"`
	// Secret string `json:"secret"`
}
