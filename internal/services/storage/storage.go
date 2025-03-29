// Package storage provides common errors and types related to URL storage operations.
// It defines error types and helpers for handling cases where URLs or short codes already exist
// in the storage system.
package storage

import (
	"errors"
	"fmt"
)

// ErrURLOrCodeExists is an error that is returned when a URL or a short code already exists in the storage.
var ErrURLOrCodeExists = errors.New("url or code exists")

// ExistsURLError is an error type that provides details about an existing URL or short code conflict.
// It includes the original URL, the conflicting short code, and the underlying error that caused the conflict.
type ExistsURLError struct {
	// OriginalURL is the URL that caused the conflict.
	OriginalURL string

	// ShortCode is the short code that caused the conflict.
	ShortCode string

	// Err is the underlying error that explains the cause of the conflict.
	Err error
}

// Error implements the error interface for ExistsURLError.
// It returns a string representation of the error that includes the original URL, the short code, and the underlying error.
func (e *ExistsURLError) Error() string {
	return fmt.Sprintf("[%s:%s] %v", e.OriginalURL, e.ShortCode, e.Err)
}
