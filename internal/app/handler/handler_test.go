package handler

import (
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
