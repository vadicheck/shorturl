// Package shorten defines the data structures and response models related to URL shortening
// as well as helper functions for handling errors.
package shorten

// CreateURLRequest represents the request body for creating a shortened URL.
// It contains the original URL that needs to be shortened.
type CreateURLRequest struct {
	// URL is the original URL to be shortened.
	URL string `json:"url"`
}

// CreateURLResponse represents the response body when a single URL is successfully shortened.
// It contains the shortened URL.
type CreateURLResponse struct {
	// Result is the shortened URL.
	Result string `json:"result"`
}

// CreateBatchURLRequest represents the request body for creating multiple shortened URLs in batch.
// It contains a correlation ID for tracking and the original URL to be shortened.
type CreateBatchURLRequest struct {
	// CorrelationID is a unique ID for tracking the batch request.
	CorrelationID string `json:"correlation_id"`

	// OriginalURL is the original URL to be shortened.
	OriginalURL string `json:"original_url"`
}

// CreateBatchURLResponse represents the response body when a batch of URLs has been successfully shortened.
// It contains the correlation ID for tracking and the shortened URL.
type CreateBatchURLResponse struct {
	// CorrelationID is a unique ID for tracking the batch request.
	CorrelationID string `json:"correlation_id"`

	// ShortURL is the shortened version of the original URL.
	ShortURL string `json:"short_url"`
}

// UserURLResponse represents the response body when a user's URLs are retrieved.
// It contains the shortened URL and its original counterpart.
type UserURLResponse struct {
	// ShortURL is the shortened URL.
	ShortURL string `json:"short_url"`

	// OriginalURL is the original URL corresponding to the shortened URL.
	OriginalURL string `json:"original_url"`
}

// ResponseError represents an error response with a message.
type ResponseError struct {
	// Error is the error message.
	Error string `json:"error"`
}

// NewError creates a new ResponseError with the provided error message.
// This is used for returning error responses in a consistent format.
func NewError(err string) ResponseError {
	return ResponseError{
		Error: err,
	}
}
