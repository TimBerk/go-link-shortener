package handler

import (
	"fmt"
	"io"
	"net/http"

	stores "github.com/TimBerk/go-link-shortener/internal/app/store"

	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store stores.URLStoreInterface
}

func NewHandler(store stores.URLStoreInterface) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if !utils.CheckParamInHeaderParam(r, "Content-Type", "text/plain") {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	shortURL := h.store.AddURL(originalURL)

	fullShortURL := fmt.Sprintf("http://localhost:8080/%s", shortURL)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fullShortURL))
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	shortURL := chi.URLParam(r, "id")
	originalURL, exists := h.store.GetOriginalURL(shortURL)
	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
