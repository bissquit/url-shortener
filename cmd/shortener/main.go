package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/repository/memory"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func badRequest(w http.ResponseWriter, message string) {
	log.Printf("bad request: %s", message)
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func shortenURLCreate(w http.ResponseWriter, r *http.Request, storage repository.URLRepository) {
	if r.Header.Get("Content-Type") != "text/plain" {
		badRequest(w, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		badRequest(w, "Cannot read request body")
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		badRequest(w, "Invalid URL")
		return
	}

	shortenID, err := generateUniqID(storage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	storage.Set(shortenID, originalURL)
	shortURL := "http://" + r.Host + "/" + shortenID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func shortenURLRedirect(w http.ResponseWriter, r *http.Request, storage repository.URLRepository) {
	id := r.URL.Path[1:]
	if id == "" {
		badRequest(w, "Invalid Path")
		return
	}

	originalURL, exists := storage.Get(id)
	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func run() error {
	storage := memory.NewURLStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/" && r.Method == http.MethodPost:
			shortenURLCreate(w, r, storage)
		case r.URL.Path != "/" && r.Method == http.MethodGet:
			shortenURLRedirect(w, r, storage)
		default:
			badRequest(w, fmt.Sprintf("path %s, method %s", r.URL.Path, r.Method))
		}
	})
	return http.ListenAndServe(`:8080`, mux)
}

// id generator
func generateShortID() (string, error) {
	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateUniqID(storage repository.URLRepository) (string, error) {
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
