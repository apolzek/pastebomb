package model

import "time"

type User struct {
	ID        *int
	Name      *string
	Username  *string
	Email     *string
	BornDate  *string
	Password  *string
	IsActive  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
