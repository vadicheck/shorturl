package memory

import (
	"encoding/json"
	"io"
	"os"

	"github.com/vadicheck/shorturl/internal/models"
)

const permission = 0666

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, permission)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
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

func (c *Consumer) Close() error {
	return c.file.Close()
}
