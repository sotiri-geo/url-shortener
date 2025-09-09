package memory_test

import (
	"errors"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/storage/memory"
)

func TestMemoryDBStore(t *testing.T) {
	t.Run("get original url from short url", func(t *testing.T) {
		store := memory.NewWithData(map[string]string{"abc123": "https://example.com"})

		got, exists := store.GetOriginalUrl("abc123")
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

		_, found := store.GetOriginalUrl("abc123")

		if found {
			t.Fatal("should not find url")
		}
	})

	t.Run("url should persist", func(t *testing.T) {
		store := memory.New()
		shortUrl := "abc123"
		want := "https://example.com"
		store.Save(shortUrl, want)

		got, exists := store.GetOriginalUrl(shortUrl)

		if !exists {
			t.Fatalf("short url %q should exist", shortUrl)
		}

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}

	})

	t.Run("handles conflicting short urls", func(t *testing.T) {
		store := memory.New()
		shortUrl, originalUrl := "abc123", "https://example.com"
		store.Save(shortUrl, originalUrl)

		// should fail
		err := store.Save(shortUrl, originalUrl)

		if err == nil {
			t.Fatal("failed to raise conflicting error")
		}
		// integrity error
		if !errors.Is(err, memory.ErrShortUrlExists) {
			t.Errorf("short url %q already exists: %v", shortUrl, err)
		}
	})
}
