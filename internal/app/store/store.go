package store

import (
	"context"
	"errors"
	"math/rand"

	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
)

const (
	// возмозжные символы для генерации ссылки
	chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	// длина ссылки
	length = 6
)

// ErrLinkExist ошибка о наличии ссылки для исходного адреса
var ErrLinkExist = errors.New("short link exist for original url")

// URLPair параметры для хранения ссылки пользователя
type URLPair struct {
	ShortURL string
	UserID   string
}

// Store интерфейс для обработки основных методов хранилища данных
type Store interface {
	// AddURL генерирует сокращенную ссылку для переданного URL от пользователя
	AddURL(ctx context.Context, originalURL string, userID string) (string, error)
	// AddURLs генерирует сокращенные ссылку для переданных URL от пользователя
	AddURLs(ctx context.Context, urls batch.BatchRequest, userID string) (batch.BatchResponse, error)
	// GetOriginalURL на основании сокращенной ссылки возвращает оригинальную ссылку пользователя
	GetOriginalURL(ctx context.Context, shortURL string, userID string) (string, bool, bool)
	// Ping проверяет подключение к БД
	Ping(ctx context.Context) error
	// DeleteURL удаляет ссылку пользователя
	DeleteURL(ctx context.Context, batch []URLPair) error
}

// IDGenerator генератор ссылок
type IDGenerator struct{}

// Generator интерфейс для генерации ссылок
type Generator interface {
	// Next - получает следующее рандомное сгенерированное значение
	Next() string
}

// NewIDGenerator возвращает новый генератор ссылок
func NewIDGenerator() Generator {
	return &IDGenerator{}
}

// Next генерирует ссылку
func (g *IDGenerator) Next() string {
	id := make([]byte, length)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}

	return string(id)
}
