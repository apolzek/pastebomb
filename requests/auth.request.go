package requests

type LoginRequest struct {
	// Email string `json:"email" form:"email" binding:"required,email"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}
