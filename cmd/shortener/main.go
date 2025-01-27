package main

import (
	"net/http"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/logger"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitConfig()
	logger.Initialize(cfg.LogLevel)
	generator := store.NewIDGenerator()
	store := store.NewURLStore(generator)
	handler := handler.NewHandler(store, cfg)

	router := chi.NewRouter()

	router.Post("/api/shorten", handler.ShortenJSONURL)
	router.Get("/{id}", handler.Redirect)
	router.Post("/", handler.ShortenURL)

	logger.Log.WithField("address", cfg.ServerAddress).Info("Starting server")
	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		logger.Log.Fatal("ListenAndServe: ", err)
	}
}
