package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	mediaApp "github.com/nextpresskit/backend/internal/modules/media/application"
)

type LocalStorage struct {
	dir            string
	publicBaseURL  string
	maxUploadBytes int64
}

func NewLocalStorage(dir string, publicBaseURL string, maxUploadBytes int64) *LocalStorage {
	return &LocalStorage{
		dir:            dir,
		publicBaseURL:  strings.TrimRight(publicBaseURL, "/"),
		maxUploadBytes: maxUploadBytes,
	}
}

func (s *LocalStorage) Save(ctx context.Context, filename string, contentType string, r io.Reader) (*mediaApp.StoredFile, error) {
	_ = ctx

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) > 12 {
		ext = ""
	}

	// Use a storage name that doesn't leak original filename and is stable enough for caching.
	now := time.Now().UTC().Format("20060102T150405Z")
	rnd := uuid.NewString()
	sum := sha256.Sum256([]byte(now + ":" + rnd + ":" + filename))
	base := hex.EncodeToString(sum[:12])
	storageName := fmt.Sprintf("%s-%s%s", now, base, ext)

	dstPath := filepath.Join(s.dir, storageName)
	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var written int64
	if s.maxUploadBytes > 0 {
		written, err = io.Copy(f, io.LimitReader(r, s.maxUploadBytes+1))
		if err != nil {
			return nil, err
		}
		if written > s.maxUploadBytes {
			_ = os.Remove(dstPath)
			return nil, fmt.Errorf("file too large")
		}
	} else {
		written, err = io.Copy(f, r)
		if err != nil {
			return nil, err
		}
	}

	publicURL := s.publicBaseURL + "/" + storageName
	return &mediaApp.StoredFile{
		StorageName: storageName,
		StoragePath: dstPath,
		PublicURL:   publicURL,
		SizeBytes:   written,
	}, nil
}

