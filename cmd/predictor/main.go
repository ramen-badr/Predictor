package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	_ "predictor/docs"
	"predictor/internal/config"
	"predictor/internal/http-server/handlers/people/delete"
	"predictor/internal/http-server/handlers/people/get"
	"predictor/internal/http-server/handlers/people/save"
	"predictor/internal/http-server/handlers/people/update"
	"predictor/internal/http-server/middleware/mwLogger"
	"predictor/internal/lib/logger/sLogger"
	"predictor/internal/storage/postgres"
)

// @title People API
// @version 1.0
// @description API for managing people_info records
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.MustLoad()

	log := sLogger.SetupLogger(cfg.Env)

	log.Info("starting predictor", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	store, err := postgres.New(cfg.Storage)
	if err != nil {
		log.Error("failed to initialize storage", sLogger.Error(err))
		return
	}

	log.Debug("storage is initialized")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/people", save.New(log, store))
	router.Get("/", get.New(log, store))
	router.Delete("/people/{id}", delete.New(log, store))
	router.Put("/people/{id}", update.New(log, store))
	router.Patch("/people/{id}", update.New(log, store))
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	log.Debug("router is initialized")
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err = srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}
