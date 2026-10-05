package module

import (
	"context"

	"github.com/nextpresskit/backend/internal/kit"
	authapp "github.com/nextpresskit/backend/internal/modules/auth/application"
	authinfra "github.com/nextpresskit/backend/internal/modules/auth/infrastructure"
	authtransport "github.com/nextpresskit/backend/internal/modules/auth/transport"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
	"gorm.io/gorm"
)

type authMod struct{}

func (authMod) ID() string { return "auth" }

func (authMod) Prepare(d *kit.Deps) error {
	passwordHasher := authinfra.NewBcryptHasher(0)
	// d.JWTProvider is created in app.Run before Prepare (shared with admin middleware).
	d.AuthService = authapp.NewService(d.UserRepo, d.JWTProvider, passwordHasher)
	d.AuthService.SetRBACReader(d.RBACRepo)
	d.AuthHandler = authtransport.NewHandler(
		d.AuthService,
		platformMiddleware.AuthRequired(d.JWTProvider, d.JWTCfg),
		d.JWTCfg,
	)
	return nil
}

func (authMod) RegisterAuth(d *kit.Deps) error {
	d.AuthHandler.RegisterRoutes(d.Auth)
	return nil
}

func (authMod) RegisterPublic(*kit.Deps) error { return nil }
func (authMod) RegisterAdmin(*kit.Deps) error  { return nil }

func (authMod) AutoMigrate(*gorm.DB) error { return nil }

func (authMod) Seed(*gorm.DB, kit.SeedOpts) error { return nil }

func (authMod) Start(context.Context, *kit.Deps) error { return nil }

func (authMod) Permissions() []string { return nil }

// Module is the auth slice (JWT + HTTP).
var Module kit.Module = authMod{}
