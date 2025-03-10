package handler

import (
	"context"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/pkg/cookies"
	"net/http"
	"time"

	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/stretchr/testify/mock"
)

func mockCookie(userID string) *http.Cookie {
	valueCookie, _ := cookies.GetEncodedValue(userID)
	return &http.Cookie{
		Name:     "user",
		Value:    valueCookie,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}
}

type MockURLStore struct {
	mock.Mock
}

func (m *MockURLStore) AddURL(ctx context.Context, originalURL string, userID string) (string, error) {
	args := m.Called(originalURL, userID)
	return args.String(0), nil
}

func (m *MockURLStore) AddURLs(ctx context.Context, urls batch.BatchRequest, userID string) (batch.BatchResponse, error) {
	args := m.Called(urls)

	var responses batch.BatchResponse
	for i := 0; i < len(urls); i++ {
		response, _ := args.Get(i).(batch.ItemResponse)
		responses = append(responses, response)
	}

	return responses, nil
}

func (m *MockURLStore) GetOriginalURL(ctx context.Context, shortURL string, userID string) (string, bool, bool) {
	args := m.Called(shortURL, userID)
	return args.String(0), args.Bool(1), false
}

func (m *MockURLStore) Ping(ctx context.Context) error {
	return nil
}
func (m *MockURLStore) DeleteURL(ctx context.Context, batch []store.URLPair) error {
	return nil
}
