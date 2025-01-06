package shortener

import (
	"math/rand"
)

const (
	letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	digits  = "0123456789"
	length  = 6
)

func (s *URLStore) generateID() string {
	chars := letters + digits

	id := make([]byte, length)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}

	return string(id)
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
