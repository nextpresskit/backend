package persistence

import (
	"encoding/json"
	"time"
)

// Plugin maps to plugins.
type Plugin struct {
	ID        string          `gorm:"column:id;type:uuid;primaryKey"`
	Name      string          `gorm:"column:name;not null;uniqueIndex"`
	Slug      string          `gorm:"column:slug;not null;uniqueIndex"`
	Enabled   bool            `gorm:"column:enabled;not null"`
	Version   string          `gorm:"column:version;not null"`
	ConfigRaw json.RawMessage `gorm:"column:config;type:jsonb;not null"`
	CreatedAt time.Time       `gorm:"column:created_at;not null"`
	UpdatedAt time.Time       `gorm:"column:updated_at;not null"`
}

func (Plugin) TableName() string { return "plugins" }
