// Package json предназначен для организации хранения данных в JSON-файле
package json

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	models "github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

// JSONRecord описывает структуру JSON-записи
type JSONRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

// JSONStore описывает структуру JSON-стора
type JSONStore struct {
	storage     map[string]JSONRecord
	fullStorage map[string]JSONRecord
	filePath    string
	gen         store.Generator
	mutex       sync.Mutex
}

// NewJSONStore на основании переданного пути и генератора создает новый JSON-стор
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

// loadStorage осуществляет загрузку и декодирование записей из файла
func (s *JSONStore) loadStorage() error {
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer utils.CloseWithLog(file, "Error closing JSON-file")

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

// saveStorage осуществляет сохранение записей в файл
func (s *JSONStore) saveStorage() error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer utils.CloseWithLog(file, "Error closing JSON-file")

	encoder := json.NewEncoder(file)
	for _, entry := range s.storage {
		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}
	return nil
}

// AddURL осуществляет добавление с генерацией короткой ссылки для пользователя
func (s *JSONStore) AddURL(ctx context.Context, originalURL string, userID string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	record, exists := s.fullStorage[originalURL]
	if exists && record.UserID == userID {
		return record.ShortURL, store.ErrLinkExist
	}

	shortURL := s.gen.Next()

	record, exists = s.storage[shortURL]
	if exists && record.UserID == userID {
		return s.AddURL(ctx, originalURL, userID)
	}

	record = JSONRecord{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UUID:        uuid.New().String(),
		UserID:      userID,
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

// AddURLs осуществляет добавление с генерацией коротких ссылок для пользователя
func (s *JSONStore) AddURLs(ctx context.Context, urls models.BatchRequest, userID string) (models.BatchResponse, error) {
	var responses models.BatchResponse

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, req := range urls {
		shortURL := s.gen.Next()

		record := JSONRecord{
			ShortURL:    shortURL,
			OriginalURL: req.OriginalURL,
			UUID:        req.CorrelationID,
			UserID:      userID,
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

// GetOriginalURL осуществляет поиск оригинальной ссылки по переданной короткой
func (s *JSONStore) GetOriginalURL(ctx context.Context, shortURL string, userID string) (string, bool, bool) {
	record, exists := s.storage[shortURL]
	return record.OriginalURL, exists, false
}

// Ping эмулирует проверку доступности стора
func (s *JSONStore) Ping(ctx context.Context) error {
	return nil
}

// DeleteURL удаляет ссылки пользователя
func (s *JSONStore) DeleteURL(ctx context.Context, batch []store.URLPair) error {
	if len(batch) == 0 {
		return nil
	}

	for _, pair := range batch {
		userLink, exists := s.storage[pair.ShortURL]
		if exists && userLink.UserID == pair.UserID {
			delete(s.fullStorage, userLink.OriginalURL)
			delete(s.storage, pair.ShortURL)
		}
	}

	return nil
}
