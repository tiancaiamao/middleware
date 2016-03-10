// Package gzip provides a simple middleware layer that performs
// gzip compression on the response.
package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Gzip is a middleware type which gzips HTTP responses. It is
// imperative that any handler which writes to a gzipped response
// specifies the Content-Type, otherwise some clients will assume
// application/x-gzip and try to download a file.
type Gzip struct{}

func New() Gzip {
	return Gzip{}
}

func (g Gzip) Chain(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		// Delete this header so gzipping is not repeated later in the chain
		r.Header.Del("Accept-Encoding")

		w.Header().Set("Content-Encoding", "gzip")
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()
		gz := gzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}

		// Any response in forward middleware will now be compressed
		h.ServeHTTP(gz, r)
	})
}

// gzipResponeWriter wraps the underlying Write method
// with a gzip.Writer to compress the output.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// WriteHeader wraps the underlying WriteHeader method to prevent
// problems with conflicting headers from proxied backends. For
// example, a backend system that calculates Content-Length would
// be wrong because it doesn't know it's being gzipped.
func (w gzipResponseWriter) WriteHeader(code int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(code)
}

// Write wraps the underlying Write method to do compression.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	n, err := w.Writer.Write(b)
	return n, err
}
