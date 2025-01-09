package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddURL(t *testing.T) {
	tests := []struct {
		name        string
		store       URLStore
		originalURL string
		want        string
	}{
		{
			name:        "Add new value in empty Store",
			store:       URLStore{linksMap: map[string]string{}},
			originalURL: "localhost:8080",
			want:        "short2",
		},
		{
			name: "Add new value in Store",
			store: URLStore{
				linksMap: map[string]string{"short1": "localhost:9090"},
			},
			originalURL: "localhost:8080",
			want:        "short2",
		},
		{
			name: "Add exist value in Store",
			store: URLStore{
				linksMap: map[string]string{"short2": "localhost:8080"},
			},
			originalURL: "localhost:8080",
			want:        "short2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalGenerateIDFunc := generateIDFunc
			generateIDFunc = func() string {
				return "short2"
			}
			defer func() {
				generateIDFunc = originalGenerateIDFunc
			}()

			assert.Equal(
				t,
				test.store.AddURL(test.originalURL),
				test.want,
				"stores.AddURL(%s) must return %t",
				test.originalURL,
				test.want,
			)
		})
	}

}

func TestGetOriginalURL(t *testing.T) {
	tests := []struct {
		name        string
		store       URLStore
		shortURL    string
		originalURL string
		exists      bool
	}{
		{
			name:        "Get url from empty Store",
			store:       URLStore{linksMap: map[string]string{}},
			shortURL:    "short1",
			originalURL: "",
			exists:      false,
		},
		{
			name: "Get exist value in Store",
			store: URLStore{
				linksMap: map[string]string{"short1": "localhost:9090"},
			},
			shortURL:    "short1",
			originalURL: "localhost:9090",
			exists:      true,
		},
		{
			name: "Get not exist value in Store",
			store: URLStore{
				linksMap: map[string]string{"short2": "localhost:8080"},
			},
			shortURL:    "short1",
			originalURL: "",
			exists:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalURL, exists := test.store.GetOriginalURL(test.shortURL)

			assert.Equal(t, test.originalURL, originalURL, "Incorrect original URL for test: %s", test.name)
			assert.Equal(t, test.exists, exists, "Incorrect flag exists for test: %s", test.name)
		})
	}
}
