package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/kk-Syuer/wasatext_2024/service"
)

type ctxKeyUsername struct{}

func AuthMiddleware(svc service.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1) Allow the login endpoint through without a token
			if r.Method == http.MethodPost && r.URL.Path == "/session" {
				next.ServeHTTP(w, r)
				return
			}

			// 2) Check for Bearer token
			auth := r.Header.Get("Authorization")
			parts := strings.Fields(auth)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 3) Validate it
			user, err := svc.Validate(r.Context(), parts[1])
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 4) Inject the username (or user struct) into context
			ctx := context.WithValue(r.Context(), ctxKeyUsername{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UsernameFromContext(ctx context.Context) string {
	if u, _ := ctx.Value(ctxKeyUsername{}).(string); u != "" {
		return u
	}
	return ""
}
