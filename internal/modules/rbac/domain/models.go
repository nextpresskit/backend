package domain

import "time"

type Role struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Permission struct {
	ID        string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

