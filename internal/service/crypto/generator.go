package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Generator struct{}

func NewRandomGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateShortID() (string, error) {
	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string for short ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
