package store

import (
	"math/rand"
)

type URLStore struct {
	linksMap map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{
		linksMap: make(map[string]string),
	}
}

const (
	letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	digits  = "0123456789"
	length  = 6
)

var generateIDFunc = func() string {
	chars := letters + digits

	id := make([]byte, length)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}

	return string(id)
}

func (s *URLStore) generateID() string {
	return generateIDFunc()
}

type URLStoreInterface interface {
	AddURL(originalURL string) string
	GetOriginalURL(shortURL string) (string, bool)
}

func (s *URLStore) AddURL(originalURL string) string {
	for shortURL, url := range s.linksMap {
		if url == originalURL {
			return shortURL
		}
	}

	shortURL := s.generateID()
	s.linksMap[shortURL] = originalURL
	return shortURL
}

func (s *URLStore) GetOriginalURL(shortURL string) (string, bool) {
	originalURL, exists := s.linksMap[shortURL]
	return originalURL, exists
}
