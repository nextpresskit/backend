package module

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nextpresskit/backend/internal/kit"
	userApp "github.com/nextpresskit/backend/internal/modules/user/application"
	userdomain "github.com/nextpresskit/backend/internal/modules/user/domain"
	userInfra "github.com/nextpresskit/backend/internal/modules/user/infrastructure"
	userp "github.com/nextpresskit/backend/internal/modules/user/persistence"
	usertransport "github.com/nextpresskit/backend/internal/modules/user/transport"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type userMod struct{}

func (userMod) ID() string { return "user" }

func (userMod) Prepare(d *kit.Deps) error {
	d.UserRepo = userInfra.NewGormRepository(d.DB)
	return nil
}

func (userMod) RegisterAuth(*kit.Deps) error   { return nil }
func (userMod) RegisterPublic(*kit.Deps) error { return nil }

// RegisterAdmin mounts GET /admin/users and GET /admin/users/:id (permission users:read).
// It needs the RBAC permission checker; without RBAC the routes are not exposed.
func (userMod) RegisterAdmin(d *kit.Deps) error {
	if d.PermissionChecker == nil || d.Admin == nil {
		return nil
	}
	var roles userdomain.RoleReader
	if d.RBACRepo != nil {
		roles = d.RBACRepo
	}
	repo := userInfra.NewGormRepository(d.DB)
	svc := userApp.NewService(repo, repo, roles)
	usertransport.NewHandler(svc).RegisterRoutes(d.Admin, func(code string) gin.HandlerFunc {
		return platformMiddleware.RequirePermission(d.PermissionChecker, code)
	})
	return nil
}

func (userMod) AutoMigrate(db *gorm.DB) error {
	return userp.AutoMigrate(db)
}

func (userMod) Seed(db *gorm.DB, _ kit.SeedOpts) error {
	return userp.SeedDemo(db)
}

func (userMod) Start(context.Context, *kit.Deps) error { return nil }

func (userMod) Permissions() []string { return []string{usertransport.ReadPermission} }

// Module is the user slice (persistence, demo seed, admin read endpoints).
var Module kit.Module = userMod{}
