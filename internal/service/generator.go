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
	for i := 0; i < maxAttempts; i++ {
		id, err := generateShortID()
		if err != nil {
			return "", err
		}

		if _, exists := storage.Get(id); !exists {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique ID")
}
