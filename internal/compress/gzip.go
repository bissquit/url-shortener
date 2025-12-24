package compress

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	// pure embedding
	http.ResponseWriter
	gw       *gzip.Writer
	compress bool
}

func newGzipWriter(w http.ResponseWriter) gzipWriter {
	return gzipWriter{
		ResponseWriter: w,
		gw:             gzip.NewWriter(w),
		compress:       false,
	}
}

func (g *gzipWriter) Header() http.Header {
	return g.ResponseWriter.Header()
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	if g.compress {
		return g.gw.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipWriter) WriteHeader(statusCode int) {
	g.compress = false
	// headers should be set before calling WriteHeader() in handlers
	// when it's not, net/http will set httpOk by default
	if g.Header().Get("Content-Type") == "" {
		log.Println("gzip: missing Content-Type")
		g.ResponseWriter.WriteHeader(statusCode)
		return
	}

	ct := strings.ToLower(g.Header().Get("Content-Type"))
	supportJSON := strings.Contains(ct, "application/json")
	supportHTML := strings.Contains(ct, "text/html")

	if supportJSON || supportHTML {
		g.compress = true
		g.Header().Set("Content-Encoding", "gzip")
		g.Header().Del("Content-Length")
	}

	g.ResponseWriter.WriteHeader(statusCode)
}

// ResponseWriter doesn't have Close() method (https://pkg.go.dev/net/http#ResponseWriter)
// but we should close gzip.Writer (https://pkg.go.dev/compress/gzip#Writer.Close)
//
// Close() closes the Writer by flushing any unwritten data to the underlying io.Writer
// and writing the GZIP footer. It does not close the underlying io.Writer.
func (g *gzipWriter) Close() error {
	if g.compress {
		return g.gw.Close()
	}
	return nil
}

func GzipResponse(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := strings.ToLower(r.Header.Get("Accept-Encoding"))
		// TODO: implement split parsing
		if !strings.Contains(encoding, "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		cw := newGzipWriter(w)
		defer cw.Close()
		h.ServeHTTP(&cw, r)
	})
}

type gzipReader struct {
	body io.ReadCloser
	zr   *gzip.Reader
}

func newGzipReader(body io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		body: body,
		zr:   zr,
	}, nil
}

func (g *gzipReader) Read(b []byte) (int, error) {
	return g.zr.Read(b)
}

func (g *gzipReader) Close() error {
	// gzip.Reader.Close() does not close the underlying reader
	// so we should close both
	err1 := g.zr.Close()
	err2 := g.body.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func GzipRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := strings.ToLower(r.Header.Get("Content-Encoding"))
		useGzip := strings.Contains(encoding, "gzip")

		// Content-Encoding is not empty bot doesn't contain gzip format
		if !useGzip && encoding != "" {
			log.Printf("gzip: unsupported Content-Encoding: %q", encoding)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		// TODO: implement split parsing
		if !useGzip {
			h.ServeHTTP(w, r)
			return
		}

		cr, err := newGzipReader(r.Body)
		if err != nil {
			log.Printf("gzip: invalid request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// defer should be only after 'if err...' because in case of error
		// cr returns nill, so defer will call nil pointer
		defer cr.Close()

		r.Body = cr
		r.Header.Del("Content-Length")
		r.Header.Del("Content-Encoding")

		h.ServeHTTP(w, r)
	})
}
