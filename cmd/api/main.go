package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/nextpresskit/backend/internal/config"
	gqlapi "github.com/nextpresskit/backend/internal/graphql"
	"github.com/nextpresskit/backend/internal/graphql/generated"
	platformDatabase "github.com/nextpresskit/backend/internal/platform/database"
	platformES "github.com/nextpresskit/backend/internal/platform/elasticsearch"
	platformLogger "github.com/nextpresskit/backend/internal/platform/logger"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
	"github.com/nextpresskit/backend/internal/server"

	authApp "github.com/nextpresskit/backend/internal/modules/auth/application"
	authInfra "github.com/nextpresskit/backend/internal/modules/auth/infrastructure"
	authTransport "github.com/nextpresskit/backend/internal/modules/auth/transport"
	mediaApp "github.com/nextpresskit/backend/internal/modules/media/application"
	mediaInfra "github.com/nextpresskit/backend/internal/modules/media/infrastructure"
	mediaTransport "github.com/nextpresskit/backend/internal/modules/media/transport"
	pluginsApp "github.com/nextpresskit/backend/internal/modules/plugins/application"
	pluginsInfra "github.com/nextpresskit/backend/internal/modules/plugins/infrastructure"
	pluginsTransport "github.com/nextpresskit/backend/internal/modules/plugins/transport"
	pagesApp "github.com/nextpresskit/backend/internal/modules/pages/application"
	pagesInfra "github.com/nextpresskit/backend/internal/modules/pages/infrastructure"
	pagesTransport "github.com/nextpresskit/backend/internal/modules/pages/transport"
	postsApp "github.com/nextpresskit/backend/internal/modules/posts/application"
	postsIdent "github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	postsInfra "github.com/nextpresskit/backend/internal/modules/posts/infrastructure"
	postsTransport "github.com/nextpresskit/backend/internal/modules/posts/transport"
	rbacApp "github.com/nextpresskit/backend/internal/modules/rbac/application"
	rbacInfra "github.com/nextpresskit/backend/internal/modules/rbac/infrastructure"
	rbacTransport "github.com/nextpresskit/backend/internal/modules/rbac/transport"
	taxApp "github.com/nextpresskit/backend/internal/modules/taxonomy/application"
	taxInfra "github.com/nextpresskit/backend/internal/modules/taxonomy/infrastructure"
	taxTransport "github.com/nextpresskit/backend/internal/modules/taxonomy/transport"
	userInfra "github.com/nextpresskit/backend/internal/modules/user/infrastructure"
)

var version = "dev"

func main() {
	// Use a dedicated context for the lifetime of the application; this makes it
	// straightforward to propagate graceful shutdown signals to all subsystems.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.LoadEnv()

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

	appCfg := config.LoadAppConfig()
	logger.Infow("starting",
		"service", appCfg.LogIdentifier,
		"version", version,
	)
	dbCfg := config.LoadDBConfig()
	jwtCfg := config.LoadJWTConfig()
	rbacCfg := config.LoadRBACConfig()
	mediaCfg := config.LoadMediaConfig()
	rateCfg := config.LoadRateLimitConfig()
	graphqlCfg := config.LoadGraphQLConfig()
	esCfg := config.LoadElasticsearchConfig(appCfg.Env)

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
	authService.SetRBACReader(rbacInfra.NewGormRepository(db))
	authHandler := authTransport.NewHandler(authService, platformMiddleware.AuthRequired(jwtProvider, jwtCfg), jwtCfg)
	permissionChecker := rbacInfra.NewGormPermissionChecker(db)
	rbacRepo := rbacInfra.NewGormRepository(db)
	rbacService := rbacApp.NewService(rbacRepo)
	rbacHandler := rbacTransport.NewHandler(rbacService)

	// Plugins repo + hook bootstrap before posts service so post save hooks are wired.
	pluginsRepo := pluginsInfra.NewGormRepository(db)
	postHooks, enabledPluginCount, err := pluginsApp.BootstrapPostHooks(ctx, pluginsRepo)
	if err != nil {
		logger.Fatalw("failed to bootstrap plugin hooks",
			"error", err,
		)
	}
	logger.Infow("plugin hooks bootstrapped", "enabled_plugins", enabledPluginCount)

	postsRepo := postsInfra.NewRepositoryAdapter(postsInfra.NewGormRepository(db))
	derivedHook := postsApp.NewDerivedFieldsHook(postsRepo)

	esClient, err := platformES.NewClient(esCfg)
	if err != nil {
		logger.Fatalw("failed to create elasticsearch client",
			"error", err,
		)
	}
	if esClient != nil && esCfg.Enabled {
		if pingErr := platformES.Ping(ctx, esClient); pingErr != nil {
			logger.Warnw("elasticsearch ping failed; search and indexing may fail until the cluster is reachable",
				"error", pingErr,
			)
		}
	}
	if esCfg.Enabled && len(esCfg.Addresses) == 0 {
		logger.Warnw("ELASTICSEARCH_ENABLED is true but ELASTICSEARCH_URLS is empty; indexing and search are inactive")
	}
	postsIdx := platformES.NewPostsIndex(esClient, esCfg, logger)
	if postsIdx != nil {
		logger.Infow("elasticsearch integration active",
			"index", postsIdx.Name(),
			"nodes", len(esCfg.Addresses),
		)
	}
	if postsIdx != nil && esCfg.AutoCreateIndex {
		if err := postsIdx.EnsureIndex(ctx); err != nil {
			logger.Fatalw("elasticsearch index setup failed",
				"error", err,
			)
		}
	}
	esHook := platformES.NewPostIndexHook(logger, postsIdx, postsRepo)
	hooks := postsApp.NewPostSaveChain(derivedHook, postHooks)
	if esHook != nil {
		hooks = postsApp.NewPostSaveChain(derivedHook, postHooks, esHook)
	}
	postsService := postsApp.NewService(postsRepo, hooks)
	postsHandler := postsTransport.NewHandlerFromServiceWithOptionalSearch(postsService, postsIdx)

	// Background loop: promote scheduled posts when due.
	go func() {
		const scheduledPublishSQL = `
UPDATE posts
 SET status = 'published',
     published_at = COALESCE(published_at, NOW()),
     workflow_stage = 'published',
     updated_at = NOW()
 WHERE scheduled_publish_at IS NOT NULL
   AND scheduled_publish_at <= NOW()
   AND status <> 'published'`
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if postsIdx == nil {
					res := db.WithContext(ctx).Exec(scheduledPublishSQL)
					if res.Error != nil {
						logger.Errorw("scheduled publish loop failed", "error", res.Error)
						continue
					}
					if res.RowsAffected > 0 {
						logger.Infow("scheduled posts promoted", "count", res.RowsAffected)
					}
					continue
				}
				rows, qerr := db.WithContext(ctx).Raw(scheduledPublishSQL + ` RETURNING id`).Rows()
				if qerr != nil {
					logger.Errorw("scheduled publish loop failed", "error", qerr)
					continue
				}
				func() {
					defer rows.Close()
					var n int64
					for rows.Next() {
						var id string
						if scanErr := rows.Scan(&id); scanErr != nil {
							logger.Errorw("scheduled publish scan failed", "error", scanErr)
							return
						}
						n++
						p, findErr := postsRepo.FindByID(ctx, postsIdent.PostID(id))
						if findErr != nil || p == nil {
							logger.Warnw("scheduled publish post reload failed", "post_id", id, "error", findErr)
							continue
						}
						postsIdx.SyncPost(ctx, p)
					}
					if n > 0 {
						logger.Infow("scheduled posts promoted", "count", n)
					}
				}()
			}
		}
	}()

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

	pluginsService := pluginsApp.NewService(pluginsRepo)
	pluginsHandler := pluginsTransport.NewHandler(pluginsService)

	// Use Gin as the central HTTP router; we keep the setup centralized in the
	// server package so that future modules can register routes cleanly.
	engine := gin.New()
	// Allow multipart uploads up to configured size (defaults to 10MB).
	engine.MaxMultipartMemory = mediaCfg.MaxUploadBytes
	readinessChecks := make([]server.ReadinessCheck, 0, 1)
	if postsIdx != nil {
		readinessChecks = append(readinessChecks, server.ReadinessCheck{
			Name: "elasticsearch",
			Check: func(ctx context.Context) error {
				return postsIdx.Ready(ctx)
			},
		})
	}
	server.ConfigureEngine(engine, logger, db, version, readinessChecks...)

	// Register APIs under an optional base path (e.g. /v1).
	api := engine.Group(appCfg.APIBasePath)
	// Rate limiting is grouped by API category to avoid coupling different
	// traffic types (public browsing vs auth vs admin writes).
	publicMax := rateCfg.PublicMaxPerMinute
	authMax := rateCfg.AuthMaxPerMinute
	adminMax := rateCfg.AdminMaxPerMinute
	if !rateCfg.Enabled {
		publicMax = 0
		authMax = 0
		adminMax = 0
	}

	type rateLimiter interface {
		Middleware(scope string) gin.HandlerFunc
	}
	var publicLimiter rateLimiter = platformMiddleware.NewFixedWindowRateLimiter(publicMax, rateCfg.Window)
	var authLimiter rateLimiter = platformMiddleware.NewFixedWindowRateLimiter(authMax, rateCfg.Window)
	var adminLimiter rateLimiter = platformMiddleware.NewFixedWindowRateLimiter(adminMax, rateCfg.Window)
	if rateCfg.RedisEnabled && strings.TrimSpace(rateCfg.RedisAddr) != "" {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     rateCfg.RedisAddr,
			Password: rateCfg.RedisPassword,
			DB:       rateCfg.RedisDB,
		})
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warnw("shared rate limit store unavailable; using in-memory limiter", "error", err)
		} else {
			counterStore := platformMiddleware.NewRedisCounterStore(redisClient, rateCfg.RedisPrefix)
			publicLimiter = platformMiddleware.NewSharedFixedWindowRateLimiter(publicMax, rateCfg.Window, counterStore)
			authLimiter = platformMiddleware.NewSharedFixedWindowRateLimiter(authMax, rateCfg.Window, counterStore)
			adminLimiter = platformMiddleware.NewSharedFixedWindowRateLimiter(adminMax, rateCfg.Window, counterStore)
			logger.Infow("shared rate limiting enabled", "backend", "redis")
		}
	}

	authGroup := api.Group("")
	authGroup.Use(authLimiter.Middleware("auth"))
	authHandler.RegisterRoutes(authGroup)

	// Public APIs (Phase 4): published content, no auth.
	publicGroup := api.Group("")
	publicGroup.Use(publicLimiter.Middleware("public"))
	postsHandler.RegisterPublicRoutes(publicGroup)
	pagesHandler.RegisterPublicRoutes(publicGroup)

	// Serve uploads in local/dev mode. In production, prefer Nginx for static files.
	if mediaCfg.PublicBaseURL != "" && mediaCfg.PublicBaseURL[0] == '/' {
		engine.StaticFS(mediaCfg.PublicBaseURL, gin.Dir(mediaCfg.StorageDir, false))
	}

	// Admin/content APIs (Phase 3-4): protected, used by admin clients and backends.
	admin := api.Group("/admin")
	// Rate limit is applied before auth so abuse without valid tokens is also
	// throttled.
	admin.Use(adminLimiter.Middleware("admin"), platformMiddleware.AuthRequired(jwtProvider, jwtCfg))

	postsHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider, jwtCfg),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	pagesHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider, jwtCfg),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	taxHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider, jwtCfg),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	mediaHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider, jwtCfg),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(permissionChecker, code) },
	)
	pluginsHandler.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(jwtProvider, jwtCfg),
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

	if graphqlCfg.Enabled {
		gqlSrv := gqlhandler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
			Resolvers: &gqlapi.Resolver{
				Auth:      authService,
				PostsCore: postsService.CorePostsService,
				Pages:     pagesService,
				Taxonomy:  taxService,
				Search:    postsIdx,
				JWT:       jwtCfg,
			},
		}))
		path := strings.TrimSpace(graphqlCfg.Path)
		if path == "" {
			path = appCfg.APIBasePath + "/graphql"
		}
		engine.POST(path, publicLimiter.Middleware("public"), func(c *gin.Context) {
			c.Request = c.Request.WithContext(gqlapi.WithGinContext(c.Request.Context(), c))
			gqlSrv.ServeHTTP(c.Writer, c.Request)
		})
		engine.GET(path, publicLimiter.Middleware("public"), func(c *gin.Context) {
			c.Request = c.Request.WithContext(gqlapi.WithGinContext(c.Request.Context(), c))
			gqlSrv.ServeHTTP(c.Writer, c.Request)
		})
		envLower := strings.ToLower(strings.TrimSpace(appCfg.Env))
		playgroundOK := graphqlCfg.PlaygroundEnabled && (envLower == "local" || envLower == "dev")
		if graphqlCfg.PlaygroundEnabled && !playgroundOK {
			logger.Warnw("GRAPHQL_PLAYGROUND_ENABLED ignored outside local/dev app environments",
				"app_env", appCfg.Env,
			)
		}
		if playgroundOK {
			engine.GET(path+"/playground", gin.WrapH(playground.Handler("GraphQL playground", path)))
		}
		logger.Infow("graphql endpoint enabled",
			"path", path,
			"playground", playgroundOK,
		)
	}

	// The Server holds the application configuration and shared dependencies
	// such as the database handle. Additional modules will be layered on top
	// of this container in subsequent phases.
	srv := server.NewServer(engine, appCfg, dbCfg, db, logger)

	// Run the HTTP server in its own goroutine so that we can listen for OS
	// signals and coordinate a controlled shutdown sequence.
	go func() {
		if err := srv.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
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

	logger.Infow("stopped cleanly", "service", appCfg.LogIdentifier)
}
