package local

import (
	"context"
	"sync"

	models "github.com/TimBerk/go-link-shortener/internal/app/models/batch"
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

func (s *URLStore) AddURL(ctx context.Context, originalURL string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if shortURL, exists := s.originalMap[originalURL]; exists {
		return shortURL, store.ErrLinkExist
	}

	shortURL := s.gen.Next()

	if _, exists := s.linksMap[shortURL]; exists {
		return s.AddURL(ctx, originalURL)
	}

	s.linksMap[shortURL] = originalURL
	s.originalMap[originalURL] = shortURL
	return shortURL, nil
}

func (s *URLStore) AddURLs(ctx context.Context, urls models.BatchRequest) (models.BatchResponse, error) {
	var responses models.BatchResponse

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, req := range urls {
		var shortURL string

		for {
			shortURL = s.gen.Next()
			if _, exists := s.linksMap[shortURL]; !exists {
				break
			}
		}

		s.linksMap[shortURL] = req.OriginalURL
		s.originalMap[req.OriginalURL] = shortURL

		responses = append(responses, models.ItemResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return responses, nil
}

func (s *URLStore) GetOriginalURL(ctx context.Context, shortURL string) (string, bool) {
	originalURL, exists := s.linksMap[shortURL]
	return originalURL, exists
}

func (s *URLStore) Ping(ctx context.Context) error {
	return nil
}
