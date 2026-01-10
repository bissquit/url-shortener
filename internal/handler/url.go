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

	shortURL, err := generateAndStoreShortURL(body.URL, h)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// prepare response payload separate from HTTP writing
	payload := responseURL{Result: shortURL}
	// convert the payload to JSON bytes before sending any headers/status
	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: cannot marshal response payload: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// start HTTP response only after JSON is ready
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(b); err != nil {
		// if writing body fails, status code is already sent, so we can only log the error
		// it doesn't make sense to send 5xx status after status is set and already sent above
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

	var badItemsCount int
	for _, item := range body {
		if err := validateURL(item.OriginalURL); err != nil {
			badItemsCount++
		}
	}
	if badItemsCount > 0 {
		BadRequest(w, "one or more item(s) in a batch has incorrect url")
		return
	}

	var (
		maxAttempts  int = 10
		id, shortURL string
		payload      []repository.BatchItemOutput
		batch        []repository.URLItem
	)
	for _, item := range body {
		id = ""
		for i := 0; i < maxAttempts; i++ {
			// trying to generate short ID
			id, err = h.generator.GenerateShortID()
			if err != nil {
				log.Printf("cannot generate shorten ID: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			// trying to retrieve id from storage
			_, err = h.storage.Get(id)
			if errors.Is(err, repository.ErrNotFound) {
				break
			} else if err != nil {
				log.Printf("unknown storage error: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			} else {
				continue
			}
		}
		if id == "" {
			log.Printf("ERROR: cannot generate id: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// prepare output
		shortURL, err = url.JoinPath(h.baseURL, id)
		if err != nil {
			log.Printf("cannot return shorten URL (baseURL=%q, id=%q): %v", h.baseURL, id, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		outputItem := repository.BatchItemOutput{
			CorrelationId: item.CorrelationId,
			ShortURL:      shortURL,
		}
		payload = append(payload, outputItem)

		// prepare batch
		batchItem := repository.URLItem{
			OriginalURL: item.OriginalURL,
			Id:          id,
		}
		batch = append(batch, batchItem)
	}

	err = h.storage.BatchCreate(batch)
	if err != nil {
		log.Printf("ERROR: cannot store batch: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: cannot marshal response payload: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// start HTTP response only after JSON is ready
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(b); err != nil {
		// if writing body fails, status code is already sent, so we can only log the error
		// it doesn't make sense to send 5xx status after status is set and already sent above
		log.Printf("ERROR: cannot write response body: %v", err)
		return
	}
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

	shortURL, err := generateAndStoreShortURL(string(body), h)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	// write body, alternative for fmt.Fprint(w, shortURL)
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

	originalURL, err := h.storage.Get(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
