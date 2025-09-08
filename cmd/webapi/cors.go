package main

import (
	"net/http"
)
// applyCORSHandler wraps the HTTP handler with permissive CORS for the dev UI.
// Allows Vite on http://localhost:5173 to call the API on :3000.
func applyCORSHandler(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")

        // Allow only the dev UI origin (safer than "*", and required if you ever use credentials).
        if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Vary", "Origin")
        }

        // Allow common methods and headers used by your API.
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
        // Only set this if you actually use cookies; it’s harmless to keep but never combine with "*" origin.
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        // Preflight: answer OPTIONS and stop here
        if r.Method == http.MethodOptions {
            // You can also set max age to reduce preflights (optional):
            // w.Header().Set("Access-Control-Max-Age", "600")
            w.WriteHeader(http.StatusNoContent)
            return
        }

        h.ServeHTTP(w, r)
    })
}

