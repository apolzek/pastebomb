package responses

type PostResponseStore struct {
	ID      *int    `json:"id"`
	Title   *string `json:"name"`
	Content *string `json:"content"`
	Author  *string `json:"author"`
	URL     string  `json:"url"`
}
