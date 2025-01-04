package memory

import (
	"encoding/json"
	"io"

	"github.com/vadicheck/shorturl/internal/models"
)

const permission = 0666

type Consumer struct {
	reader  *io.Reader
	decoder *json.Decoder
}

func NewConsumer(reader io.Reader) (*Consumer, error) {
	return &Consumer{
		reader:  &reader,
		decoder: json.NewDecoder(reader),
	}, nil
}

func (c *Consumer) ReadURL() (*models.URL, error) {
	url := &models.URL{}
	if err := c.decoder.Decode(&url); err != nil {
		return nil, err
	}

	return url, nil
}

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
