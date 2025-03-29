// Package memory provides an in-memory storage solution for managing URL data.
// It implements the Storage interface for saving, retrieving, and deleting URLs,
// with functionality for batch processing and user-specific URL management.
package memory

import (
	"encoding/json"
	"io"

	"github.com/vadicheck/shorturl/internal/models"
)

// Producer is responsible for encoding and writing URL data to a destination writer.
// It uses a JSON encoder to serialize URLs before writing them.
type Producer struct {
	// writer is the destination where the encoded URL data is written.
	writer *io.Writer

	// encoder is the JSON encoder used to serialize URL data.
	encoder *json.Encoder
}

// NewProducer creates and initializes a new Producer instance.
// It takes an io.Writer as a destination for writing the encoded URL data.
// It returns a pointer to the Producer instance and any error encountered during initialization.
func NewProducer(writer io.Writer) (*Producer, error) {
	return &Producer{
		writer:  &writer,
		encoder: json.NewEncoder(writer),
	}, nil
}

// WriteURL serializes the given URL and writes it to the destination writer.
// It returns an error if encoding or writing the URL fails.
func (p *Producer) WriteURL(url *models.URL) error {
	return p.encoder.Encode(&url)
}
