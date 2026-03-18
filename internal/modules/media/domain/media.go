package domain

import "time"

type MediaID string

type Media struct {
	ID           MediaID
	UploaderID   string
	OriginalName string
	StorageName  string
	MimeType     string
	SizeBytes    int64
	StoragePath  string
	PublicURL    string
	CreatedAt    time.Time
}

