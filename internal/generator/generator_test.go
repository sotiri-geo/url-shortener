package generator_test

import (
	"testing"

	"github.com/sotiri-geo/url-shortener/internal/generator"
)

func TestRandomGenerator(t *testing.T) {
	t.Run("generates a random string of length 6", func(t *testing.T) {
		gen := generator.New(generator.RandomGenSize)
		got := gen.Generate()
		if len(got) != generator.RandomGenSize {
			t.Errorf("got length %d, want %d", len(got), generator.RandomGenSize)
		}
	})
}
