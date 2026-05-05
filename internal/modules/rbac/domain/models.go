package domain

import "time"

type Role struct {
	ID        int64
	UUID      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Permission struct {
	ID        int64
	UUID      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
