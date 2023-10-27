package requests

type UserRequest struct {
	Name     string `json:"name" form:"name" binding:"required"`
	Email    string `json:"email" form:"email" binding:"required"`
	BornDate string `json:"born_date" form:"born_date" binding:"required"`
	Password string `json:"password" form:"password""`
}
