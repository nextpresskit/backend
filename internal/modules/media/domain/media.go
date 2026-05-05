package domain

import "time"

type MediaID int64

type Media struct {
	ID           MediaID
	UUID         string
	UploaderID   string
	OriginalName string
	StorageName  string
	MimeType     string
	SizeBytes    int64
	StoragePath  string
	PublicURL    string
	CreatedAt    time.Time
}
