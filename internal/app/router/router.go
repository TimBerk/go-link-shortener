// Package router обрабатывает пути для приложения
package router

import (
	"context"
	"net/http/pprof"

	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/checker"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/compress"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

// addPprof - регистрирует пути pprof-обработчиков
func addPprof(router chi.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("goroutine"))
}

// RegisterRouters - регистрирует пути приложения
func RegisterRouters(dataStore store.Store, cfg *config.Config, ctx context.Context, urlChan chan store.URLPair) chi.Router {
	h := handler.NewHandler(dataStore, cfg, ctx, urlChan)

	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GzipMiddleware)

	if cfg.TrustedSubnet != "" {
		router.Use(checker.TrustedSubnetMiddleware(cfg))
		router.Get("/api/internal/stats", h.StatsHandler)
	}

	addPprof(router)

	router.Get("/ping", h.Ping)
	router.Get("/api/user/urls", h.UserURLsHandler)
	router.Delete("/api/user/urls", h.DeleteURLsHandler)
	router.Post("/api/shorten/batch", h.ShortenBatch)
	router.Post("/api/shorten", h.ShortenJSONURL)
	router.Get("/{id}", h.Redirect)
	router.Post("/", h.ShortenURL)

	// Swagger documentation route
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	))

	return router
}
