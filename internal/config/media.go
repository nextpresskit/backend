package config

type MediaConfig struct {
	StorageDir     string
	PublicBaseURL  string
	MaxUploadBytes int64
}

func LoadMediaConfig() MediaConfig {
	return MediaConfig{
		StorageDir:     GetEnv("MEDIA_STORAGE_DIR", "storage/uploads"),
		PublicBaseURL:  GetEnv("MEDIA_PUBLIC_BASE_URL", "/uploads"),
		MaxUploadBytes: parseInt64(GetEnv("MEDIA_MAX_UPLOAD_BYTES", "10485760"), 10*1024*1024), // 10MB
	}
}

func parseInt64(v string, fallback int64) int64 {
	n := int64(0)
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c < '0' || c > '9' {
			return fallback
		}
		n = n*10 + int64(c-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
}

