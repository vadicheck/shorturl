// Package ping provides a handler for checking the availability of the URL storage service.
package ping

import (
	"context"
	"net/http"
	"time"
)

// URLStorage defines the interface for interacting with the URL storage system.
type URLStorage interface {
	PingContext(ctx context.Context) error
}

// New creates a new handler function for checking the availability of the URL storage service.
//
// It sends a ping request to the storage service and returns an HTTP status code based on the result.
// A successful ping returns an HTTP 200 OK status, while a failure returns an HTTP 500 Internal Server Error.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - storage: The URL storage service to check for availability.
//
// Returns:
// - An HTTP handler function that processes the ping request and returns the appropriate status code.
func New(ctx context.Context, storage URLStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		reqCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		res.Header().Set("Content-Type", "application/json")

		if err := storage.PingContext(reqCtx); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		}

		res.WriteHeader(http.StatusOK)
	}
}
