package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/kk-Syuer/wasatext_2024/service"
)

// ctxKeyUsername is used to put/get the username in request context
type ctxKeyUsername struct{}

// UsernameFromContext retrieves the username stored in context
func UsernameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyUsername{}).(string); ok {
		return v
	}
	return ""
}

// AuthMiddleware enforces Bearer auth, but lets /session & OPTIONS pass through.
func AuthMiddleware(next http.Handler, sessSvc service.SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1) Always allow CORS preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 2) Allow login endpoint without Authorization
		//    Normalize trailing slash just in case.
		path := strings.TrimSuffix(r.URL.Path, "/")
		if r.Method == http.MethodPost && path == "/session" {
			next.ServeHTTP(w, r)
			return
		}

		// 3) Public: static uploads (GET/HEAD)
		if (r.Method == http.MethodGet || r.Method == http.MethodHead) &&
			strings.HasPrefix(path, "/uploads/") {
			// Allow images to be embedded from other ports (e.g., Vite dev server)
			w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
			next.ServeHTTP(w, r)
			return
		}

		// 4) Everything else: require Bearer <identifier> (username in this project)
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		identifier := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))

		// Validate identifier (can be username or UUID depending on your choice)
		if _, err := sessSvc.Validate(r.Context(), identifier); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Put the (validated) identifier into context as "username"
		ctx := context.WithValue(r.Context(), ctxKeyUsername{}, identifier)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
