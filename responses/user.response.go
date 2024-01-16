package responses

import "time"

type UserResponse struct {
	ID        *int       `json:"id"`
	Name      *string    `json:"name"`
	Username  *string    `json:"username"`
	Email     *string    `json:"email"`
	BornDate  *string    `json:"born_date"`
	IsActive  *int       `json:"is_active"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UserResponseStore struct {
	ID       *int    `json:"id"`
	Name     *string `json:"name"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	BornDate *string `json:"born_date"`
}
