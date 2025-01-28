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
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitConfig()
	logger.Initialize(cfg.LogLevel)
	generator := store.NewIDGenerator()

	var dataStore store.MainStoreInterface
	if cfg.UseLocalStore {
		dataStore = local.NewURLStore(generator)
	} else {
		dataStore = json.NewJSONStore(cfg.FileStoragePath, generator)
	}
	handler := handler.NewHandler(dataStore, cfg)

	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GzipMiddleware)

	router.Post("/api/shorten", handler.ShortenJSONURL)
	router.Get("/{id}", handler.Redirect)
	router.Post("/", handler.ShortenURL)

	logger.Log.WithField("address", cfg.ServerAddress).Info("Starting server")
	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		logger.Log.Fatal("ListenAndServe: ", err)
	}
}
