package requests

type PostRequest struct {
	Title          string `json:"title"           form:"title"           binding:"required,max=255"`
	Content        string `json:"content"         form:"content"         binding:"required,max=1048576"` // 1 MiB cap
	Category       string `json:"category"        form:"category"        binding:"max=255"`
	IsPublic       int8   `json:"is_public"       form:"is_public"`
	ExpirationDate string `json:"expiration_date" form:"expiration_date" binding:"max=40"`
	SecretAccess   string `json:"secret_access"   form:"secret_access"   binding:"max=72"`
}
