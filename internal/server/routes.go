package server

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thediligencedev/betteridn/internal/auth"
)

func RegisterRoutes(mux *http.ServeMux, pool *pgxpool.Pool, sessionManager *scs.SessionManager, googleHandler *auth.GoogleHandler, confirmHandler *auth.Handler) {
	// confirmHandler is the same as authHandler but includes ConfirmationService usage
	// or you might unify them. Shown separate for clarity.

	// POST /api/v1/auth/signup
	mux.Handle("/api/v1/auth/signup",
		Chain(http.HandlerFunc(confirmHandler.SignUp), Logger, CORS),
	)

	// POST /api/v1/auth/signin
	mux.Handle("/api/v1/auth/signin",
		Chain(http.HandlerFunc(confirmHandler.SignIn), Logger, CORS),
	)

	// POST /api/v1/auth/signout
	mux.Handle("/api/v1/auth/signout",
		Chain(http.HandlerFunc(confirmHandler.SignOut), Logger, CORS, WithAuth(sessionManager)),
	)

	// GET /api/v1/auth/session
	mux.Handle("/api/v1/auth/session",
		Chain(http.HandlerFunc(confirmHandler.GetCurrentSession), Logger, CORS, WithAuth(sessionManager)),
	)

	// Google OAuth endpoints
	mux.Handle("/api/v1/auth/google/login",
		Chain(http.HandlerFunc(googleHandler.GoogleLogin), Logger, CORS),
	)
	mux.Handle("/api/v1/auth/google/callback",
		Chain(http.HandlerFunc(googleHandler.GoogleCallback), Logger, CORS),
	)

	// GET /api/v1/auth/confirm-email?token=xxx
	mux.Handle("/api/v1/auth/confirm-email",
		Chain(http.HandlerFunc(confirmHandler.ConfirmEmail), Logger, CORS),
	)

	// POST /api/v1/auth/resend-confirmation
	mux.Handle("/api/v1/auth/resend-confirmation",
		Chain(http.HandlerFunc(confirmHandler.ResendConfirmation), Logger, CORS, WithAuth(sessionManager)),
	)

	MountSwaggerDocs(mux)

	// Protected example
	mux.Handle("/home",
		Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome Home! You are logged in."))
		}), Logger, CORS, WithAuth(sessionManager)),
	)
}
