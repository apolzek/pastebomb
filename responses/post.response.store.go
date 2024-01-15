package responses

type PostResponseStore struct {
	ID       *int8   `json:"id"`
	Title    *string `json:"name"`
	Content  *string `json:"content"`
	Category *string `json:"category"`
	UrlID    *string `json:"url_id"`
	IsPublic *int8   `json:"is_public"`
}
