// Package memory provides functionality for reading and decoding URL data from an input reader.
// It defines a Consumer that can read and load URL data into memory.
package memory

import (
	"encoding/json"
	"io"

	"github.com/vadicheck/shorturl/internal/models"
)

const permission = 0600

// Consumer is a type that provides methods for reading and decoding URL data from an input reader.
type Consumer struct {
	// reader is the input reader from which URL data will be read.
	reader *io.Reader

	// decoder is the JSON decoder used to decode the input data.
	decoder *json.Decoder
}

// NewConsumer creates a new Consumer instance using the provided input reader.
// It returns a pointer to a Consumer and any potential error encountered.
func NewConsumer(reader io.Reader) (*Consumer, error) {
	return &Consumer{
		reader:  &reader,
		decoder: json.NewDecoder(reader),
	}, nil
}

// ReadURL reads a single URL entry from the input data and decodes it into a models.URL object.
// It returns a pointer to the URL object and any error encountered during decoding.
func (c *Consumer) ReadURL() (*models.URL, error) {
	url := &models.URL{}
	if err := c.decoder.Decode(&url); err != nil {
		return nil, err
	}

	return url, nil
}

// Load reads and decodes all URL entries from the input data into a map of URLs.
// The map uses the URL code as the key and the decoded URL as the value.
// It returns the map of URLs and any error encountered during decoding.
func (c *Consumer) Load() (map[string]models.URL, error) {
	urlMap := make(map[string]models.URL)

	for {
		url := &models.URL{}
		if err := c.decoder.Decode(&url); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		urlMap[url.Code] = *url
	}

	return urlMap, nil
}
