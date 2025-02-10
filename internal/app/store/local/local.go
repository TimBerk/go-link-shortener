package local

import (
	"sync"

	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

type URLStore struct {
	linksMap    map[string]string
	originalMap map[string]string
	gen         store.Generator
	mutex       sync.Mutex
}

func NewURLStore(gen store.Generator) (*URLStore, error) {
	return &URLStore{
		linksMap:    make(map[string]string),
		originalMap: make(map[string]string),
		gen:         gen,
	}, nil
}

func (s *URLStore) AddURL(originalURL string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if shortURL, exists := s.originalMap[originalURL]; exists {
		return shortURL, nil
	}

	shortURL := s.gen.Next()

	if _, exists := s.linksMap[shortURL]; exists {
		return s.AddURL(originalURL)
	}

	s.linksMap[shortURL] = originalURL
	s.originalMap[originalURL] = shortURL
	return shortURL, nil
}

func (s *URLStore) GetOriginalURL(shortURL string) (string, bool) {
	originalURL, exists := s.linksMap[shortURL]
	return originalURL, exists
}
