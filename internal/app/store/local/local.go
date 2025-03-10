package local

import (
	"context"
	"sync"

	models "github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

type UserLink struct {
	UserID string
	Link   string
}

type URLStore struct {
	linksMap    map[string]UserLink
	originalMap map[string]UserLink
	userMap     map[string]string
	gen         store.Generator
	mutex       sync.Mutex
}

func NewURLStore(gen store.Generator) (*URLStore, error) {
	return &URLStore{
		linksMap:    make(map[string]UserLink),
		originalMap: make(map[string]UserLink),
		userMap:     make(map[string]string),
		gen:         gen,
	}, nil
}

func (s *URLStore) AddURL(ctx context.Context, originalURL string, userID string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if userLink, exists := s.originalMap[originalURL]; exists {
		return userLink.Link, store.ErrLinkExist
	}

	shortURL := s.gen.Next()

	if _, exists := s.linksMap[shortURL]; exists {
		return s.AddURL(ctx, originalURL, userID)
	}

	s.linksMap[shortURL] = UserLink{userID, originalURL}
	s.originalMap[originalURL] = UserLink{userID, shortURL}
	return shortURL, nil
}

func (s *URLStore) AddURLs(ctx context.Context, urls models.BatchRequest, userID string) (models.BatchResponse, error) {
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

		s.linksMap[shortURL] = UserLink{userID, req.OriginalURL}
		s.originalMap[req.OriginalURL] = UserLink{userID, shortURL}

		responses = append(responses, models.ItemResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return responses, nil
}

func (s *URLStore) GetOriginalURL(ctx context.Context, shortURL string, userID string) (string, bool, bool) {
	userLink, exists := s.linksMap[shortURL]
	return userLink.Link, exists, false
}

func (s *URLStore) Ping(ctx context.Context) error {
	return nil
}

func (s *URLStore) DeleteURL(ctx context.Context, batch []store.URLPair) error {
	if len(batch) == 0 {
		return nil
	}

	for _, pair := range batch {
		userLink, exists := s.linksMap[pair.ShortURL]
		if exists && userLink.UserID == pair.UserID {
			delete(s.originalMap, userLink.Link)
			delete(s.linksMap, pair.ShortURL)
		}
	}

	return nil
}
