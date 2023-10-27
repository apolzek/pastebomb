package responses

import "time"

type PostResponse struct {
	ID       *int      `json:"id"`
	Title    *string   `json:"name"`
	Content  *string   `json:"content"`
	DateTime time.Time `json:"datetime"`
	Author   *string   `json:"author"`
	URL      *string   `json:"url"`
}
