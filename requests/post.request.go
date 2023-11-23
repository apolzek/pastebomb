package requests

type PostRequest struct {
	Title    string `json:"title" form:"title" binding:"required"`
	Content  string `json:"content" form:"content" binding:"required"`
	IsPublic int8   `json:"is_public" form:"is_public"`
}
