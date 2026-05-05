package persistence

import (
	"time"

	"gorm.io/gorm"
)

// User maps to the users table (bigint primary key id + public uuid).
type User struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string         `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	FirstName string         `gorm:"not null"`
	LastName  string         `gorm:"not null"`
	Email     string         `gorm:"not null;unique"`
	Password  string         `gorm:"not null"`
	Active    bool           `gorm:"not null;default:true"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string { return "users" }
