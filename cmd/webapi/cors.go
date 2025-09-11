package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler applies the API CORS policy.
// - Allow all origins ("*")  [required by the assignment]
// - Max-Age = 1              [required by the assignment]
// - Allow common methods used by the API
// - Allow the headers the frontend sends (JSON + auth)
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // do not change
		handlers.MaxAge(1),                     // do not change

		handlers.AllowedMethods([]string{
			"GET", "POST", "PATCH", "DELETE", "OPTIONS", "PUT",
		}),
		handlers.AllowedHeaders([]string{
			"Content-Type", "Authorization", "Accept", "X-Requested-With",
		}),
		// IMPORTANT: don't add handlers.AllowCredentials()
	)(h)
}
