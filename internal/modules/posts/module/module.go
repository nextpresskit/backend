package module

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nextpresskit/backend/internal/kit"
	postsapp "github.com/nextpresskit/backend/internal/modules/posts/application"
	postsIdent "github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	postPorts "github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
	postsinfra "github.com/nextpresskit/backend/internal/modules/posts/infrastructure"
	postsp "github.com/nextpresskit/backend/internal/modules/posts/persistence"
	poststransport "github.com/nextpresskit/backend/internal/modules/posts/transport"
	platformES "github.com/nextpresskit/backend/internal/platform/elasticsearch"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type postsMod struct {
	handler *poststransport.Handler
}

func (m *postsMod) ID() string { return "posts" }

func (m *postsMod) Prepare(d *kit.Deps) error {
	repo := postsinfra.NewRepositoryAdapter(postsinfra.NewGormRepository(d.DB))
	derivedHook := postsapp.NewDerivedFieldsHook(repo)
	esHook := platformES.NewPostIndexHook(d.Log, d.PostsIdx, repo)
	hooks := postsapp.NewPostSaveChain(derivedHook)
	if esHook != nil {
		hooks = postsapp.NewPostSaveChain(derivedHook, esHook)
	}
	svc := postsapp.NewService(repo, hooks)
	m.handler = poststransport.NewHandlerFromServiceWithOptionalSearch(svc, d.PostsIdx)
	d.PostsRepo = repo
	return nil
}

func (m *postsMod) RegisterAuth(*kit.Deps) error { return nil }

func (m *postsMod) RegisterPublic(d *kit.Deps) error {
	m.handler.RegisterPublicRoutes(d.Public)
	return nil
}

func (m *postsMod) RegisterAdmin(d *kit.Deps) error {
	m.handler.RegisterRoutes(
		d.Admin,
		platformMiddleware.AuthRequired(d.JWTProvider, d.JWTCfg),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(d.PermissionChecker, code) },
	)
	return nil
}

func (m *postsMod) AutoMigrate(db *gorm.DB) error {
	return postsp.AutoMigrate(db)
}

func (m *postsMod) Seed(db *gorm.DB, _ kit.SeedOpts) error {
	return postsp.SeedDemo(db)
}

func (m *postsMod) Start(ctx context.Context, d *kit.Deps) error {
	if d.PostsRepo == nil {
		return nil
	}
	go runScheduledPublishLoop(ctx, d.DB, d.Log, d.PostsIdx, d.PostsRepo)
	return nil
}

func (m *postsMod) Permissions() []string {
	return []string{"posts:read", "posts:write"}
}

func runScheduledPublishLoop(ctx context.Context, db *gorm.DB, logger *zap.SugaredLogger, postsIdx *platformES.PostsIndex, postsRepo postPorts.Repository) {
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
					var id int64
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
}

// Module is the posts slice.
var Module kit.Module = new(postsMod)
