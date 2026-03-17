package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/Petar-V-Nikolov/nextpress-backend/pkg/migrate"
)

func main() {
	_ = godotenv.Load(".env")

	var (
		command = flag.String("command", "up", "up, down, force, version, drop")
		steps   = flag.Int("steps", 0, "steps for up/down (0 = all)")
		version = flag.Int("version", 0, "version for force")
	)
	flag.Parse()

	dsn := buildDSN("")
	dir := "migrations"
	if d := os.Getenv("MIGRATIONS_DIR"); d != "" {
		dir = d
	}

	r, err := migrate.NewRunner(dsn, dir)
	if err != nil && strings.Contains(err.Error(), "unsupported sslmode") {
		dsn = buildDSN("require")
		r, err = migrate.NewRunner(dsn, dir)
	}
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}
	defer r.Close()

	switch *command {
	case "up":
		if err := r.Up(*steps); err != nil {
			log.Fatalf("up: %v", err)
		}
		log.Println("up ok")

	case "down":
		if err := r.Down(*steps); err != nil {
			log.Fatalf("down: %v", err)
		}
		log.Println("down ok")

	case "force":
		if *version == 0 {
			log.Fatal("force requires -version")
		}
		if err := r.Force(*version); err != nil {
			log.Fatalf("force: %v", err)
		}
		log.Printf("forced version %d", *version)

	case "version":
		ver, dirty, err := r.Version()
		if err != nil {
			log.Fatalf("version: %v", err)
		}
		fmt.Printf("version=%d dirty=%v\n", ver, dirty)

	case "drop":
		if err := r.Drop(); err != nil {
			log.Fatalf("drop: %v", err)
		}
		log.Println("schema_migrations dropped")

	default:
		log.Fatalf("unknown command: %s", *command)
	}
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

