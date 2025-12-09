package handler

import (
	"errors"
)

type errorReader struct{}

// implement dummy method to emulate io.Reader error
func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("dummy error")
}
