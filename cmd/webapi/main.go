/*
Webapi is the executable for the main web server.
It builds a web server around APIs from `service/api`.
Webapi connects to external resources needed (database) and starts two web servers: the API web server, and the debug.
Everything is served via the API web server, except debug variables (/debug/vars) and profiler infos (pprof).

Usage:

	webapi [flags]

Flags and configurations are handled automatically by the code in `load-configuration.go`.

Return values (exit codes):

	0
		The program ended successfully (no errors, stopped by signal)

	> 0
		The program ended due to an error

Note that this program will update the schema of the database to the latest version available (embedded in the
executable during the build).
*/
package main

import (
	"context"
	"database/sql"
	"errors"
	_ "expvar"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf"
	"github.com/julienschmidt/httprouter"
	"github.com/kk-Syuer/wasatext_2024/service"
	apipkg "github.com/kk-Syuer/wasatext_2024/service/api"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/kk-Syuer/wasatext_2024/service/globaltime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Seed the clock for any timestamp logic
	rand.Seed(globaltime.Now().UnixNano())

	// Load configuration and defaults
	cfg, err := loadConfiguration()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return err
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.Infof("application initializing")

	// Initialize database
	logger.Info("initializing database support")

	dbConn, err := sql.Open("sqlite3", cfg.DB.Filename)
	logger.Infof("using DB file: %s", cfg.DB.Filename)

	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = dbConn.Close()
	}()

	appDB, err := database.New(dbConn)
	if err != nil {
		logger.WithError(err).Error("error creating AppDatabase")
		return fmt.Errorf("creating AppDatabase: %w", err)
	}

	// Channels for shutdown and server errors
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	serverErrors := make(chan error, 1)

	// Start debug server (expvar + pprof) on cfg.Web.DebugHost
	go func() {
		logger.Infof("debug endpoints listening on %s", cfg.Web.DebugHost)
		serverErrors <- http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux)
	}()

	// Create API router
	logger.Info("initializing API server")

	//initiate all the handlers
	sessionSvc := service.NewSessionService(appDB)

	apiRouter, err := apipkg.New(apipkg.Config{
		Logger:   logger,
		Database: appDB, // or `db` if that’s your name
	})
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("creating the API server instance: %w", err)
	}

	// 1) grab the underlying httprouter.Router so we can mount public routes
	rawHandler := apiRouter.Handler()
	rtr, ok := rawHandler.(*httprouter.Router)
	if !ok {
		return fmt.Errorf("expected Handler() to be *httprouter.Router, got %T", rawHandler)
	}

	// 3) protect everything else
	protected := apipkg.AuthMiddleware(sessionSvc)(rtr)

	// 4) register web UI and apply CORS (once)
	protected, err = registerWebUI(protected)
	if err != nil {
		logger.WithError(err).Error("error registering web UI handler")
		return fmt.Errorf("registering web UI handler: %w", err)
	}
	protected = applyCORSHandler(protected)

	// 5) start the HTTP server with 'protected' as the handler
	apiServer := &http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           protected,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	go func() {
		logger.Infof("API listening on %s", apiServer.Addr)
		serverErrors <- apiServer.ListenAndServe()
		logger.Info("stopping API server")
	}()

	// Wait for shutdown or error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Infof("signal %v received, shutting down", sig)

		// Close the API router if it holds resources
		if err := apiRouter.Close(); err != nil {
			logger.WithError(err).Warning("error closing API router")
		}

		// Graceful shutdown of the HTTP server
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := apiServer.Shutdown(ctx); err != nil {
			logger.WithError(err).Warning("error during graceful shutdown")
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
