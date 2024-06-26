package requests

type PostRequest struct {
	Title          string `json:"title" form:"title" binding:"required"`
	Content        string `json:"content" form:"content" binding:"required"`
	Category       string `json:"category" form:"category"`
	IsPublic       int8   `json:"is_public" form:"is_public"`
	ExpirationDate string `json:"expiration_date" form:"expiration_date"`
	SecretAccess   string `json:"secret_access" form:"secret_access"`
}
