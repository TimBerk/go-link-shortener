// Package handler обрабатывает данные для api-эндопоинтов.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/models/simple"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/pkg/cookies"
	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"
)

// userURLs - локальный store для быстрого нахождения ссылок пользователя
var userURLs = make(map[string][]map[string]string)

// Handler - структура для хранения настроек и обработчиков данных
type Handler struct {
	store   store.Store
	cfg     *config.Config
	ctx     context.Context
	urlChan chan store.URLPair
}

// addShortURL - обработчик локального хранилища ссылок пользователя
func addShortURL(userID, shortURL, originalURL string) {
	userURLs[userID] = append(userURLs[userID], map[string]string{
		"short_url":    shortURL,
		"original_url": originalURL,
	})
}

// NewHandler - инициализация нового обработчика на основании переаданных настроек
func NewHandler(store store.Store, cfg *config.Config, ctx context.Context, urlChan chan store.URLPair) *Handler {
	return &Handler{store: store, cfg: cfg, ctx: ctx, urlChan: urlChan}
}

// ErrorResponse стандартный формат ошибки API
// swagger:model
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// RequestJSON запрос на сокращение URL
// swagger:model
type RequestJSON struct {
	URL string `json:"url" example:"https://example.com"`
}

// ShortenURL обрабатывает запрос на сокращение URL
// @Summary Сократить URL (текстовый формат)
// @Description Создает короткую версию переданного URL
// @Accept  text/plain
// @Produce text/plain
// @Param   url body string true "Оригинальный URL"
// @Success 201 {string} string "Сокращенный URL"
// @Success 409 {string} string "URL уже существует"
// @Failure 400 {string} string "Неверный запрос"
// @Router / [post]
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

	shortURL, err := h.store.AddURL(h.ctx, originalURL, userID)
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

// ShortenJSONURL обрабатывает запрос на сокращение URL в JSON формате
// @Summary Сократить URL (JSON формат)
// @Description Создает короткую версию переданного URL
// @Accept  json
// @Produce json
// @Param   request body simple.RequestJSON true "Запрос с URL"
// @Success 201 {object} simple.ResponseJSON
// @Success 409 {object} simple.ResponseJSON
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Router /api/shorten [post]
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

	shortURL, err := h.store.AddURL(h.ctx, jsonBody.URL, userID)
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

// Redirect выполняет перенаправление по короткому URL
// @Summary Перенаправить по короткому URL
// @Description Перенаправляет на оригинальный URL по короткому идентификатору
// @Param   id path string true "Короткий идентификатор URL"
// @Success 307 "Перенаправление на оригинальный URL"
// @Failure 404 {string} string "URL не найден"
// @Failure 410 {string} string "URL удален"
// @Router /{id} [get]
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	userID, err := cookies.GetUserID(r)
	if err != nil {
		userID = cookies.GenerateUserID()
		cookies.SetUserCookie(w, userID)
	}

	shortURL := chi.URLParam(r, "id")
	originalURL, exists, isDeleted := h.store.GetOriginalURL(h.ctx, shortURL, userID)
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

// Ping проверяет соединение с базой данных
// @Summary Проверить соединение с БД
// @Description Проверяет доступность базы данных
// @Success 200 "База данных доступна"
// @Failure 500 {string} string "Ошибка соединения с БД"
// @Router /ping [get]
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

// ShortenBatch обрабатывает пакетное создание коротких URL
// @Summary Пакетное создание коротких URL
// @Description Создает несколько коротких URL за один запрос
// @Accept  json
// @Produce json
// @Param   urls body []batch.BatchRequest true "Массив URL для сокращения"
// @Success 201 {array} batch.BatchResponse
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /api/shorten/batch [post]
func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	var batchRequests batch.BatchRequest

	userID, err := cookies.GetUserID(r)
	if err != nil {
		userID = cookies.GenerateUserID()
		cookies.SetUserCookie(w, userID)
	}

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

	batchResponses, err := h.store.AddURLs(h.ctx, batchRequests, userID)
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

// UserURLsHandler возвращает все URL пользователя
// @Summary Получить URL пользователя
// @Description Возвращает все сокращенные URL текущего пользователя
// @Produce json
// @Success 200 {array} map[string]string "Массив URL пользователя"
// @Success 204 "Нет сохраненных URL"
// @Router /api/user/urls [get]
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

// DeleteURLsHandler помечает URL как удаленные
// @Summary Удалить URL пользователя
// @Description Помечает указанные URL как удаленные (асинхронно)
// @Accept  json
// @Produce json
// @Param   urls body []string true "Массив коротких URL для удаления"
// @Success 202 "Запрос принят в обработку"
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Failure 401 {string} string "Пользователь не авторизован"
// @Router /api/user/urls [delete]
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
		logrus.WithFields(logrus.Fields{
			"shortURL": shortURL,
			"UserID":   userID,
		}).Info("Deleted user link")
		h.urlChan <- store.URLPair{ShortURL: shortURL, UserID: userID}
	}

	w.WriteHeader(http.StatusAccepted)
}
