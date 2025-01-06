package shortener

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"
)

type Handler struct {
	store *URLStore
}

func NewHandler(store *URLStore) *Handler {
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

	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, exists := h.store.GetOriginalURL(shortURL)
	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
