package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/models/simple"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/mailru/easyjson"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/pkg/cookies"
	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"
	"github.com/go-chi/chi/v5"
)

var userURLs = make(map[string][]map[string]string)

func addShortURL(userID, shortURL, originalURL string) {
	userURLs[userID] = append(userURLs[userID], map[string]string{
		"short_url":    shortURL,
		"original_url": originalURL,
	})
}

type Handler struct {
	store   store.Store
	cfg     *config.Config
	ctx     context.Context
	urlChan chan store.URLPair
}

func NewHandler(store store.Store, cfg *config.Config, ctx context.Context, urlChan chan store.URLPair) *Handler {
	return &Handler{store: store, cfg: cfg, ctx: ctx, urlChan: urlChan}
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

	userID, err := cookies.GetUserID(r)
	if err != nil {
		userID = cookies.GenerateUserID()
		cookies.SetUserCookie(w, userID)
	}

	shortURL, err := h.store.AddURL(h.ctx, originalURL)
	existLink := errors.Is(err, store.ErrLinkExist)
	if err != nil && !existLink {
		http.Error(w, "Error getting url", http.StatusBadRequest)
		return
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", h.cfg.ServerAddress, shortURL)
	addShortURL(userID, fullShortURL, originalURL)

	if !existLink {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
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

	userID, err := cookies.GetUserID(r)
	if err != nil {
		userID = cookies.GenerateUserID()
		cookies.SetUserCookie(w, userID)
	}

	shortURL, err := h.store.AddURL(h.ctx, jsonBody.URL)
	existLink := errors.Is(err, store.ErrLinkExist)
	if err != nil && !existLink {
		utils.WriteJSONError(w, "Error getting url", http.StatusBadRequest)
		return
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", h.cfg.ServerAddress, shortURL)
	responseJSON := simple.ResponseJSON{Result: fullShortURL}
	addShortURL(userID, fullShortURL, jsonBody.URL)

	response, err := easyjson.Marshal(responseJSON)
	if err != nil {
		utils.WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if !existLink {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
	w.Write(response)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	originalURL, exists, isDeleted := h.store.GetOriginalURL(h.ctx, shortURL)
	if !exists {
		logrus.WithFields(logrus.Fields{
			"uri":      originalURL,
			"shortUri": shortURL,
		}).Error("Short URL not found")
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if isDeleted {
		logrus.WithFields(logrus.Fields{
			"uri":      originalURL,
			"shortUri": shortURL,
		}).Error("Short URL is deleted")
		http.Error(w, "Short URL is deleted", http.StatusGone)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.store.Ping(ctx)
	if err != nil {
		logrus.WithField("err", err).Error("Check connection to DB")
		http.Error(w, "failed to check connection to DB", http.StatusInternalServerError)
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

	batchResponses, err := h.store.AddURLs(h.ctx, batchRequests)
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

func (h *Handler) UserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := cookies.GetUserID(r)
	if err != nil {
		userID = cookies.GenerateUserID()
		cookies.SetUserCookie(w, userID)
	}

	urls, exists := userURLs[userID]
	if !exists || len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}

func (h *Handler) DeleteURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := cookies.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Отправляем данные в канал
	for _, shortURL := range shortURLs {
		h.urlChan <- store.URLPair{ShortURL: shortURL, UserID: userID}
	}

	w.WriteHeader(http.StatusAccepted)
}
