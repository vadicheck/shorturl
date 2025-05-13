// Package error provides utility functions for handling and responding with error messages.
package error

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/models/shorten"
)

// RespondWithError sends an HTTP response with a JSON-encoded error message.
//
// This function sets the appropriate HTTP status code and Content-Type header, then encodes and sends
// the error message as a JSON response body. If the encoding fails, it logs the error.
//
// Parameters:
// - w: The `http.ResponseWriter` used to send the HTTP response.
// - statusCode: The HTTP status code to set for the response.
// - message: The error message to be included in the response body as part of the error structure.
//
// Example usage:
//
//	http.Error(w, "Invalid request", http.StatusBadRequest)
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(shorten.NewError(message)); encodeErr != nil {
		slog.Error(fmt.Sprintf("cannot encode response JSON body: %s", encodeErr))
	}
}
