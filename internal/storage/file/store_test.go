package file_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/storage/file"
)

/*
 Start with the basic interfaces like io.Reader, io.Writer
*/

func TestFileStore(t *testing.T) {

	t.Run("read file and check short code exists", func(t *testing.T) {
		// setup
		shortCode, originalUrl := "abc123", "https://example.com"
		dummyData := map[string]string{shortCode: originalUrl}
		f, _ := os.CreateTemp("", "db")
		defer os.Remove(f.Name()) // clean up
		json.NewEncoder(f).Encode(&dummyData)

		fs := file.NewFileStore(f)

		// execute
		shortCodeExists, err := fs.Exists(shortCode)

		if err != nil {
			t.Fatalf("failed to execute exists: %v", err)
		}

		if !shortCodeExists {
			t.Errorf("could not find short code %q", shortCode)
		}
	})

	t.Run("read file and get original url", func(t *testing.T) {
		// setup
		shortCode, originalUrl := "abc123", "https://example.com"
		dummyData := map[string]string{shortCode: originalUrl}
		f, _ := os.CreateTemp("", "db")
		defer os.Remove(f.Name()) // clean up
		json.NewEncoder(f).Encode(&dummyData)

		fs := file.NewFileStore(f)

		gotUrl, err := fs.GetOriginalURL(shortCode)

		if err != nil {
			t.Fatalf("failed to get original url from store: %v", err)
		}

		if gotUrl != originalUrl {
			t.Errorf("got %q, want %q", gotUrl, originalUrl)
		}

	})

	// t.Run("save short code", func(t *testing.T) {
	// 	shortCode, originalUrl := "abc123", "https://example.com"
	// 	// We need to persist this and create a temp file to write to
	// 	// we will need to modify the interface slightly
	// 	f, _ := os.CreateTemp("", "db")
	// 	f.WriteString(`{}`)
	// 	fs := file.FileStore{Database: f}
	// 	// execute - temp ignoring error
	// 	err := fs.Save(shortCode, originalUrl)

	// 	if err != nil {
	// 		t.Fatalf("failed during save: %v", err)
	// 	}

	// 	// assert - short code exists
	// 	shortCodeExists, err := fs.Exists(shortCode)
	// 	if err != nil {
	// 		t.Fatalf("failed to check for existance: %v", err)
	// 	}
	// 	if !shortCodeExists {
	// 		t.Errorf("failed to persist short code: %q", shortCode)
	// 	}
	// 	defer os.Remove(f.Name()) // clean up
	// })

}
