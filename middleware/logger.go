// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// ResponseWriter wraps the response so we can capture status code and size.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

// NewResponseWriter creates a new ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the HTTP status code for logging.
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body size for logging.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n

	return n, err
}

// Unwrap returns the wrapped response, so http.ResponseController can flush streamed responses through it.
func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Logger creates a new logger middleware
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()

		wrapped := NewResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		elapsed := time.Since(start)
		if !IsAssetPath(r.URL.Path) {
			log.Info().
				Str("request_id", GetRequestId(r.Context())).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.statusCode).
				Dur("duration", elapsed).
				Msg("Request completed")
		}
	})
}
