package kit

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nextpresskit/backend/internal/config"
	authapp "github.com/nextpresskit/backend/internal/modules/auth/application"
	authinfra "github.com/nextpresskit/backend/internal/modules/auth/infrastructure"
	authtransport "github.com/nextpresskit/backend/internal/modules/auth/transport"
	postPorts "github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
	rbacapp "github.com/nextpresskit/backend/internal/modules/rbac/application"
	rbacinfra "github.com/nextpresskit/backend/internal/modules/rbac/infrastructure"
	rbactransport "github.com/nextpresskit/backend/internal/modules/rbac/transport"
	userdomain "github.com/nextpresskit/backend/internal/modules/user/domain"
	platformES "github.com/nextpresskit/backend/internal/platform/elasticsearch"
)

// RateLimiter matches platform rate limit middleware.
type RateLimiter interface {
	Middleware(scope string) gin.HandlerFunc
}

// Deps carries shared infrastructure and wiring produced during Prepare for use in Register* and Start.
type Deps struct {
	Ctx context.Context
	Log *zap.SugaredLogger
	DB  *gorm.DB

	AppCfg   config.AppConfig
	JWTCfg   config.JWTConfig
	RBACCfg  config.RBACConfig
	MediaCfg config.MediaConfig
	RateCfg  config.RateLimitConfig
	ESCfg    config.ElasticsearchConfig

	Engine *gin.Engine
	API    *gin.RouterGroup
	Public *gin.RouterGroup
	Auth   *gin.RouterGroup
	Admin  *gin.RouterGroup

	PublicLimiter RateLimiter
	AuthLimiter   RateLimiter
	AdminLimiter  RateLimiter

	Version string

	JWTProvider       *authinfra.JWTProvider
	UserRepo          userdomain.Repository
	RBACRepo          *rbacinfra.GormRepository
	PermissionChecker *rbacinfra.GormPermissionChecker
	RBACService       *rbacapp.Service
	AuthService       *authapp.Service
	AuthHandler       *authtransport.Handler
	RBACHandler       *rbactransport.Handler

	ESClient  *elasticsearch.Client
	PostsIdx  *platformES.PostsIndex
	PostsRepo postPorts.Repository
}

// SeedOpts is passed to Module.Seed after RBAC default rows are upserted.
type SeedOpts struct {
	// DefaultPermissionCodes is the merged list from CollectPermissionCodes(modules).
	DefaultPermissionCodes []string
}

// Module is a composable feature slice: migrations, HTTP, seeds, background work.
type Module interface {
	ID() string
	Prepare(*Deps) error
	RegisterAuth(*Deps) error
	RegisterPublic(*Deps) error
	RegisterAdmin(*Deps) error
	AutoMigrate(*gorm.DB) error
	Seed(*gorm.DB, SeedOpts) error
	Start(context.Context, *Deps) error
	Permissions() []string
}
