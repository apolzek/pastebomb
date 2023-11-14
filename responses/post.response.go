package responses

import "time"

type PostResponse struct {
	ID        *int       `json:"id"`
	Title     *string    `json:"name"`
	Content   *string    `json:"content"`
	CreatedAt *time.Time `json:"created_at"`
	Author    *string    `json:"author"`
	UrlID     *string    `json:"url_id"`
	IsPublic  *int8      `json:"is_public"`
}
