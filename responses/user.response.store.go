package responses

type UserResponseStore struct {
	ID       *int    `json:"id"`
	Name     *string `json:"name"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	BornDate *string `json:"born_date"`
}
