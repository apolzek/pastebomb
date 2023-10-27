package models

type Post struct {
	ID      *int    `json:"id"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
	// DateTime *time.Time `json:"datetime"`
	Author *string `json:"author,omitempty"`
	URL    *string `json:"url"`
	// PrivacyStatus string    `json:"privacy_status"` // Pode ser "public", "private" ou "unlisted".
	// ExpiryDate    time.Time `json:"expiry_date,omitempty"`
	// Comments string   `json:"comments,omitempty"`
	// Tags   []string `json:"tags,omitempty"`
	// Secret string `json:"secret"`
}
