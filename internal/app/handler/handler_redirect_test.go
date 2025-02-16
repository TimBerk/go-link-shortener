package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/stretchr/testify/assert"
)

type MockStore struct {
	originalURL string
	exists      bool
	addedURL    string
	addedURLs   batch.BatchRequest
}

func (m *MockStore) GetOriginalURL(shortURL string) (string, bool) {
	return m.originalURL, m.exists
}

func (m *MockStore) AddURLs(urls batch.BatchRequest) (batch.BatchResponse, error) {
	m.addedURLs = urls

	responses := make(batch.BatchResponse, 1)
	responses[0] = batch.ItemResponse{CorrelationID: "test_1", ShortURL: "abc123"}
	return responses, nil
}

func (m *MockStore) AddURL(url string) (string, error) {
	m.addedURL = url
	return "abc123", nil
}

func (m *MockStore) Ping(ctx context.Context) error {
	return nil
}

func TestShortenURL_Success(t *testing.T) {
	mockConfig := config.NewConfig("localhost:8021", "http://base.url", true)
	mockStore := &MockStore{}
	handler := NewHandler(mockStore, mockConfig)
	body := strings.NewReader("https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/shorten", body)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.ShortenURL(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	expectedResponse := "http://localhost:8021/abc123"
	assert.Equal(t, expectedResponse, w.Body.String())
}

func TestRedirect_Success(t *testing.T) {
	mockConfig := config.NewConfig("localhost:8021", "http://base:url", true)
	mockStore := &MockStore{
		originalURL: "https://example.com",
		exists:      true,
	}
	handler := NewHandler(mockStore, mockConfig)
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()

	handler.Redirect(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
}
