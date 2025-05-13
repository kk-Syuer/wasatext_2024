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
			auth := r.Header.Get("Authorization")
			parts := strings.Fields(auth)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				if user, err := svc.Validate(r.Context(), parts[1]); err == nil {
					ctx := context.WithValue(r.Context(), ctxKeyUsername{}, user)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UsernameFromContext(ctx context.Context) string {
	if u, _ := ctx.Value(ctxKeyUsername{}).(string); u != "" {
		return u
	}
	return ""
}
