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

// AuthMiddleware returns an http.Handler that enforces Bearer <username> authentication.
func AuthMiddleware(next http.Handler, sessSvc service.SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))

		// Validate: in this simplified model, the token == username
		if _, err := sessSvc.Validate(r.Context(), username); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), ctxKeyUsername{}, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
