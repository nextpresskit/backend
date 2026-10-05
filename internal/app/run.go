package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/nextpresskit/backend/internal/config"
	"github.com/nextpresskit/backend/internal/kit"
	authinfra "github.com/nextpresskit/backend/internal/modules/auth/infrastructure"
	platformDatabase "github.com/nextpresskit/backend/internal/platform/database"
	platformES "github.com/nextpresskit/backend/internal/platform/elasticsearch"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
	"github.com/nextpresskit/backend/internal/server"
)

// Run boots the HTTP API: shared infra, module Prepare/Register*, background jobs, graceful shutdown.
// mods should come from kit.ResolveModulesFromRegistry(logger, appregistry.ModuleRegistry()).
func Run(ctx context.Context, version string, baseLogger *zap.Logger, mods []kit.Module) error {
	logger := baseLogger.Sugar()
	defer func() { _ = baseLogger.Sync() }()

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
	esCfg := config.LoadElasticsearchConfig(appCfg.Env)

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
		logger.Fatalw("failed to initialize database connection", "error", err)
	}

	esClient, err := platformES.NewClient(esCfg)
	if err != nil {
		logger.Fatalw("failed to create elasticsearch client", "error", err)
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
			logger.Fatalw("elasticsearch index setup failed", "error", err)
		}
	}

	if len(mods) == 0 {
		return fmt.Errorf("no modules resolved (check MODULES)")
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.MaxMultipartMemory = mediaCfg.MaxUploadBytes

	readinessChecks := make([]server.ReadinessCheck, 0, 1)
	if postsIdx != nil {
		readinessChecks = append(readinessChecks, server.ReadinessCheck{
			Name: "elasticsearch",
			Check: func(c context.Context) error {
				return postsIdx.Ready(c)
			},
		})
	}
	server.ConfigureEngine(engine, logger, db, version, readinessChecks...)

	api := engine.Group(appCfg.APIBasePath)
	publicMax := rateCfg.PublicMaxPerMinute
	authMax := rateCfg.AuthMaxPerMinute
	adminMax := rateCfg.AdminMaxPerMinute
	if !rateCfg.Enabled {
		publicMax = 0
		authMax = 0
		adminMax = 0
	}

	var publicLimiter kit.RateLimiter = platformMiddleware.NewFixedWindowRateLimiter(publicMax, rateCfg.Window)
	var authLimiter kit.RateLimiter = platformMiddleware.NewFixedWindowRateLimiter(authMax, rateCfg.Window)
	var adminLimiter kit.RateLimiter = platformMiddleware.NewFixedWindowRateLimiter(adminMax, rateCfg.Window)
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

	jwtProvider := authinfra.NewJWTProvider(jwtCfg.Secret, jwtCfg.AccessTTL, jwtCfg.RefreshTTL)

	authGroup := api.Group("")
	authGroup.Use(authLimiter.Middleware("auth"))
	publicGroup := api.Group("")
	publicGroup.Use(publicLimiter.Middleware("public"))
	admin := api.Group("/admin")
	admin.Use(adminLimiter.Middleware("admin"), platformMiddleware.AuthRequired(jwtProvider, jwtCfg))

	d := &kit.Deps{
		Ctx:           ctx,
		Log:           logger,
		DB:            db,
		AppCfg:        appCfg,
		JWTCfg:        jwtCfg,
		RBACCfg:       rbacCfg,
		MediaCfg:      mediaCfg,
		RateCfg:       rateCfg,
		ESCfg:         esCfg,
		Engine:        engine,
		API:           api,
		Public:        publicGroup,
		Auth:          authGroup,
		Admin:         admin,
		PublicLimiter: publicLimiter,
		AuthLimiter:   authLimiter,
		AdminLimiter:  adminLimiter,
		Version:       version,
		JWTProvider:   jwtProvider,
		ESClient:      esClient,
		PostsIdx:      postsIdx,
	}

	for _, m := range mods {
		if err := m.Prepare(d); err != nil {
			return fmt.Errorf("prepare %s: %w", m.ID(), err)
		}
	}

	for _, m := range mods {
		if err := m.RegisterAuth(d); err != nil {
			return fmt.Errorf("register auth %s: %w", m.ID(), err)
		}
	}

	if err := registerPublicOrdered(mods, d); err != nil {
		return err
	}

	if err := registerAdminOrdered(mods, d); err != nil {
		return err
	}

	for _, m := range mods {
		if err := m.Start(ctx, d); err != nil {
			return fmt.Errorf("start %s: %w", m.ID(), err)
		}
	}

	srv := server.NewServer(engine, appCfg, dbCfg, db, logger)

	go func() {
		if err := srv.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			logger.Fatalw("http server exited with error", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Infow("received shutdown signal", "signal", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("graceful shutdown failed: %v", err)
	}

	logger.Infow("stopped cleanly", "service", appCfg.LogIdentifier)
	return nil
}

func registerPublicOrdered(mods []kit.Module, d *kit.Deps) error {
	order := []string{"posts", "pages", "media", "taxonomy"}
	seen := map[string]struct{}{}
	for _, id := range order {
		for _, m := range mods {
			if m.ID() == id {
				if err := m.RegisterPublic(d); err != nil {
					return fmt.Errorf("register public %s: %w", m.ID(), err)
				}
				seen[id] = struct{}{}
				break
			}
		}
	}
	for _, m := range mods {
		if _, ok := seen[m.ID()]; ok {
			continue
		}
		if err := m.RegisterPublic(d); err != nil {
			return fmt.Errorf("register public %s: %w", m.ID(), err)
		}
	}
	return nil
}

func registerAdminOrdered(mods []kit.Module, d *kit.Deps) error {
	order := []string{"posts", "pages", "taxonomy", "media", "rbac", "user"}
	for _, id := range order {
		for _, m := range mods {
			if m.ID() == id {
				if err := m.RegisterAdmin(d); err != nil {
					return fmt.Errorf("register admin %s: %w", m.ID(), err)
				}
				break
			}
		}
	}
	return nil
}
