package main

import (
	"net/http"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/compress"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/app/store/json"
	"github.com/TimBerk/go-link-shortener/internal/app/store/local"
	"github.com/TimBerk/go-link-shortener/internal/app/store/pg"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitConfig()
	logger.Initialize(cfg.LogLevel)
	generator := store.NewIDGenerator()

	var dataStore store.MainStoreInterface
	var errStore error

	if cfg.DatabaseDSN != "" {
		dataStore, errStore = pg.NewPgStore(generator, cfg)
	} else if cfg.UseLocalStore {
		dataStore, errStore = local.NewURLStore(generator)
	} else {
		dataStore, errStore = json.NewJSONStore(cfg.FileStoragePath, generator)
	}
	if errStore != nil {
		logger.Log.Fatal("Read Store: ", errStore)
	}

	handler := handler.NewHandler(dataStore, cfg)

	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GzipMiddleware)

	router.Get("/ping", handler.Ping)
	router.Post("/api/shorten/batch", handler.ShortenBatch)
	router.Post("/api/shorten", handler.ShortenJSONURL)
	router.Get("/{id}", handler.Redirect)
	router.Post("/", handler.ShortenURL)

	logger.Log.WithField("address", cfg.ServerAddress).Info("Starting server")
	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		logger.Log.Fatal("ListenAndServe: ", err)
	}
}
