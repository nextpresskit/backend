package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Petar-V-Nikolov/nextpress-backend/internal/config"
	platformDatabase "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/database"
	platformLogger "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/logger"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/server"

	authApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/application"
	authInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/infrastructure"
	authTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/transport"
	userInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/infrastructure"
)

func main() {
	// Use a dedicated context for the lifetime of the application; this makes it
	// straightforward to propagate graceful shutdown signals to all subsystems.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize structured logging as early as possible so we can rely on a
	// consistent, production-ready logger for all subsequent operations.
	baseLogger, err := platformLogger.New()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	logger := baseLogger.Sugar()
	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(baseLogger)

	logger.Infow("starting nextpress-backend",
		"version", "0.1.0-phase2",
	)

	// Load environment variables (from .env if present) and app configuration
	// before touching any external resources (DB, message buses, etc.).
	config.LoadEnv()
	appCfg := config.LoadAppConfig()
	dbCfg := config.LoadDBConfig()
	jwtCfg := config.LoadJWTConfig()

	// Initialize a single database connection for the lifetime of the process.
	// This avoids connection storms and keeps pooling behaviour predictable.
	db, err := platformDatabase.New(platformDatabase.Config{
		Driver:   dbCfg.Driver,
		Host:     dbCfg.Host,
		Port:     dbCfg.Port,
		User:     dbCfg.User,
		Password: dbCfg.Password,
		Name:     dbCfg.Name,
		SSLMode:  dbCfg.SSLMode,
	})
	if err != nil {
		logger.Fatalw("failed to initialize database connection",
			"error", err,
		)
	}

	// Composition: user repo + hasher + jwt + auth service/handler
	userRepo := userInfra.NewGormRepository(db)
	passwordHasher := authInfra.NewBcryptHasher(0)
	jwtProvider := authInfra.NewJWTProvider(jwtCfg.Secret, jwtCfg.AccessTTL, jwtCfg.RefreshTTL)
	authService := authApp.NewService(userRepo, jwtProvider, passwordHasher)
	authHandler := authTransport.NewHandler(authService)

	// Use Gin as the central HTTP router; we keep the setup centralized in the
	// server package so that future modules can register routes cleanly.
	engine := gin.New()
	server.ConfigureEngine(engine, logger, db)

	// API v1 group
	v1 := engine.Group("/v1")
	authHandler.RegisterRoutes(v1)

	// The Server holds the application configuration and shared dependencies
	// such as the database handle. Additional modules will be layered on top
	// of this container in subsequent phases.
	srv := server.NewServer(engine, appCfg, dbCfg, db, logger)

	// Run the HTTP server in its own goroutine so that we can listen for OS
	// signals and coordinate a controlled shutdown sequence.
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatalw("http server exited with error",
				"error", err,
			)
		}
	}()

	// Capture SIGINT/SIGTERM and use them as a trigger for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logger.Infow("received shutdown signal",
		"signal", sig.String(),
	)

	// Apply a hard timeout to shutdown to avoid hanging the process indefinitely.
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		// At this point we log to the standard log package as a last resort in case
		// the structured logger is already partially torn down.
		log.Printf("graceful shutdown failed: %v\n", err)
	}

	logger.Info("nextpress-backend stopped cleanly")
}
