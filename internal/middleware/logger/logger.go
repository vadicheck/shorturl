// Package logger provides a middleware for logging HTTP request and response details.
// It logs information about the request's URI, method, response status, size, and the duration of the request handling.
package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type (
	// responseData holds the status code and size of the response.
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter is a custom ResponseWriter that allows capturing response status and size.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write captures the response size while writing the response body.
// It overrides the default Write method of the ResponseWriter to track the size of the response.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader captures the HTTP status code while writing the response header.
// It overrides the default WriteHeader method of the ResponseWriter.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// New returns a middleware function for logging HTTP request details.
//
// The middleware logs the following information for each incoming request:
//   - Request URI
//   - HTTP Method (e.g., GET, POST)
//   - Duration of request handling
//   - Response status code
//   - Response size (in bytes)
//
// The logged message is recorded using the `slog` logging package.
//
// Parameters:
//   - None (this is a middleware factory function)
//
// Returns:
//   - A middleware function that can be used with `http.Handle` or other HTTP routers.
func New() func(next http.Handler) http.Handler {
	slog.Info("logger middleware enabled")

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			message := fmt.Sprintf("uri: %s, method: %s, duration: %s, status: %d, size: %d",
				r.RequestURI,
				r.Method,
				duration,
				responseData.status,
				responseData.size,
			)

			slog.Info(message)
		}
		return http.HandlerFunc(fn)
	}
}
