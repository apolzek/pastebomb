package requests

type PostRequest struct {
	Title   string `json:"title" form:"title" binding:"required"`
	Content string `json:"content" form:"content" binding:"required"`
	// DateTime time.Time `json:"datetime" form:"datetime" binding:"required"`
	UserId   int    `json:"user_id" form:"user_id"`
	Author   string `json:"author" form:"author"`
	IsPublic int8   `json:"is_public" form:"is_public"`
}
