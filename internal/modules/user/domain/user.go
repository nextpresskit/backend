package domain

import "time"

type UserID int64

type User struct {
	ID        int64
	UUID      string
	FirstName string
	LastName  string
	Email     string
	Password  string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
