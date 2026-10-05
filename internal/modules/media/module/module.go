package module

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/nextpresskit/backend/internal/kit"
	mediaapp "github.com/nextpresskit/backend/internal/modules/media/application"
	mediainfra "github.com/nextpresskit/backend/internal/modules/media/infrastructure"
	mediap "github.com/nextpresskit/backend/internal/modules/media/persistence"
	mediatransport "github.com/nextpresskit/backend/internal/modules/media/transport"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
	"gorm.io/gorm"
)

type mediaMod struct {
	handler *mediatransport.Handler
}

func (m *mediaMod) ID() string { return "media" }

func (m *mediaMod) Prepare(d *kit.Deps) error {
	repo := mediainfra.NewGormRepository(d.DB)
	storage := mediainfra.NewLocalStorage(d.MediaCfg.StorageDir, d.MediaCfg.PublicBaseURL, d.MediaCfg.MaxUploadBytes)
	svc := mediaapp.NewService(repo, storage)
	m.handler = mediatransport.NewHandler(svc)
	return nil
}

func (m *mediaMod) RegisterAuth(*kit.Deps) error { return nil }

func (m *mediaMod) RegisterPublic(d *kit.Deps) error {
	if d.MediaCfg.PublicBaseURL != "" && d.MediaCfg.PublicBaseURL[0] == '/' {
		d.Engine.StaticFS(d.MediaCfg.PublicBaseURL, gin.Dir(d.MediaCfg.StorageDir, false))
	}
	return nil
}

func (m *mediaMod) RegisterAdmin(d *kit.Deps) error {
	m.handler.RegisterRoutes(
		d.Admin,
		platformMiddleware.AuthRequired(d.JWTProvider, d.JWTCfg),
		func(code string) gin.HandlerFunc {
			return platformMiddleware.RequirePermission(d.PermissionChecker, code)
		},
	)
	return nil
}

func (m *mediaMod) AutoMigrate(db *gorm.DB) error {
	return mediap.AutoMigrate(db)
}

func (m *mediaMod) Seed(db *gorm.DB, _ kit.SeedOpts) error {
	return mediap.SeedDemo(db)
}

func (m *mediaMod) Start(context.Context, *kit.Deps) error { return nil }

func (m *mediaMod) Permissions() []string {
	return []string{"media:read", "media:write"}
}

// Module is the media slice.
var Module kit.Module = new(mediaMod)
