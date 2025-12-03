package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateShortID() (string, error) {
	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string for short ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
