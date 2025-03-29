// Package repository defines the data structures related to batch URL processing.
package repository

// BatchURL represents a shortened URL entry with a correlation ID for batch processing.
// It contains the correlation ID and the shortened code.
type BatchURL struct {
	// CorrelationID is a unique identifier for a batch operation.
	CorrelationID string `json:"correlation_id"`

	// ShortCode is the shortened code associated with the URL in the batch process.
	ShortCode string `json:"short_code"`
}

// BatchURLDto is a data transfer object (DTO) used for representing the details of a batch URL.
// It includes the correlation ID, the original URL, and the shortened code.
type BatchURLDto struct {
	// CorrelationID is a unique identifier for a batch operation.
	CorrelationID string `json:"correlation_id"`

	// OriginalURL is the original URL that was provided for shortening.
	OriginalURL string `json:"original_url"`

	// ShortCode is the shortened code generated for the original URL.
	ShortCode string `json:"short_code"`
}
