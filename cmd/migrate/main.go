package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/nextpresskit/backend/internal/appregistry"
	"github.com/nextpresskit/backend/internal/config"
	"github.com/nextpresskit/backend/internal/kit"
	platformdb "github.com/nextpresskit/backend/internal/platform/database"
	"github.com/nextpresskit/backend/internal/platform/dbmigrate"
)

func main() {
	_ = godotenv.Load(".env")

	command := flag.String("command", "up", "up: apply GORM AutoMigrate; drop: remove all public tables (dev only)")
	flag.Parse()

	zl, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	logger := zl.Sugar()
	defer func() { _ = zl.Sync() }()

	dbCfg := config.LoadDBConfig()
	db, err := platformdb.New(platformdb.Config{
		Driver:   dbCfg.Driver,
		Host:     dbCfg.Host,
		Port:     dbCfg.Port,
		User:     dbCfg.User,
		Password: dbCfg.Password,
		Name:     dbCfg.Name,
		SSLMode:  dbCfg.SSLMode,
	})
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	switch *command {
	case "up":
		mods := kit.ResolveModulesFromRegistry(logger, appregistry.ModuleRegistry())
		if err := dbmigrate.AutoMigrateAll(db, mods); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrate up ok (GORM AutoMigrate)")

	case "drop":
		if os.Getenv("ALLOW_SCHEMA_DROP") != "1" {
			log.Fatal("drop refused: set ALLOW_SCHEMA_DROP=1")
		}
		if err := dbmigrate.DropPublicSchema(db); err != nil {
			log.Fatalf("drop: %v", err)
		}
		log.Println("public schema dropped")

	default:
		log.Fatalf("unknown command: %s (use up or drop)", *command)
	}
}
