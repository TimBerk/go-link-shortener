package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/pkg/cookies"
)

const (
	userID string = "777"
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
	args := m.Called(ctx, originalURL, userID)
	return args.String(0), nil
}

func (m *MockURLStore) AddURLs(ctx context.Context, urls batch.BatchRequest, userID string) (batch.BatchResponse, error) {
	args := m.Called(ctx, urls, userID)
	return args.Get(0).(batch.BatchResponse), args.Error(1)
}

func (m *MockURLStore) GetOriginalURL(ctx context.Context, shortURL string, userID string) (string, bool, bool) {
	args := m.Called(ctx, shortURL, userID)
	return args.String(0), args.Bool(1), false
}

func (m *MockURLStore) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockURLStore) DeleteURL(ctx context.Context, batch []store.URLPair) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *MockURLStore) GetURLCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockURLStore) GetUserCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
