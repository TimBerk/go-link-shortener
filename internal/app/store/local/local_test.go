package local

import (
	"context"
	"testing"

	"reflect"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	base "github.com/TimBerk/go-link-shortener/internal/app/store"
)

type MockGenerator struct{}

func (m *MockGenerator) Next() string {
	return "short2"
}

func TestAddURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		store       *URLStore
		originalURL string
		want        string
	}{
		{
			name: "Add new value in empty Store",
			store: &URLStore{
				linksMap:    map[string]UserLink{},
				originalMap: map[string]UserLink{},
				gen:         base.NewIDGenerator(),
			},
			originalURL: "localhost:8080",
			want:        "short2",
		},
		{
			name: "Add new value in Store",
			store: &URLStore{
				linksMap:    map[string]UserLink{"short1": {"test", "localhost:9090"}},
				originalMap: map[string]UserLink{"localhost:9090": {"test", "short1"}},
				gen:         base.NewIDGenerator(),
			},
			originalURL: "localhost:8080",
			want:        "short2",
		},
		{
			name: "Add exist value in Store",
			store: &URLStore{
				linksMap:    map[string]UserLink{"short2": {"test", "localhost:8080"}},
				originalMap: map[string]UserLink{"localhost:8080": {"test", "short2"}},
				gen:         base.NewIDGenerator(),
			},
			originalURL: "localhost:8080",
			want:        "short2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockGen := &base.IDGenerator{}
			patch := monkey.PatchInstanceMethod(
				reflect.TypeOf(mockGen),
				"Next",
				func(*base.IDGenerator) string {
					return "short2"
				},
			)
			defer patch.Unpatch()

			currentLink, _ := test.store.AddURL(ctx, test.originalURL, "test")

			assert.Equal(
				t,
				currentLink,
				test.want,
				"stores.AddURL(%s) must return %t",
				test.originalURL,
				test.want,
			)
		})
	}

}

func TestGetOriginalURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		store       *URLStore
		shortURL    string
		originalURL string
		exists      bool
	}{
		{
			name: "Get url from empty Store",
			store: &URLStore{
				linksMap:    map[string]UserLink{},
				originalMap: map[string]UserLink{},
				gen:         base.NewIDGenerator(),
			},
			shortURL:    "short1",
			originalURL: "",
			exists:      false,
		},
		{
			name: "Get exist value in Store",
			store: &URLStore{
				linksMap:    map[string]UserLink{"short1": {"test", "localhost:9090"}},
				originalMap: map[string]UserLink{"localhost:9090": {"test", "short1"}},
				gen:         base.NewIDGenerator(),
			},
			shortURL:    "short1",
			originalURL: "localhost:9090",
			exists:      true,
		},
		{
			name: "Get not exist value in Store",
			store: &URLStore{
				linksMap:    map[string]UserLink{"short2": {"test", "localhost:8080"}},
				originalMap: map[string]UserLink{"localhost:8080": {"test", "short2"}},
				gen:         base.NewIDGenerator(),
			},
			shortURL:    "short1",
			originalURL: "",
			exists:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalURL, exists, _ := test.store.GetOriginalURL(ctx, test.shortURL, "test")

			assert.Equal(t, test.originalURL, originalURL, "Incorrect original URL for test: %s", test.name)
			assert.Equal(t, test.exists, exists, "Incorrect flag exists for test: %s", test.name)
		})
	}
}
