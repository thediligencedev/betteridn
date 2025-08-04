// server/middleware.go
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/thediligencedev/betteridn/internal/config"
	"github.com/thediligencedev/betteridn/internal/models"
	"github.com/thediligencedev/betteridn/pkg/response"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

func Logger(sessionManager *scs.SessionManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusResponseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			userID := sessionManager.GetString(r.Context(), "user_id")
			if userID == "" {
				userID = "unauthenticated"
			}

			logger := slog.Default().With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Float64("duration_ms", float64(time.Since(start).Nanoseconds())/1e6),
				slog.String("user_id", userID),
				slog.String("ip", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			logger.Info("HTTP request")
		})
	}
}

func WithAuth(sessionManager *scs.SessionManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from the session
			userID := sessionManager.GetString(r.Context(), "user_id")
			slog.Info("Session user ID", slog.String("user_id", userID))

			if userID == "" {
				response.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			// Add user ID to the context
			ctx := context.WithValue(r.Context(), models.UserContextKey, userID)
			r = r.WithContext(ctx)

			// Pass control to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

func Optional(sessionManager *scs.SessionManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Session is already loaded by scs middleware
			// This middleware is for semantic purposes only. no-op middleware
			next.ServeHTTP(w, r)
		})
	}
}

func CORS(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use frontend URL from config, fallback to localhost:5173
			allowOrigin := cfg.FrontendURL
			if allowOrigin == "" {
				allowOrigin = "http://localhost:5173"
			}

			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, hx-current-url, hx-request, hx-target, hx-trigger, Accept, Content-Length, Accept-Encoding, Accept-Language, Credentials")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
