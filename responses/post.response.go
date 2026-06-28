package responses

import "time"

type PostResponse struct {
	ID             *int64     `json:"id"`
	Title          *string    `json:"name"`
	Content        *string    `json:"content"`
	Category       *string    `json:"category"`
	CreatedAt      *time.Time `json:"created_at"`
	UrlID          *string    `json:"url_id"`
	IsPublic       *int8      `json:"is_public"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

type ListPostResponse struct {
	ID    *int64  `json:"id"`
	Title *string `json:"name"`
	UrlID *string `json:"url_id"`
}

type PostResponseStore struct {
	ID       *int64  `json:"id"`
	Title    *string `json:"name"`
	Content  *string `json:"content"`
	Category *string `json:"category"`
	UrlID    *string `json:"url_id"`
	IsPublic *int8   `json:"is_public"`
}
