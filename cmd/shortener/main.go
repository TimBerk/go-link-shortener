package main

import (
	"log"
	"net/http"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitConfig()
	store := store.NewURLStore()
	handler := handler.NewHandler(store, cfg)

	router := chi.NewRouter()

	router.Post("/", handler.ShortenURL)
	router.Get("/{id}", handler.Redirect)

	log.Printf("Starting server on %s ...\n", cfg.ServerAddress)
	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
