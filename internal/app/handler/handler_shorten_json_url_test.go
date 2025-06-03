package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

func TestShortenJsonURLHandler(t *testing.T) {
	ctx := context.Background()
	urlChan := make(chan store.URLPair, 1000)
	mockConfig := config.NewConfig("localhost:8021", "http://base.loc", true, "192.168.1.0/24")
	mockStore := new(MockURLStore)
	testHandler := NewHandler(mockStore, mockConfig, ctx, urlChan)
	testCookie := mockCookie(userID)

	tests := []struct {
		name               string
		method             string
		contentType        string
		body               string
		bodyURL            string
		mockReturnShortURL string
		expectedStatus     int
		expectedResponse   string
	}{
		{
			name:               "Valid Json POST request",
			method:             http.MethodPost,
			contentType:        "application/json",
			body:               `{"url":"https://example.com"}`,
			bodyURL:            "https://example.com",
			mockReturnShortURL: "short1",
			expectedStatus:     http.StatusCreated,
			expectedResponse:   `{"result":"http://localhost:8021/short1"}`,
		},
		{
			name:             "Empty body",
			method:           http.MethodPost,
			contentType:      "application/json",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"Failed to decode request body"}`,
		},
		{
			name:             "Empty body",
			method:           http.MethodPost,
			contentType:      "application/json",
			body:             `{}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"Empty request body"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockReturnShortURL != "" {
				mockStore.On("AddURL", mock.Anything, test.bodyURL, userID).Return(test.mockReturnShortURL)
			}
			req := httptest.NewRequest(test.method, "/api/shorten", bytes.NewBufferString(test.body))
			req.Header.Set("Content-Type", test.contentType)
			req.AddCookie(testCookie)
			recorder := httptest.NewRecorder()

			testHandler.ShortenJSONURL(recorder, req)

			assert.Equal(t, test.expectedStatus, recorder.Code, "Неверный статус код для теста: %s", test.name)
			assert.Equal(t, test.expectedResponse, strings.TrimSuffix(recorder.Body.String(), "\n"), "Неверное тело ответа для теста: %s", test.name)
			mockStore.AssertExpectations(t)
		})
	}
}
