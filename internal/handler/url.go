package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/go-chi/chi/v5"
)

func (h *URLHandlers) CreateJSON(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	defer r.Body.Close()
	if err != nil {
		BadRequest(w, "wrong Content-Type")
		return
	}
	if mediaType != "application/json" {
		BadRequest(w, "Content-Type must be application/json")
		return
	}

	var body requestURL
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}
	if err := validateURL(body.URL); err != nil {
		BadRequest(w, err.Error())
		return
	}

	shortURL, created, err := generateAndStoreShortURL(body.URL, h)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	status := http.StatusConflict
	if created {
		status = http.StatusCreated
	}

	payload := responseURL{Result: shortURL}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: cannot marshal response payload: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(b); err != nil {
		log.Printf("ERROR: cannot write response body: %v", err)
		return
	}
}

func (h *URLHandlers) CreateBatch(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	defer r.Body.Close()
	if err != nil {
		BadRequest(w, "wrong Content-Type")
		return
	}
	if mediaType != "application/json" {
		BadRequest(w, "Content-Type must be application/json")
		return
	}

	var body []repository.BatchItemInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}
	if len(body) == 0 {
		BadRequest(w, "Empty request body")
		return
	}

	// return if even one url is invalid
	for _, item := range body {
		if err := validateURL(item.OriginalURL); err != nil {
			BadRequest(w, "invalid URL in a batch: "+item.OriginalURL)
			return
		}
	}

	const (
		maxBatchAttempts = 10
		maxIDAttempts    = 10
	)

	for attempt := 0; attempt < maxBatchAttempts; attempt++ {
		// 1) Собираем batch + payload (только генерация, без обращений к storage)
		seenIDs := make(map[string]struct{}, len(body))

		batch := make([]repository.URLItem, 0, len(body))
		payload := make([]repository.BatchItemOutput, 0, len(body))

		for _, item := range body {
			// генерируем id, уникальный внутри запроса
			var id string
			unique := false

			for i := 0; i < maxIDAttempts; i++ {
				id, err = h.generator.GenerateShortID()
				if err != nil {
					log.Printf("cannot generate shorten ID: %v", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				if id == "" {
					log.Printf("ERROR: generator returned empty id")
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				if _, ok := seenIDs[id]; ok {
					continue
				}
				unique = true
				break
			}

			if !unique {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			seenIDs[id] = struct{}{}

			shortURL, err := url.JoinPath(h.baseURL, id)
			if err != nil {
				log.Printf("cannot build shorten URL (baseURL=%q, id=%q): %v", h.baseURL, id, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			payload = append(payload, repository.BatchItemOutput{
				CorrelationID: item.CorrelationID,
				ShortURL:      shortURL,
			})
			batch = append(batch, repository.URLItem{
				ID:          id,
				OriginalURL: item.OriginalURL,
			})
		}

		// 2) Пытаемся вставить целиком
		err = h.storage.CreateBatch(batch)

		switch {
		case err == nil:
			b, err := json.Marshal(payload)
			if err != nil {
				log.Printf("ERROR: cannot marshal response payload: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write(b); err != nil {
				log.Printf("ERROR: cannot write response body: %v", err)
			}
			return

		case errors.Is(err, repository.ErrIDAlreadyExists):
			// коллизия short_id → просто повторяем весь батч с новыми id
			log.Printf("INFO: short_id collision in batch attempt %d/%d: %v", attempt+1, maxBatchAttempts, err)
			continue

		case errors.Is(err, repository.ErrURLAlreadyExists):
			// уже существующий original_url → не ретраим, иначе будет вечный цикл
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return

		default:
			log.Printf("ERROR: cannot insert batch: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (h *URLHandlers) Create(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	defer r.Body.Close()
	if err != nil {
		BadRequest(w, "wrong Content-Type")
		return
	}
	if mediaType != "text/plain" {
		BadRequest(w, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}
	if err := validateURL(string(body)); err != nil {
		BadRequest(w, err.Error())
		return
	}

	shortURL, created, err := generateAndStoreShortURL(string(body), h)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	status := http.StatusConflict
	if created {
		status = http.StatusCreated
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(shortURL))
}

func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var id string
	// Chi params is only set when Chi router is configured
	// but in tests we don't use Chi router, just raw methods
	if paramID := chi.URLParam(r, "id"); paramID != "" {
		id = paramID
	} else {
		id = r.URL.Path[1:]
	}

	if id == "" {
		BadRequest(w, "Invalid Path")
		return
	}

	originalURL, err := h.storage.GetURLByID(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
