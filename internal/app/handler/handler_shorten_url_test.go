package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/app/store/local"
	"github.com/stretchr/testify/assert"
)

func TestShortenURL(t *testing.T) {
	ctx := context.Background()
	urlChan := make(chan store.URLPair, 1000)
	mockConfig := config.NewConfig("localhost:8021", "http://base.loc", true)
	mockStore := new(MockURLStore)
	testHandler := NewHandler(mockStore, mockConfig, ctx, urlChan)
	userID := "test"
	testCookie := mockCookie(userID)

	tests := []struct {
		name               string
		method             string
		contentType        string
		body               string
		mockReturnShortURL string
		expectedStatus     int
		expectedResponse   string
	}{
		{
			name:               "Valid POST request",
			method:             http.MethodPost,
			contentType:        "text/plain",
			body:               "https://example.com",
			mockReturnShortURL: "short1",
			expectedStatus:     http.StatusCreated,
			expectedResponse:   "http://localhost:8021/short1",
		},
		{
			name:             "Empty body",
			method:           http.MethodPost,
			contentType:      "text/plain",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Empty request body\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockReturnShortURL != "" {
				usedParam := test.body
				if usedParam == "" {
					usedParam = mockConfig.BaseURL
				}
				mockStore.On("AddURL", usedParam, userID).Return(test.mockReturnShortURL)
			}
			req := httptest.NewRequest(test.method, "/shorten", bytes.NewBufferString(test.body))
			req.Header.Set("Content-Type", test.contentType)
			req.AddCookie(testCookie)
			recorder := httptest.NewRecorder()

			testHandler.ShortenURL(recorder, req)

			assert.Equal(t, test.expectedStatus, recorder.Code, "Неверный статус код для теста: %s", test.name)
			assert.Equal(t, test.expectedResponse, recorder.Body.String(), "Неверное тело ответа для теста: %s", test.name)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestAddURL_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	var results []string
	ctx := context.Background()
	testGen := store.NewIDGenerator()
	userID := "test"
	testStore, _ := local.NewURLStore(testGen)
	originalURL := "https://example.com"

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			shortLink, _ := testStore.AddURL(ctx, originalURL, userID)
			results = append(results, shortLink)
		}()
	}

	wg.Wait()

	firstResult := results[0]
	for _, result := range results {
		if result != firstResult {
			t.Errorf("Expected all results to be %s, got %s", firstResult, result)
		}
	}
}
