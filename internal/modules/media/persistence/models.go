package persistence

import "time"

// Media maps to media (uploader_id is users.id).
type Media struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID         string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	UploaderID   int64     `gorm:"column:uploader_id;not null;index"`
	OriginalName string    `gorm:"column:original_name;not null"`
	StorageName  string    `gorm:"column:storage_name;not null;unique"`
	MimeType     string    `gorm:"column:mime_type;not null"`
	SizeBytes    int64     `gorm:"column:size_bytes;not null"`
	StoragePath  string    `gorm:"column:storage_path;not null"`
	PublicURL    string    `gorm:"column:public_url;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (Media) TableName() string { return "media" }
