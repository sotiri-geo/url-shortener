package generator

import "crypto/rand"

type Generator interface {
	Generate() string
}

const RandomGenSize = 6

type RandomChars struct {
	length int
}

func (r *RandomChars) Generate() string {
	return rand.Text()[:r.length]
}

func New(length int) *RandomChars {
	return &RandomChars{length: length}
}
