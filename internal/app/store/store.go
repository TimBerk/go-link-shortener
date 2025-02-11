package store

import (
	"errors"
	"math/rand"

	"github.com/TimBerk/go-link-shortener/internal/app/models/batch"
)

const (
	chars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	length = 6
)

var ErrLinkExist = errors.New("short link exist for original url")

type IDGenerator struct{}

type Generator interface {
	Next() string
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

type MainStoreInterface interface {
	AddURL(originalURL string) (string, error)
	AddURLs(urls batch.BatchRequest) (batch.BatchResponse, error)
	GetOriginalURL(shortURL string) (string, bool)
}
