package json

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	models "github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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

func NewJSONStore(filePath string, gen store.Generator) (*JSONStore, error) {
	store := &JSONStore{
		storage:     make(map[string]JSONRecord),
		fullStorage: make(map[string]JSONRecord),
		filePath:    filePath,
		gen:         gen,
	}

	err := store.loadStorage()
	if err != nil {
		return nil, fmt.Errorf("error loading json store: %s", err)
	}

	return store, nil
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

func (s *JSONStore) AddURL(originalURL string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if record, exists := s.fullStorage[originalURL]; exists {
		return record.ShortURL, nil
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

	err := s.saveStorage()
	if err != nil {
		logrus.WithField("err", err).Error("Error saving json store")
		return "", err
	}
	return shortURL, nil
}

func (s *JSONStore) AddURLs(urls models.BatchRequest) (models.BatchResponse, error) {
	var responses models.BatchResponse

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, req := range urls {
		shortURL := s.gen.Next()

		record := JSONRecord{
			ShortURL:    shortURL,
			OriginalURL: req.OriginalURL,
			UUID:        req.CorrelationID,
		}

		s.storage[shortURL] = record
		s.fullStorage[req.OriginalURL] = record

		err := s.saveStorage()
		if err != nil {
			logrus.WithField("err", err).Error("Error saving json store")
		} else {
			responses = append(responses, models.ItemResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      shortURL,
			})
		}
	}

	return responses, nil
}

func (s *JSONStore) GetOriginalURL(shortURL string) (string, bool) {
	record, exists := s.storage[shortURL]
	return record.OriginalURL, exists
}
