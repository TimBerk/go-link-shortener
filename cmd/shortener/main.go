package main

import (
	"log"
	"net/http"

	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/go-chi/chi/v5"
)

func main() {
	store := store.NewURLStore()
	handler := handler.NewHandler(store)

	router := chi.NewRouter()

	router.Post("/", handler.ShortenURL)
	router.Get("/{id}", handler.Redirect)

	log.Println("Starting server on :8080...")
	err := http.ListenAndServe("localhost:8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
