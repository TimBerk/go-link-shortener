package handler

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/mock"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/models/simple"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

func setupTestHandler(mockStore *MockURLStore) *Handler {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		LogLevel:      "fatal",
	}
	urlChan := make(chan store.URLPair, 100)
	return NewHandler(mockStore, cfg, context.Background(), urlChan)
}

func BenchmarkShortenURL(b *testing.B) {
	testCookie := mockCookie(userID)
	mockStore := new(MockURLStore)
	mockStore.On("AddURL", mock.Anything, "http://example.com", userID).
		Return("short123", nil)
	h := setupTestHandler(mockStore)

	body := bytes.NewBufferString("http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", body)
		r.AddCookie(testCookie)
		h.ShortenURL(w, r)
		body.Reset()
	}

	mockStore.AssertExpectations(b)
}

func BenchmarkShortenJSONURL(b *testing.B) {
	testCookie := mockCookie(userID)
	mockStore := new(MockURLStore)
	mockStore.On("AddURL", mock.Anything, "http://example.com", userID).
		Return("short123", nil)
	h := setupTestHandler(mockStore)

	jsonBody := simple.RequestJSON{URL: "http://example.com"}
	body, _ := easyjson.Marshal(jsonBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.AddCookie(testCookie)
		h.ShortenJSONURL(w, r)
	}

	mockStore.AssertExpectations(b)
}

func BenchmarkPing(b *testing.B) {
	mockStore := new(MockURLStore)
	mockStore.On("Ping", mock.Anything).Return(nil)
	h := setupTestHandler(mockStore)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/ping", nil)
		h.Ping(w, r)
	}

	mockStore.AssertExpectations(b)
}

func BenchmarkShortenBatch(b *testing.B) {
	testCookie := mockCookie(userID)
	mockStore := new(MockURLStore)
	h := setupTestHandler(mockStore)

	batchReq := batch.BatchRequest{
		{CorrelationID: "1", OriginalURL: "http://example.com/1"},
		{CorrelationID: "2", OriginalURL: "http://example.com/2"},
	}
	expectedResponse := batch.BatchResponse{
		{CorrelationID: "1", ShortURL: "short_1"},
		{CorrelationID: "2", ShortURL: "short_2"},
	}

	mockStore.On("AddURLs", mock.Anything, batchReq, userID).
		Return(expectedResponse, nil).Times(b.N)

	body, _ := easyjson.Marshal(batchReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.AddCookie(testCookie)
		h.ShortenBatch(w, r)
	}

	mockStore.AssertExpectations(b)
}

func BenchmarkUserURLsHandler(b *testing.B) {
	testCookie := mockCookie(userID)
	mockStore := new(MockURLStore)
	h := setupTestHandler(mockStore)

	addShortURL(userID, "http://short/1", "http://original/1")
	addShortURL(userID, "http://short/2", "http://original/2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/user/urls", nil)
		r.AddCookie(testCookie)
		h.UserURLsHandler(w, r)
	}
}
