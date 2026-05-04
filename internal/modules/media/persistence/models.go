package persistence

import "time"

// Media maps to media (uploader_id is users.public_id).
type Media struct {
	ID           string    `gorm:"column:id;type:uuid;primaryKey"`
	UploaderID   int64     `gorm:"column:uploader_id;not null;index"`
	OriginalName string    `gorm:"column:original_name;not null"`
	StorageName  string    `gorm:"column:storage_name;not null;uniqueIndex"`
	MimeType     string    `gorm:"column:mime_type;not null"`
	SizeBytes    int64     `gorm:"column:size_bytes;not null"`
	StoragePath  string    `gorm:"column:storage_path;not null"`
	PublicURL    string    `gorm:"column:public_url;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (Media) TableName() string { return "media" }
