package json

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/google/uuid"
)

type JSONRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type JSONStore struct {
	storage     map[string]JSONRecord
	fullStorage map[string]JSONRecord
	filePath    string
	gen         store.Generator
	mutex       sync.Mutex
}

func NewJSONStore(filePath string, gen store.Generator) *JSONStore {
	return &JSONStore{
		storage:     make(map[string]JSONRecord),
		fullStorage: make(map[string]JSONRecord),
		filePath:    filePath,
		gen:         gen,
	}
}

func (s *JSONStore) loadStorage() error {
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry JSONRecord
		if err := decoder.Decode(&entry); err != nil {
			return err
		}
		s.storage[entry.ShortURL] = entry
	}
	return nil
}

func (s *JSONStore) saveStorage() error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, entry := range s.storage {
		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}
	return nil
}

func (s *JSONStore) AddURL(originalURL string) string {
	s.mutex.Lock()
	s.loadStorage()
	defer s.mutex.Unlock()

	if record, exists := s.fullStorage[originalURL]; exists {
		return record.ShortURL
	}

	shortURL := s.gen.Next()

	if _, exists := s.storage[shortURL]; exists {
		return s.AddURL(originalURL)
	}

	record := JSONRecord{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UUID:        uuid.New().String(),
	}

	s.storage[shortURL] = record
	s.fullStorage[originalURL] = record
	s.saveStorage()
	return shortURL
}

func (s *JSONStore) GetOriginalURL(shortURL string) (string, bool) {
	record, exists := s.storage[shortURL]
	return record.OriginalURL, exists
}
