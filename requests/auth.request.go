package requests

type LoginRequest struct {
	Email    string `json:"email"    form:"email"    binding:"required,email,max=255"`
	Password string `json:"password" form:"password" binding:"required,max=72"`
}
