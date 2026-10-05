package module

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/nextpresskit/backend/internal/kit"
	pagesapp "github.com/nextpresskit/backend/internal/modules/pages/application"
	pagesinfra "github.com/nextpresskit/backend/internal/modules/pages/infrastructure"
	pagesp "github.com/nextpresskit/backend/internal/modules/pages/persistence"
	pagestransport "github.com/nextpresskit/backend/internal/modules/pages/transport"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
	"gorm.io/gorm"
)

type pagesMod struct {
	handler *pagestransport.Handler
}

func (m *pagesMod) ID() string { return "pages" }

func (m *pagesMod) Prepare(d *kit.Deps) error {
	repo := pagesinfra.NewGormRepository(d.DB)
	svc := pagesapp.NewService(repo)
	m.handler = pagestransport.NewHandler(svc)
	return nil
}

func (m *pagesMod) RegisterAuth(*kit.Deps) error { return nil }

func (m *pagesMod) RegisterPublic(d *kit.Deps) error {
	m.handler.RegisterPublicRoutes(d.Public)
	return nil
}

func (m *pagesMod) RegisterAdmin(d *kit.Deps) error {
	m.handler.RegisterRoutes(
		d.Admin,
		platformMiddleware.AuthRequired(d.JWTProvider, d.JWTCfg),
		func(code string) gin.HandlerFunc {
			return platformMiddleware.RequirePermission(d.PermissionChecker, code)
		},
	)
	return nil
}

func (m *pagesMod) AutoMigrate(db *gorm.DB) error {
	return pagesp.AutoMigrate(db)
}

func (m *pagesMod) Seed(db *gorm.DB, _ kit.SeedOpts) error {
	return pagesp.SeedDemo(db)
}

func (m *pagesMod) Start(context.Context, *kit.Deps) error { return nil }

func (m *pagesMod) Permissions() []string {
	return []string{"pages:read", "pages:write"}
}

// Module is the pages slice.
var Module kit.Module = new(pagesMod)
