package domain

import "time"

type UserID string

type User struct {
	ID        UserID
	FirstName string
	LastName  string
	Email     string
	Password  string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
