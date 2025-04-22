package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/stretchr/testify/mock"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	models "github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

var (
	mockStore     = new(MockURLStore)
	cfg           = &config.Config{ServerAddress: "localhost:8090"}
	h             = NewHandler(mockStore, cfg, context.Background(), make(chan store.URLPair))
	exampleUserID = "888"
	testCookie    = mockCookie(exampleUserID)
	originalURL   = "https://example.com/original"
	shortCode     = "abc123"
	shortURL      = "http://localhost:8090/abc123"
)

func ExampleHandler_ShortenURL() {
	mockStore.On("AddURL", mock.Anything, originalURL, exampleUserID).Return(shortCode, nil)
	// Подготовка тестового запроса
	body := bytes.NewBufferString(originalURL)
	req := httptest.NewRequest("POST", "/", body)
	req.AddCookie(testCookie)
	rec := httptest.NewRecorder()

	// Вызов обработчика
	h.ShortenURL(rec, req)

	// Проверка ответа
	fmt.Printf("Status: %d\n", rec.Code)
	fmt.Printf("Location:%s\n", rec.Header().Get("Location"))
	fmt.Printf("Body: %s\n", rec.Body.String())

	// Output:
	// Status: 201
	// Location:
	// Body: http://localhost:8090/abc123
}

func ExampleHandler_ShortenJSONURL() {
	mockStore.On("AddURL", mock.Anything, originalURL, exampleUserID).Return(shortCode, nil)
	// Подготовка JSON запроса
	jsonBody := `{"url": "https://example.com/original"}`
	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(testCookie)
	rec := httptest.NewRecorder()

	// Вызов обработчика
	h.ShortenJSONURL(rec, req)

	// Проверка ответа
	fmt.Printf("Status: %d\n", rec.Code)
	fmt.Printf("Content-Type: %s\n", rec.Header().Get("Content-Type"))
	fmt.Printf("Body: %s\n", rec.Body.String())

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Body: {"result":"http://localhost:8090/abc123"}
}

func ExampleHandler_Ping() {
	mockStore.On("Ping", mock.Anything).Return(nil)

	req := httptest.NewRequest("GET", "/ping", nil)
	req.AddCookie(testCookie)
	rec := httptest.NewRecorder()

	h.Ping(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)

	// Output:
	// Status: 200
}

func ExampleHandler_ShortenBatch() {
	// Подготовка batch запроса
	batchReq := models.BatchRequest{
		{CorrelationID: "1", OriginalURL: originalURL},
	}
	mockStore.On("AddURLs", mock.Anything, batchReq, exampleUserID).Return(models.BatchResponse{}, nil)
	body, _ := json.Marshal(batchReq)

	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(testCookie)
	rec := httptest.NewRecorder()

	h.ShortenBatch(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	fmt.Printf("Content-Type: %s\n", rec.Header().Get("Content-Type"))

	// Output:
	// Status: 201
	// Content-Type: application/json
}

func ExampleHandler_UserURLsHandler() {
	// Добавление тестовых данных
	addShortURL(exampleUserID, shortURL, originalURL)

	// Подготовка запроса
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	req.AddCookie(testCookie)
	rec := httptest.NewRecorder()

	h.UserURLsHandler(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	fmt.Printf("Content-Type: %s\n", rec.Header().Get("Content-Type"))

	// Output:
	// Status: 200
	// Content-Type: application/json
}
