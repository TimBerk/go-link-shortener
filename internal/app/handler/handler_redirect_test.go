package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockStore struct {
	originalURL string
	exists      bool
	addedURL    string
}

func (m *MockStore) GetOriginalURL(shortURL string) (string, bool) {
	return m.originalURL, m.exists
}

func (m *MockStore) AddURL(url string) string {
	m.addedURL = url
	return "abc123"
}

func TestShortenURL_Success(t *testing.T) {
	mockStore := &MockStore{}
	handler := NewHandler(mockStore)
	body := strings.NewReader("https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/shorten", body)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.ShortenURL(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	expectedResponse := "http://localhost:8080/abc123"
	assert.Equal(t, expectedResponse, w.Body.String())
}

func TestRedirect_Success(t *testing.T) {
	mockStore := &MockStore{
		originalURL: "https://example.com",
		exists:      true,
	}
	handler := NewHandler(mockStore)
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()

	handler.Redirect(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
}
