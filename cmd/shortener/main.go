package main

import (
	"log"
	"net/http"

	"github.com/TimBerk/go-link-shortener/internal/app/handler"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

func main() {
	store := store.NewURLStore()
	handler := handler.NewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.ShortenURL(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/{id}", handler.Redirect)

	log.Println("Starting server on :8080...")
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
