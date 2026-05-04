package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nextpresskit/backend/internal/config"
)

// Server wraps the underlying http.Server and carries the core dependencies
// that are required at the edge of the system (HTTP).
//
// Phase 1 only wires the HTTP layer and configuration. The global DB instance
// and module registration will be added in later phases without changing this
// abstraction.
type Server struct {
	engine *gin.Engine
	http   *http.Server

	appCfg config.AppConfig
	dbCfg  config.DBConfig

	db *gorm.DB

	log *zap.SugaredLogger
}

// NewServer constructs a new Server instance. The constructor is intentionally
// explicit about its dependencies to keep wiring straightforward and avoid any
// hidden singletons beyond the planned global DB instance.
func NewServer(
	engine *gin.Engine,
	appCfg config.AppConfig,
	dbCfg config.DBConfig,
	db *gorm.DB,
	log *zap.SugaredLogger,
) *Server {
	addr := fmt.Sprintf(":%s", appCfg.Port)

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	return &Server{
		engine: engine,
		http:   httpSrv,
		appCfg: appCfg,
		dbCfg:  dbCfg,
		db:     db,
		log:    log,
	}
}

// Start launches the HTTP server and returns only when the listener fails.
// It is expected to be called from a goroutine so that the caller can control
// the shutdown sequence.
func (s *Server) Start() error {
	s.log.Infow("http server starting",
		"addr", s.http.Addr,
		"env", s.appCfg.Env,
	)
	return s.http.ListenAndServe()
}

// Shutdown attempts a graceful shutdown of the HTTP server. A bounded timeout
// should be applied to the context by the caller to avoid hanging indefinitely.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Infow("http server shutting down")

	// Give Gin a moment to complete in-flight requests.
	shutdownErr := s.http.Shutdown(ctx)
	if shutdownErr != nil {
		s.log.Errorw("http server shutdown error", "error", shutdownErr)
		return shutdownErr
	}

	// Optionally wait for a short period so any background cleanup in handlers
	// can complete. This is intentionally conservative and can be tuned later.
	select {
	case <-ctx.Done():
		// Context deadline exceeded or cancelled; nothing more to wait for.
	default:
		time.Sleep(100 * time.Millisecond)
	}

	s.log.Infow("http server shutdown complete")
	return nil
}

