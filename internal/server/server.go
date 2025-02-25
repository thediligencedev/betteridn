package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/auth"
	"github.com/thediligencedev/betteridn/internal/config"
)

type Server struct {
	pool           *pgxpool.Pool
	port           string
	sessionManager *scs.SessionManager
	httpServer     *http.Server
}

func New(pool *pgxpool.Pool, cfg *config.Config) *Server {
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(pool)
	sessionManager.Lifetime = cfg.SessionExpiry // e.g. 7 days
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	// For local dev, typically:
	sessionManager.Cookie.Secure = false
	sessionManager.Cookie.Path = "/"

	// For production, you might set:
	// sessionManager.Cookie.Secure = true
	// sessionManager.Cookie.SameSite = http.SameSiteStrictMode

	// Initialize the global Google OAuth config
	auth.InitGoogleOAuth(cfg)
	googleHandler := auth.NewGoogleHandler(pool, sessionManager, cfg)

	mux := http.NewServeMux()
	s := &Server{
		pool:           pool,
		port:           cfg.ServerPort,
		sessionManager: sessionManager,
	}

	RegisterRoutes(mux, pool, sessionManager, googleHandler)

	handler := sessionManager.LoadAndSave(mux)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: handler,
	}
	return s
}

func (s *Server) Start() error {
	log.Printf("Starting HTTP server on port %s", s.port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Stopping HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
