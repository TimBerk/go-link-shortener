package router

import (
	"context"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/compress"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/go-chi/chi/v5"
)

func RegisterRouters(dataStore store.Store, cfg *config.Config, ctx context.Context) chi.Router {
	handler := handler.NewHandler(dataStore, cfg, ctx)

	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GzipMiddleware)

	router.Get("/ping", handler.Ping)
	router.Get("/api/user/urls", handler.UserURLsHandler)
	router.Delete("/api/user/urls", handler.DeleteURLsHandler)
	router.Post("/api/shorten/batch", handler.ShortenBatch)
	router.Post("/api/shorten", handler.ShortenJSONURL)
	router.Get("/{id}", handler.Redirect)
	router.Post("/", handler.ShortenURL)

	return router
}
