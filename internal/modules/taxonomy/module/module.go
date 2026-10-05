package module

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/nextpresskit/backend/internal/kit"
	taxapp "github.com/nextpresskit/backend/internal/modules/taxonomy/application"
	taxinfra "github.com/nextpresskit/backend/internal/modules/taxonomy/infrastructure"
	taxp "github.com/nextpresskit/backend/internal/modules/taxonomy/persistence"
	taxtransport "github.com/nextpresskit/backend/internal/modules/taxonomy/transport"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
	"gorm.io/gorm"
)

type taxMod struct {
	handler *taxtransport.Handler
}

func (m *taxMod) ID() string { return "taxonomy" }

func (m *taxMod) Prepare(d *kit.Deps) error {
	repo := taxinfra.NewGormRepository(d.DB)
	svc := taxapp.NewService(repo)
	m.handler = taxtransport.NewHandler(svc)
	return nil
}

func (m *taxMod) RegisterAuth(*kit.Deps) error   { return nil }
func (m *taxMod) RegisterPublic(*kit.Deps) error { return nil }

func (m *taxMod) RegisterAdmin(d *kit.Deps) error {
	m.handler.RegisterRoutes(
		d.Admin,
		platformMiddleware.AuthRequired(d.JWTProvider, d.JWTCfg),
		func(code string) gin.HandlerFunc {
			return platformMiddleware.RequirePermission(d.PermissionChecker, code)
		},
	)
	return nil
}

func (m *taxMod) AutoMigrate(db *gorm.DB) error {
	return taxp.AutoMigrate(db)
}

func (m *taxMod) Seed(db *gorm.DB, _ kit.SeedOpts) error {
	return taxp.SeedDemo(db)
}

func (m *taxMod) Start(context.Context, *kit.Deps) error { return nil }

func (m *taxMod) Permissions() []string {
	return []string{
		"categories:read", "categories:write",
		"tags:read", "tags:write",
	}
}

// Module is the taxonomy slice.
var Module kit.Module = new(taxMod)
