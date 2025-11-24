package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

func shortenURLCreate(w http.ResponseWriter, r *http.Request) {
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

	// TODO: create handler
	shortURL := "http://localhost:8080/EwHXdJfB"

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func shortenURLRedirect(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	if id == "" {
		badRequest(w, "Invalid Path")
		return
	}

	// TODO: create handler
	originalURL := "https://practicum.yandex.ru/"

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/" && r.Method == http.MethodPost:
			shortenURLCreate(w, r)
		case r.URL.Path != "/" && r.Method == http.MethodGet:
			shortenURLRedirect(w, r)
		default:
			badRequest(w, fmt.Sprintf("path %s, method %s", r.URL.Path, r.Method))
		}
	})
	return http.ListenAndServe(`:8080`, mux)
}
