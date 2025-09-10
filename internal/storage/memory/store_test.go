package memory_test

import (
	"errors"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/storage"
	"github.com/sotiri-geo/url-shortener/internal/storage/memory"
)

func TestMemoryDBStore(t *testing.T) {
	t.Run("get original url from short code", func(t *testing.T) {
		store := memory.NewWithData(map[string]string{"abc123": "https://example.com"})

		got, exists := store.GetOriginalURL("abc123")
		want := "https://example.com"
		if !exists {
			t.Fatal("url should exist in store")
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("url does not exist in store", func(t *testing.T) {
		store := memory.New()

		_, found := store.GetOriginalURL("abc123")

		if found {
			t.Fatal("should not find url")
		}
	})

	t.Run("url should persist", func(t *testing.T) {
		store := memory.New()
		shortCode := "abc123"
		want := "https://example.com"
		store.Save(shortCode, want)

		got, exists := store.GetOriginalURL(shortCode)

		if !exists {
			t.Fatalf("short url %q should exist", shortCode)
		}

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}

	})

	t.Run("handles conflicting short code", func(t *testing.T) {
		store := memory.New()
		shortCode, originalUrl := "abc123", "https://example.com"
		store.Save(shortCode, originalUrl)

		// should fail
		err := store.Save(shortCode, originalUrl)

		if err == nil {
			t.Fatal("failed to raise conflicting error")
		}
		// integrity error
		if !errors.Is(err, memory.ErrShortCodeExists) {
			t.Errorf("short url %q already exists: %v", shortCode, err)
		}
	})

	t.Run("checks existence of short code", func(t *testing.T) {
		store := memory.New()
		shortCode, originalUrl := "abc123", "https://example.com"
		err := store.Save(shortCode, originalUrl)

		if err != nil {
			t.Fatal("should not fail during save")
		}

		exists := store.Exists(shortCode)

		if !exists {
			t.Error("should exist in store")
		}
	})
}

// Used for contract testing
func TestMemoryContract(t *testing.T) {

	storage.URLStoreContract{
		NewStore: func() storage.URLStore { return memory.New() },
	}.Test(t)
}
