package main

import (
	"fmt"
	"net/http"

	"github.com/bissquit/url-shortener/internal/handler"
	"github.com/bissquit/url-shortener/internal/repository/memory"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	storage := memory.NewURLStorage()
	handlers := handler.NewURLHandlers(storage, "http://localhost:8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/" && r.Method == http.MethodPost:
			handlers.Create(w, r)
		case r.URL.Path != "/" && r.Method == http.MethodGet:
			handlers.Redirect(w, r)
		default:
			handler.BadRequest(w, fmt.Sprintf("path %s, method %s", r.URL.Path, r.Method))
		}
	})
	return http.ListenAndServe(`:8080`, mux)
}
