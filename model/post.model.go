package model

import "time"

type Post struct {
	ID             *int64
	Title          *string
	Content        *string
	Category       *string
	UserID         *int64
	UrlID          *string
	Author         *string
	IsPublic       *int8
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpirationDate *time.Time
	SecretAccess   *string
	NumReports     *int
}
