package server

import (
	"net/http"

	"github.com/thediligencedev/betteridn/internal/auth"
	"github.com/thediligencedev/betteridn/internal/post"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Initialize Google OAuth
	auth.InitGoogleOAuth(s.cfg)

	// Initialize handlers
	confirmationService := auth.NewConfirmationService(s.pool, s.emailWorker)
	googleHandler := auth.NewGoogleHandler(s.pool, s.sessionManager, s.cfg)

	authHandler := auth.NewHandler(s.pool, s.sessionManager, confirmationService)
	postHandler := post.NewHandler(s.pool)

	// Middleware stacks
	public := []Middleware{Logger(s.sessionManager), CORS}
	protected := []Middleware{Logger(s.sessionManager), WithAuth(s.sessionManager), CORS}
	optional := []Middleware{Logger(s.sessionManager), Optional(s.sessionManager), CORS}

	// Map to track registered OPTIONS patterns
	registeredOptions := make(map[string]bool)

	register := func(method, pattern string, h http.Handler, mw []Middleware) {
		// Register the actual method (e.g., POST, GET)
		mux.Handle(method+" "+pattern, Chain(h, mw...))

		// Automatically register an OPTIONS handler for the same pattern
		if method != http.MethodOptions && !registeredOptions[pattern] {
			mux.Handle("OPTIONS "+pattern, Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}), mw...))
			registeredOptions[pattern] = true // Mark as registered
		}
	}

	// Auth routes
	register("POST", "/api/v1/auth/signup", http.HandlerFunc(authHandler.SignUp), public)
	register("POST", "/api/v1/auth/signin", http.HandlerFunc(authHandler.SignIn), public)
	register("POST", "/api/v1/auth/signout", http.HandlerFunc(authHandler.SignOut), protected)
	register("GET", "/api/v1/auth/session", http.HandlerFunc(authHandler.GetCurrentSession), protected)
	register("GET", "/api/v1/auth/google/login", http.HandlerFunc(googleHandler.GoogleLogin), public)
	register("GET", "/api/v1/auth/google/callback", http.HandlerFunc(googleHandler.GoogleCallback), public)
	register("GET", "/api/v1/auth/confirm-email", http.HandlerFunc(authHandler.ConfirmEmail), public)
	register("POST", "/api/v1/auth/resend-confirmation", http.HandlerFunc(authHandler.ResendConfirmation), protected)

	// Post routes
	register("POST", "/api/v1/posts", http.HandlerFunc(postHandler.CreatePost), protected)
	register("GET", "/api/v1/posts", http.HandlerFunc(postHandler.GetPosts), optional)
	register("GET", "/api/v1/posts/{postId}", http.HandlerFunc(postHandler.GetPostByID), optional)
	register("PUT", "/api/v1/posts/{postId}", http.HandlerFunc(postHandler.UpdatePost), protected)
	register("POST", "/api/v1/posts/{postId}/vote", http.HandlerFunc(postHandler.VotePost), protected)

	MountSwaggerDocs(mux)

	// Protected example
	register("GET", "/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome Home! You are logged in."))
	}), protected)
}

// // server/routes.go
// package server

// import (
// 	"net/http"

// 	"github.com/alexedwards/scs/v2"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/thediligencedev/betteridn/internal/auth"
// 	"github.com/thediligencedev/betteridn/internal/post"
// )

// func RegisterRoutes(mux *http.ServeMux, pool *pgxpool.Pool, sessionManager *scs.SessionManager, googleHandler *auth.GoogleHandler, confirmHandler *auth.Handler) {
// 	postHandler := post.NewHandler(pool)

// 	// Middleware stacks
// 	public := []Middleware{Logger(sessionManager), CORS}
// 	protected := []Middleware{Logger(sessionManager), CORS, WithAuth(sessionManager)}
// 	optional := []Middleware{Logger(sessionManager), CORS, Optional(sessionManager)}

// 	register := func(method, pattern string, h http.Handler, mw []Middleware) {
// 		mux.Handle(method+" "+pattern, Chain(h, mw...))
// 	}

// 	// Auth routes
// 	register("POST", "/api/v1/auth/signup", http.HandlerFunc(confirmHandler.SignUp), public)
// 	register("POST", "/api/v1/auth/signin", http.HandlerFunc(confirmHandler.SignIn), public)
// 	register("POST", "/api/v1/auth/signout", http.HandlerFunc(confirmHandler.SignOut), protected)
// 	register("GET", "/api/v1/auth/session", http.HandlerFunc(confirmHandler.GetCurrentSession), protected)
// 	register("GET", "/api/v1/auth/google/login", http.HandlerFunc(googleHandler.GoogleLogin), public)
// 	register("GET", "/api/v1/auth/google/callback", http.HandlerFunc(googleHandler.GoogleCallback), public)
// 	register("POST", "/api/v1/auth/confirm-email", http.HandlerFunc(confirmHandler.ConfirmEmail), public)
// 	register("POST", "/api/v1/auth/resend-confirmation", http.HandlerFunc(confirmHandler.ResendConfirmation), protected)

// 	// Post routes
// 	register("POST", "/api/v1/posts", http.HandlerFunc(postHandler.CreatePost), protected)
// 	register("GET", "/api/v1/posts", http.HandlerFunc(postHandler.GetPosts), optional)
// 	register("GET", "/api/v1/posts/{postId}", http.HandlerFunc(postHandler.GetPostByID), optional)
// 	register("PUT", "/api/v1/posts/{postId}", http.HandlerFunc(postHandler.UpdatePost), protected)
// 	register("POST", "/api/v1/posts/{postId}/vote", http.HandlerFunc(postHandler.VotePost), protected)

// 	MountSwaggerDocs(mux)

// 	// Protected example
// 	register("GET", "/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("Welcome Home! You are logged in."))
// 	}), protected)
// }
