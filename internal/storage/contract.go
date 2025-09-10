package storage

import "testing"

type URLStoreContract struct {
	NewStore func() URLStore
}

// Now we can enforce our contract on the interface and test across multiple implementations
// to make sure behaviour is consistent

func (u URLStoreContract) Test(t *testing.T) {
	t.Run("can save, check for existence and get original URL", func(t *testing.T) {
		// setup
		store := u.NewStore()
		shortCode, originalUrl := "abc123", "https://example.com"

		// execute - storing data
		err := store.Save(shortCode, originalUrl)

		if err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		// assert
		exists := store.Exists(shortCode)
		if !exists {
			t.Fatalf("could not find short code %q", shortCode)
		}

		gotUrl, found := store.GetOriginalURL(shortCode)

		if !found {
			t.Error("original url should be found")
		}
		if gotUrl != originalUrl {
			t.Errorf("got %q, want %q", gotUrl, originalUrl)
		}
	})
}
