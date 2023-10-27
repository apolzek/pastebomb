package requests

type PostRequest struct {
	Title   string `json:"title" form:"title" binding:"required"`
	Content string `json:"content" form:"content" binding:"required"`
	// DateTime time.Time `json:"datetime" form:"datetime" binding:"required"`
	Author string `json:"author" form:"author" binding:"required"`
	URL    string `json:"url" form:"url" binding:"required"`
}
