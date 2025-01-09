package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockURLStore struct {
	mock.Mock
}

func (m *MockURLStore) AddURL(originalURL string) string {
	args := m.Called(originalURL)
	return args.String(0)
}

func (m *MockURLStore) GetOriginalURL(shortURL string) (string, bool) {
	args := m.Called(shortURL)
	return args.String(0), args.Bool(1)
}

func TestShortenURL(t *testing.T) {
	mockConfig := config.NewConfig("localhost:8021", "http://base.loc")
	mockStore := new(MockURLStore)
	testHandler := NewHandler(mockStore, mockConfig)

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
			name:               "Empty body",
			method:             http.MethodPost,
			contentType:        "text/plain",
			mockReturnShortURL: "short2",
			expectedStatus:     http.StatusCreated,
			expectedResponse:   "http://localhost:8021/short2",
		},
		{
			name:             "Invalid request method",
			method:           http.MethodGet,
			contentType:      "text/plain",
			body:             "https://example.com",
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedResponse: "Invalid request method\n",
		},
		{
			name:             "Invalid Content-Type",
			method:           http.MethodPost,
			contentType:      "application/json",
			body:             "https://example.com",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Invalid Content-Type\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.mockReturnShortURL != "" {
				usedParam := test.body
				if usedParam == "" {
					usedParam = mockConfig.BaseURL
				}
				mockStore.On("AddURL", usedParam).Return(test.mockReturnShortURL)
			}
			req := httptest.NewRequest(test.method, "/shorten", bytes.NewBufferString(test.body))
			req.Header.Set("Content-Type", test.contentType)
			recorder := httptest.NewRecorder()

			testHandler.ShortenURL(recorder, req)

			assert.Equal(t, test.expectedStatus, recorder.Code, "Неверный статус код для теста: %s", test.name)
			assert.Equal(t, test.expectedResponse, recorder.Body.String(), "Неверное тело ответа для теста: %s", test.name)
			mockStore.AssertExpectations(t)
		})
	}
}
