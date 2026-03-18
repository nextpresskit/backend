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
	platformMiddleware "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/middleware"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/server"

	authApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/application"
	authInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/infrastructure"
	authTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/transport"
	mediaApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/media/application"
	mediaInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/media/infrastructure"
	mediaTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/media/transport"
	menusApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/application"
	menusInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/infrastructure"
	menusTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/transport"
	pagesApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/application"
	pagesInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/infrastructure"
	pagesTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/transport"
	postsApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/application"
	postsInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/infrastructure"
	postsTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/transport"
	rbacApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/rbac/application"
	rbacInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/rbac/infrastructure"
	rbacTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/rbac/transport"
	taxApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/taxonomy/application"
	taxInfra "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/taxonomy/infrastructure"
	taxTransport "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/taxonomy/transport"
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
	rbacCfg := config.LoadRBACConfig()
	mediaCfg := config.LoadMediaConfig()

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
	permissionChecker := rbacInfra.NewGormPermissionChecker(db)
	rbacRepo := rbacInfra.NewGormRepository(db)
	rbacService := rbacApp.NewService(rbacRepo)
	rbacHandler := rbacTransport.NewHandler(rbacService)
	postsRepo := postsInfra.NewGormRepository(db)
	postsService := postsApp.NewService(postsRepo)
	postsHandler := postsTransport.NewHandler(postsService)
	pagesRepo := pagesInfra.NewGormRepository(db)
	pagesService := pagesApp.NewService(pagesRepo)
	pagesHandler := pagesTransport.NewHandler(pagesService)
	taxRepo := taxInfra.NewGormRepository(db)
	taxService := taxApp.NewService(taxRepo)
	taxHandler := taxTransport.NewHandler(taxService)
	mediaRepo := mediaInfra.NewGormRepository(db)
	mediaStorage := mediaInfra.NewLocalStorage(mediaCfg.StorageDir, mediaCfg.PublicBaseURL, mediaCfg.MaxUploadBytes)
	mediaService := mediaApp.NewService(mediaRepo, mediaStorage)
	mediaHandler := mediaTransport.NewHandler(mediaService)
	menusRepo := menusInfra.NewGormRepository(db)
	menusService := menusApp.NewService(menusRepo)
	menusHandler := menusTransport.NewHandler(menusService)

	// Use Gin as the central HTTP router; we keep the setup centralized in the
	// server package so that future modules can register routes cleanly.
	engine := gin.New()
	// Allow multipart uploads up to configured size (defaults to 10MB).
	engine.MaxMultipartMemory = mediaCfg.MaxUploadBytes
	server.ConfigureEngine(engine, logger, db)

	// API v1 group
	v1 := engine.Group("/v1")
	authHandler.RegisterRoutes(v1)

	// Public APIs (Phase 4): published content, no auth.
	postsHandler.RegisterPublicRoutes(v1)
	pagesHandler.RegisterPublicRoutes(v1)
	menusHandler.RegisterPublicRoutes(v1)

	// Serve uploads in local/dev mode. In production, prefer Nginx for static files.
	if mediaCfg.PublicBaseURL != "" && mediaCfg.PublicBaseURL[0] == '/' {
		engine.StaticFS(mediaCfg.PublicBaseURL, gin.Dir(mediaCfg.StorageDir, false))
	}

	// Admin/content APIs (Phase 3–4): protected, used by CMS/admin UI.
	admin := v1.Group("/admin")
	admin.Use(platformMiddleware.AuthRequired(jwtProvider))

	postsHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	pagesHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	taxHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	mediaHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	menusHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)

	// RBAC admin APIs (Phase 3)
	adminManagement := admin.Group("")
	adminManagement.Use(platformMiddleware.RequirePermission(permissionChecker, "rbac:manage"))
	rbacHandler.RegisterRoutes(adminManagement)

	admin.GET("/ping",
		platformMiddleware.RequirePermission(permissionChecker, "admin:ping"),
		func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		},
	)

	// One-time RBAC bootstrap (explicitly opt-in).
	// This solves the "chicken-and-egg" problem of needing `rbac:manage` to grant `rbac:manage`.
	if rbacCfg.BootstrapEnabled {
		admin.POST("/bootstrap/claim-admin", func(c *gin.Context) {
			var existing int64
			if err := db.WithContext(c.Request.Context()).Table("user_roles").Count(&existing).Error; err != nil {
				c.JSON(500, gin.H{"error": "internal_error"})
				return
			}
			if existing > 0 {
				c.JSON(409, gin.H{"error": "bootstrap_already_completed"})
				return
			}

			userID, _ := c.Get(platformMiddleware.ContextUserIDKey)
			uid, _ := userID.(string)
			if uid == "" {
				c.JSON(401, gin.H{"error": "invalid_user_context"})
				return
			}

			// Seeded admin role ID from seeder `pkg/seed` (run `make seed` / `go run ./cmd/seed`).
			if err := rbacService.AssignRoleToUser(c.Request.Context(), uid, "00000000-0000-0000-0000-000000000001"); err != nil {
				c.JSON(500, gin.H{"error": "internal_error"})
				return
			}

			c.JSON(200, gin.H{"ok": true})
		})
	}

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
