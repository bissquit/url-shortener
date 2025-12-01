package service

import (
	"crypto/rand"
	"encoding/hex"
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
		if err != nil {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique ID after %d attempts", maxAttempts)
}
