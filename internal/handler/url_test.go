package handler

import (
	"errors"
	"fmt"
)

type errorReader struct{}

// implement dummy method to emulate io.Reader error
func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("dummy error")
}

type DummyGenerator struct{}

func NewDummyGenerator() *DummyGenerator {
	return &DummyGenerator{}
}

// implement dummy method  to emulate short ID generation error
func (g *DummyGenerator) GenerateShortID() (string, error) {
	return "", fmt.Errorf("fake error: %w", errors.New("dummy error"))
}
