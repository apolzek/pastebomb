package models

import "time"

type Post struct {
	ID             *int8
	Title          *string
	Content        *string
	Category       *string
	UserID         *int8
	UrlID          *string
	Author         *string
	IsPublic       *int8
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpirationDate *string
	SecretAccess   *string
	NumReports     *int
}
