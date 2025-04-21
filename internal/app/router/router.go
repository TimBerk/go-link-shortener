package router

import (
	"context"
	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/compress"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/go-chi/chi/v5"
	"net/http/pprof"
)

func addPprof(router chi.Router) {
	// Регистрация pprof-обработчиков
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

func RegisterRouters(dataStore store.Store, cfg *config.Config, ctx context.Context, urlChan chan store.URLPair) chi.Router {
	handler := handler.NewHandler(dataStore, cfg, ctx, urlChan)

	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Use(compress.GzipMiddleware)

	addPprof(router)

	router.Get("/ping", handler.Ping)
	router.Get("/api/user/urls", handler.UserURLsHandler)
	router.Delete("/api/user/urls", handler.DeleteURLsHandler)
	router.Post("/api/shorten/batch", handler.ShortenBatch)
	router.Post("/api/shorten", handler.ShortenJSONURL)
	router.Get("/{id}", handler.Redirect)
	router.Post("/", handler.ShortenURL)

	return router
}
