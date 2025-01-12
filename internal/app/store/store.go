package store

import (
	"math/rand"
)

const (
	chars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	length = 6
)

type IDGenerator struct{}

type Generator interface {
	Next() string
}

type URLStore struct {
	linksMap    map[string]string
	originalMap map[string]string
	gen         Generator
}

func NewIDGenerator() Generator {
	return &IDGenerator{}
}

func (g *IDGenerator) Next() string {
	id := make([]byte, length)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}

	return string(id)
}

func NewURLStore(gen Generator) *URLStore {
	return &URLStore{
		linksMap:    make(map[string]string),
		originalMap: make(map[string]string),
		gen:         gen,
	}
}

type URLStoreInterface interface {
	AddURL(originalURL string) string
	GetOriginalURL(shortURL string) (string, bool)
}

func (s *URLStore) AddURL(originalURL string) string {
	if shortURL, exists := s.originalMap[originalURL]; exists {
		return shortURL
	}

	shortURL := s.gen.Next()

	if _, exists := s.linksMap[shortURL]; exists {
		return s.AddURL(originalURL)
	}

	s.linksMap[shortURL] = originalURL
	s.originalMap[originalURL] = shortURL
	return shortURL
}

func (s *URLStore) GetOriginalURL(shortURL string) (string, bool) {
	originalURL, exists := s.linksMap[shortURL]
	return originalURL, exists
}
