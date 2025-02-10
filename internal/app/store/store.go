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
	GetOriginalURL(shortURL string) (string, bool)
}
