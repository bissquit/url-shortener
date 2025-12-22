package compress

import (
	"compress/gzip"
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
	// headers should be set before calling WriteHeader()
	// when it's not, http will set httpOk
	if g.Header().Get("Content-Type") == "" {
		log.Println("gzip: missing Content-Type")
		g.compress = false
		return
	}

	ct := strings.ToLower(g.Header().Get("Content-Type"))
	supportJSON := strings.Contains(ct, "application/json")
	supportHTML := strings.Contains(ct, "text/html")

	if supportJSON || supportHTML {
		g.compress = true
		g.Header().Set("Content-Encoding", "gzip")
		g.Header().Del("Content-Length")
	} else {
		g.compress = false
	}

	g.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipWriter) Close() error {
	if g.compress {
		return g.gw.Close()
	}
	return nil
}

func GzipResponse(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := strings.ToLower(r.Header.Get("Accept-Encoding"))
		if !strings.Contains(encoding, "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		cw := newGzipWriter(w)
		defer cw.Close()
		h.ServeHTTP(&cw, r)
	})
}
