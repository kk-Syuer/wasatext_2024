/*
Package api exposes the main API engine. All HTTP APIs are handled here - so-called "business logic" should be here, or
in a dedicated package (if that logic is complex enough).

To use this package, you should create a new instance with New() passing a valid Config. The resulting Router will have
the Router.Handler() function that returns a handler that can be used in a http.Server (or in other middlewares).

Example:

	// Create the API router
	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: appdb,
	})
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("error creating the API server instance: %w", err)
	}
	router := apirouter.Handler()

	// ... other stuff here, like middleware chaining, etc.

	// Create the API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests in a separate goroutine
	apiserver.ListenAndServe()

See the `main.go` file inside the `cmd/webapi` for a full usage example.
*/
package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kk-Syuer/wasatext_2024/service"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/sirupsen/logrus"
)

// Config is used to provide dependencies and configuration to the New function.
type Config struct {
	// Logger where log entries are sent
	Logger logrus.FieldLogger

	// Database is the instance of *database.AppDatabase where data are saved
	Database *database.AppDatabase
}

// Router is the package API interface representing an API handler builder
type Router interface {
	// Handler returns an HTTP handler for APIs provided in this package
	Handler() http.Handler

	// Close terminates any resource used in the package
	Close() error
}

// New returns a new Router instance, wires up all your endpoints.
func New(cfg Config) (Router, error) {
	// Validate configuration
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}

	// Create underlying httprouter
	r := httprouter.New()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// Instantiate business-logic services
	sessionSvc := service.NewSessionService(cfg.Database)
	userSvc := service.NewUserService(cfg.Database)
	// TODO: instantiate ConversationService, MessageService, GroupService

	// Instantiate HTTP handlers
	sessH := NewSessionHandler(sessionSvc)
	userH := NewUserHandler(userSvc)
	// TODO: NewConversationHandler, NewMessageHandler, NewGroupHandler

	// Register session endpoints
	r.POST("/session", adapter(sessH.DoLogin))

	// Register user endpoints
	r.GET("/users", adapter(userH.ListUsers))
	r.GET("/users/:username", wrap(userH.GetUser))
	r.PATCH("/users/:username/name", wrap(userH.UpdateName))
	r.PUT("/users/:username/photo", wrap(userH.UpdatePhoto))

	// TODO: register conversation, message, group routes

	return &_router{
		router:     r,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}, nil
}

type _router struct {
	router     *httprouter.Router
	baseLogger logrus.FieldLogger
	db         *database.AppDatabase
}

func (r *_router) Handler() http.Handler {
	return r.router
}

func (r *_router) Close() error {
	// Nothing to clean up for now
	return nil
}

// adapter converts a standard http.HandlerFunc into a httprouter.Handle,
// ignoring URL parameters.
func adapter(fn func(http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		fn(w, req)
	}
}

// wrap converts an http.HandlerFunc into a httprouter.Handle,
// injecting URL params into the request context.
func wrap(fn func(http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(req.Context(), httprouter.ParamsKey, ps)
		fn(w, req.WithContext(ctx))
	}
}
