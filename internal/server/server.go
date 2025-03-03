package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/config"
	"github.com/thediligencedev/betteridn/internal/worker"
)

type Server struct {
	pool           *pgxpool.Pool
	cfg            *config.Config
	sessionManager *scs.SessionManager
	httpServer     *http.Server
	emailWorker    *worker.EmailWorker
}

func New(pool *pgxpool.Pool, cfg *config.Config) *Server {
	// Initialize session manager
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(pool)
	sessionManager.Lifetime = cfg.SessionExpiry
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = false
	sessionManager.Cookie.Path = "/"

	// Initialize email worker
	emailWorker := worker.NewEmailWorker(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPFrom,
		cfg.SMTPUser,
		cfg.SMTPPass,
	)

	s := &Server{
		pool:           pool,
		cfg:            cfg,
		sessionManager: sessionManager,
		emailWorker:    emailWorker,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	handler := sessionManager.LoadAndSave(mux)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: handler,
	}
	return s
}

func (s *Server) Start() error {
	log.Printf("Starting HTTP server on port %s", s.cfg.ServerPort)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Stopping HTTP server...")
	s.emailWorker.Close()
	return s.httpServer.Shutdown(ctx)
}

// package server

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/alexedwards/scs/pgxstore"
// 	"github.com/alexedwards/scs/v2"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/thediligencedev/betteridn/internal/auth"
// 	"github.com/thediligencedev/betteridn/internal/config"
// 	"github.com/thediligencedev/betteridn/internal/worker"
// )

// type Server struct {
// 	pool           *pgxpool.Pool
// 	port           string
// 	sessionManager *scs.SessionManager
// 	httpServer     *http.Server
// 	emailWorker    *worker.EmailWorker
// }

// func New(pool *pgxpool.Pool, cfg *config.Config) *Server {
// 	// 1. Initialize session manager
// 	sessionManager := scs.New()
// 	sessionManager.Store = pgxstore.New(pool)
// 	sessionManager.Lifetime = cfg.SessionExpiry
// 	sessionManager.Cookie.HttpOnly = true
// 	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
// 	sessionManager.Cookie.Secure = false
// 	sessionManager.Cookie.Path = "/"

// 	// 2. Initialize email worker (SMTP config from env)
// 	emailWorker := worker.NewEmailWorker(
// 		cfg.SMTPHost,
// 		cfg.SMTPPort,
// 		cfg.SMTPFrom,
// 		cfg.SMTPUser,
// 		cfg.SMTPPass,
// 	)

// 	// 3. Initialize ConfirmationService using pgxpool.Pool
// 	confirmationService := auth.NewConfirmationService(pool, emailWorker)

// 	// 4. Initialize Google OAuth
// 	auth.InitGoogleOAuth(cfg)
// 	// Pass the confirmationService as 4th param
// 	googleHandler := auth.NewGoogleHandler(pool, sessionManager, cfg)

// 	// 5. Build your main AuthHandler with the same confirmationService
// 	authHandler := auth.NewHandler(pool, sessionManager, confirmationService)

// 	mux := http.NewServeMux()
// 	s := &Server{
// 		pool:           pool,
// 		port:           cfg.ServerPort,
// 		sessionManager: sessionManager,
// 		emailWorker:    emailWorker,
// 	}

// 	// 6. Register your routes
// 	RegisterRoutes(mux, pool, sessionManager, googleHandler, authHandler)

// 	handler := sessionManager.LoadAndSave(mux)

// 	s.httpServer = &http.Server{
// 		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
// 		Handler: handler,
// 	}
// 	return s
// }

// func (s *Server) Start() error {
// 	log.Printf("Starting HTTP server on port %s", s.port)
// 	return s.httpServer.ListenAndServe()
// }

// func (s *Server) Shutdown(ctx context.Context) error {
// 	log.Println("Stopping HTTP server...")
// 	// close the email worker
// 	s.emailWorker.Close()
// 	return s.httpServer.Shutdown(ctx)
// }
