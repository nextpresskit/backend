package persistence

import (
	"time"

	"gorm.io/gorm"
)

// User maps to the users table (UUID primary key id + bigint public_id).
type User struct {
	PublicID  int64          `gorm:"column:public_id;autoIncrement;uniqueIndex;not null"`
	UUID      string         `gorm:"column:id;type:uuid;primaryKey"`
	FirstName string         `gorm:"not null"`
	LastName  string         `gorm:"not null"`
	Email     string         `gorm:"not null;uniqueIndex"`
	Password  string         `gorm:"not null"`
	Active    bool           `gorm:"not null;default:true"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string { return "users" }
