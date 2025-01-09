package handler

import (
	"fmt"
	"io"
	"net/http"

	stores "github.com/TimBerk/go-link-shortener/internal/app/store"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store stores.URLStoreInterface
	cfg   *config.Config
}

func NewHandler(store stores.URLStoreInterface, cfg *config.Config) *Handler {
	return &Handler{store: store, cfg: cfg}
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
	if originalURL == "" {
		originalURL = h.cfg.BaseURL
	}
	shortURL := h.store.AddURL(originalURL)

	fullShortURL := fmt.Sprintf("http://%s/%s", h.cfg.ServerAddress, shortURL)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fullShortURL))
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	originalURL, exists := h.store.GetOriginalURL(shortURL)
	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
