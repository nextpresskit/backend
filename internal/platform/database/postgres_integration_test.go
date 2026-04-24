//go:build integration

package database

import (
	"os"
	"testing"
)

func TestPostgresConnection_Integration(t *testing.T) {
	cfg := Config{
		Driver:   "postgres",
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Name == "" {
		t.Skip("integration database env vars not set")
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB handle: %v", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("failed to ping postgres: %v", err)
	}
}
