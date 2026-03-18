package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Petar-V-Nikolov/nextpress-backend/pkg/seed"
)

func main() {
	loadEnv()
	dsn := buildDSN("")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil && strings.Contains(err.Error(), "unsupported sslmode") {
		dsn = buildDSN("require")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if err := seed.Run(db); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("all seeders completed successfully")
}

func loadEnv() {
	_ = godotenv.Load(".env")
}

func buildDSN(sslmodeOverride string) string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "")
	pass := getEnv("DB_PASSWORD", "")
	db := getEnv("DB_NAME", "")
	sslmode := sslmodeOverride
	if sslmode == "" {
		sslmode = getEnv("DB_SSLMODE", "disable")
	}
	userEnc := url.QueryEscape(user)
	passEnc := url.QueryEscape(pass)
	options := url.QueryEscape("-c search_path=public")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&options=%s", userEnc, passEnc, host, port, db, sslmode, options)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

