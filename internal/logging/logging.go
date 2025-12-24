package logging

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

type (
	// create struct to store data from response
	responseData struct {
		status int
		size   int
	}

	// add http.ResponseWriter implementation
	loggingResponseWriter struct {
		http.ResponseWriter // embed original http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.responseData.status == 0 {
		// If the handler does not call WriteHeader explicitly, the status will remain 0.
		// The default should be 200 OK.
		r.responseData.status = http.StatusOK
	}
	// write response using original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // catch size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// If WriteHeader is called multiple times (which is possible with some errors),
	// the status will be overwritten and the original ResponseWriter will receive
	// WriteHeader multiple times, which is prohibited.
	if r.responseData.status == 0 {
		r.responseData.status = statusCode
		// write response code using original http.ResponseWriter
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		uri := r.RequestURI // get uri
		method := r.Method  // get method

		h.ServeHTTP(lw, r) // handle original request

		duration := time.Since(start) // count request duration

		sugar.Infow("HTTP request:",
			"uri", uri,
			"method", method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}

// - called automatically when a package is initialized
// - there can be multiple init() functions in a single package (executed in the order they are declared)
// - executed before the main() function
// - used to initialize global state
func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Printf("Failed to create zap logger: %v", err)
		logger = zap.NewNop()
	}
	sugar = logger.Sugar()
}
