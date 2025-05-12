package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kk-Syuer/wasatext_2024/service"
)

// NewRouter wires up all your HTTP endpoints and returns a configured router.
// Right now we only register the /session route; add your others (users, messages…)
// by following the same pattern.
func NewRouter(sessionSvc service.SessionService) *httprouter.Router {
	router := httprouter.New()

	// Session endpoint
	sessionH := NewSessionHandler(sessionSvc)
	router.POST("/session", adapter(sessionH.DoLogin))

	// TODO: add more routes here
	// userH := NewUserHandler(userSvc)
	// router.GET("/users/:id", adapter(userH.GetProfile))
	// ...

	return router
}

// adapter converts a standard http.HandlerFunc into a httprouter.Handle,
// ignoring URL parameters.
func adapter(fn func(http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fn(w, r)
	}
}
