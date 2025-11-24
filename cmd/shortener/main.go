package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
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

func shortenURLCreate(w http.ResponseWriter, r *http.Request, storage *URLStorage) {
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

	shortenID := generateShortID()
	storage.Set(shortenID, originalURL)
	shortURL := "http://" + r.Host + "/" + shortenID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func shortenURLRedirect(w http.ResponseWriter, r *http.Request, storage *URLStorage) {
	id := r.URL.Path[1:]
	if id == "" {
		badRequest(w, "Invalid Path")
		return
	}

	originalURL, exists := storage.Get(id)
	if !exists {
		badRequest(w, "URL not found")
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func run() error {
	storage := NewURLStorage()

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

// in-memory url storage
type URLStorage struct {
	mux  sync.RWMutex
	data map[string]string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		data: make(map[string]string),
	}
}

func (s *URLStorage) Set(id, originalURL string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[id] = originalURL
}

func (s *URLStorage) Get(id string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	url, ok := s.data[id]
	return url, ok
}

// id generator
func generateShortID() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
