package handler

import (
	"errors"
)

type errorReader struct{}

// implement dummy method to emulate io.Reader error
func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("dummy error")
}

type DummyGenerator struct {
	id  string
	err error
}

func NewDummyGenerator() *DummyGenerator {
	return &DummyGenerator{}
}

// implement dummy method  to emulate ID generation error
func (g *DummyGenerator) GenerateShortID() (string, error) {
	var err error
	if g.err != nil {
		err = g.err
	}
	return g.id, err
}
