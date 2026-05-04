package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/nextpresskit/backend/internal/config"
	platformdb "github.com/nextpresskit/backend/internal/platform/database"
	"github.com/nextpresskit/backend/pkg/seed"
)

func main() {
	_ = godotenv.Load(".env")

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

	if err := seed.Run(db); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("all seeders completed successfully")
}
