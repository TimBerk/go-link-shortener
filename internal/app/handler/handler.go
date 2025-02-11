package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/models/simple"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/app/store/pg"
	"github.com/mailru/easyjson"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store store.MainStoreInterface
	cfg   *config.Config
}

func NewHandler(store store.MainStoreInterface, cfg *config.Config) *Handler {
	return &Handler{store: store, cfg: cfg}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.store.AddURL(originalURL)
	if err != nil {
		http.Error(w, "Error getting url", http.StatusBadRequest)
		return
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", h.cfg.ServerAddress, shortURL)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fullShortURL))
}

func (h *Handler) ShortenJSONURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !utils.CheckParamInHeaderParam(r, "Content-Type", "application/json") {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	var jsonBody simple.RequestJSON
	if err := easyjson.UnmarshalFromReader(r.Body, &jsonBody); err != nil {
		utils.WriteJSONError(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	if jsonBody.URL == "" {
		utils.WriteJSONError(w, "Empty request body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.store.AddURL(jsonBody.URL)
	if err != nil {
		utils.WriteJSONError(w, "Error getting url", http.StatusBadRequest)
		return
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", h.cfg.ServerAddress, shortURL)
	responseJSON := simple.ResponseJSON{Result: fullShortURL}

	response, err := easyjson.Marshal(responseJSON)
	if err != nil {
		utils.WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	originalURL, exists := h.store.GetOriginalURL(shortURL)
	if !exists {
		logrus.WithFields(logrus.Fields{
			"uri":      originalURL,
			"shortUri": shortURL,
		}).Error("Short URL not found")
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Second)
	defer cancel()

	pgService, err := pg.NewPgPool(ctx, h.cfg.DatabaseDSN)
	if err != nil {
		logrus.WithField("err", err).Error("Connect to DB")
		http.Error(w, "error to create connection to DB", http.StatusInternalServerError)
		return
	}

	if pgService.Ping(ctx) != nil {
		logrus.WithField("err", err).Error("Check connection to DB")
		http.Error(w, "error to check connection to DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	var batchRequests batch.BatchRequest

	if err := easyjson.UnmarshalFromReader(r.Body, &batchRequests); err != nil {
		logrus.WithField("err", err).Error("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(batchRequests) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	batchResponses, err := h.store.AddURLs(batchRequests)
	if err != nil {
		logrus.WithField("err", err).Error("Error shortening URLs")
		http.Error(w, fmt.Sprintf("Error shortening URLs: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := easyjson.Marshal(batchResponses)
	if err != nil {
		logrus.WithField("err", err).Error("Error encoding response")
		utils.WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
