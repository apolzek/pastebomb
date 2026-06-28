package requests

type UserRequest struct {
	Name     string `json:"name"      form:"name"      binding:"required,max=255"`
	Username string `json:"username"  form:"username"  binding:"required,min=3,max=100"`
	Email    string `json:"email"     form:"email"     binding:"required,email,max=255"`
	BornDate string `json:"born_date" form:"born_date" binding:"required,max=10"`
	Password string `json:"password"  form:"password"  binding:"required,min=8,max=72"`
	IsActive int    `json:"is_active" form:"is_active"`
}
