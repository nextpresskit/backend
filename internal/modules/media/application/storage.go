package application

import (
	"context"
	"io"
)

type StoredFile struct {
	StorageName string
	StoragePath string
	PublicURL   string
	SizeBytes   int64
}

type Storage interface {
	Save(ctx context.Context, filename string, contentType string, r io.Reader) (*StoredFile, error)
}

