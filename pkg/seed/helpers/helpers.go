package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

func EnvOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func SeedUUID(namespace, index int) string {
	return fmt.Sprintf("00000000-0000-0000-%04x-%012x", namespace, index)
}

func UserPublicIDFromUUID(tx *gorm.DB, table string, userUUID string) int64 {
	var id int64
	_ = tx.Table(table).Select("id").Where("uuid = ?", userUUID).Scan(&id).Error
	return id
}

func J(s string) json.RawMessage { return json.RawMessage([]byte(s)) }

func Jf(format string, a ...any) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(format, a...))
}

func Int64Ptr(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	return &v
}

func PtrTime(t time.Time) *time.Time { return &t }
