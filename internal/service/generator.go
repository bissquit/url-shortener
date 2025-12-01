package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/bissquit/url-shortener/internal/repository"
)

func generateShortID() (string, error) {
	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateUniqID(storage repository.URLRepository) (string, error) {
	maxAttempts := 10
	for range maxAttempts {
		id, err := generateShortID()
		if err != nil {
			return "", err
		}

		_, err = storage.Get(id)
		if errors.Is(err, repository.ErrNotFound) {
			return id, nil
		}

		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return "", fmt.Errorf("storage error: %w", err)
		}
	}
	return "", fmt.Errorf("failed to generate unique ID after %d attempts", maxAttempts)
}
