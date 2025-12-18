package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/service"
)

type URLHandlers struct {
	storage   repository.URLRepository
	baseURL   string
	generator service.IDGenerator
}

func NewURLHandlers(storage repository.URLRepository, baseURL string, generator service.IDGenerator) *URLHandlers {
	return &URLHandlers{
		storage:   storage,
		baseURL:   baseURL,
		generator: generator,
	}
}

type requestURL struct {
	URL string `json:"url"`
}

type responseURL struct {
	Result string `json:"result"`
}

func generateAndStoreShortURL(originalURL string, h *URLHandlers) (string, int) {
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		log.Printf("%s: %s", http.StatusText(http.StatusBadRequest), "Invalid URL")
		return "", http.StatusBadRequest
	}

	var shortenID string
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		// trying to generate short ID
		id, err := h.generator.GenerateShortID()
		if err != nil {
			log.Printf("ERROR: cannot generate shorten ID: %v", err)
			//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return "", http.StatusInternalServerError
		}

		// trying to save ID
		err = h.storage.Create(id, originalURL)
		if err == nil {
			shortenID = id
			// exit if success
			break
		}

		// go next loop iteration if ID is already exist
		if errors.Is(err, repository.ErrAlreadyExists) {
			log.Printf("INFO: shorten ID collision detected in generation attempt %d (max %d): %v", i+1, maxAttempts, err)
			continue
		}

		// raise unknown error just in case if break and continue fail before
		log.Printf("ERROR: unknown storage error: %v", err)
		//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return "", http.StatusInternalServerError
	}

	if shortenID == "" {
		log.Printf("ERROR: failed to generate unique ID after %d attempts", maxAttempts)
		//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return "", http.StatusInternalServerError
	}

	shortURL, err := url.JoinPath(h.baseURL, shortenID)
	if err != nil {
		log.Printf("ERROR: cannot return shorten URL (baseURL=%q, id=%q): %v",
			h.baseURL, shortenID, err)
		//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return "", http.StatusInternalServerError
	}

	return shortURL, 0
}

func (h *URLHandlers) CreateJSON(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	defer r.Body.Close()
	if err != nil {
		BadRequest(w, "wrong Content-Type")
		return
	}

	if mediaType != "application/json" {
		BadRequest(w, "Content-Type not application/json")
		return
	}

	var body requestURL

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}

	if body.URL == "" {
		BadRequest(w, "Empty URL value in request body")
		return
	}

	shortURL, code := generateAndStoreShortURL(body.URL, h)
	if code != 0 {
		http.Error(w, http.StatusText(code), code)
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

func (h *URLHandlers) Create(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		BadRequest(w, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}

	shortURL, code := generateAndStoreShortURL(string(body), h)
	if code != 0 {
		http.Error(w, http.StatusText(code), code)
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

func BadRequest(w http.ResponseWriter, message string) {
	log.Printf("bad request: %s", message)
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}
