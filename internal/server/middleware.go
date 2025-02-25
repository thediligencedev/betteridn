package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

type Middleware func(http.Handler) http.Handler

// Chain applies a list of middlewares in order.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// TODO: fix the origin, allow methods, and allow headers as needed
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For development only
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, hx-current-url, hx-request, hx-target, Accept, Content-Length, Accept-Encoding, Accept-Language")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// WithAuth checks if a user is authenticated by checking session data.
// If no user is found, it returns 401 Unauthorized.
func WithAuth(sessionManager *scs.SessionManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If preflight request, pass through
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			userID := sessionManager.GetString(r.Context(), "user_id")
			if userID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				resp := map[string]string{"message": "unauthorized"}
				_ = json.NewEncoder(w).Encode(resp)
				fmt.Println("dasdsa")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
