package models

import "time"

type User struct {
	ID        *int      `json:"id"`
	Name      *string   `json:"name"`
	Username  *string   `json:"username"`
	Email     *string   `json:"email"`
	BornDate  *string   `json:"born_date"`
	Password  *string   `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
