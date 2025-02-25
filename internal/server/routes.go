package server

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/auth"
)

// RegisterRoutes attaches each endpoint to the mux with clear method/endpoint info.
func RegisterRoutes(mux *http.ServeMux, pool *pgxpool.Pool, sessionManager *scs.SessionManager, googleHandler *auth.GoogleHandler) {
	authHandler := auth.NewHandler(pool, sessionManager)

	// POST /api/v1/auth/signup
	mux.Handle("/api/v1/auth/signup",
		Chain(http.HandlerFunc(authHandler.SignUp), Logger, CORS),
	)

	// POST /api/v1/auth/signin
	mux.Handle("/api/v1/auth/signin",
		Chain(http.HandlerFunc(authHandler.SignIn), Logger, CORS),
	)

	// POST /api/v1/auth/signout
	mux.Handle("/api/v1/auth/signout",
		Chain(http.HandlerFunc(authHandler.SignOut), Logger, WithAuth(sessionManager), CORS),
	)

	// GET /api/v1/auth/session
	mux.Handle("/api/v1/auth/session",
		Chain(http.HandlerFunc(authHandler.GetCurrentSession), Logger, WithAuth(sessionManager), CORS),
	)

	// GET /api/v1/auth/google/login
	mux.Handle("/api/v1/auth/google/login",
		Chain(http.HandlerFunc(googleHandler.GoogleLogin), Logger, CORS),
	)

	// GET /api/v1/auth/google/callback
	mux.Handle("/api/v1/auth/google/callback",
		Chain(http.HandlerFunc(googleHandler.GoogleCallback), Logger, CORS),
	)

	// Protected Example: GET /home
	mux.Handle("/home",
		Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome Home! You are logged in."))
		}), Logger, CORS, WithAuth(sessionManager)),
	)
}
