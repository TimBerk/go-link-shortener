package handler

import (
	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/stretchr/testify/mock"
)

type MockURLStore struct {
	mock.Mock
}

func (m *MockURLStore) AddURL(originalURL string) (string, error) {
	args := m.Called(originalURL)
	return args.String(0), nil
}

func (m *MockURLStore) AddURLs(urls batch.BatchRequest) (batch.BatchResponse, error) {
	args := m.Called(urls)

	var responses batch.BatchResponse
	for i := 0; i < len(urls); i++ {
		response, _ := args.Get(i).(batch.ItemResponse)
		responses = append(responses, response)
	}

	return responses, nil
}

func (m *MockURLStore) GetOriginalURL(shortURL string) (string, bool) {
	args := m.Called(shortURL)
	return args.String(0), args.Bool(1)
}
