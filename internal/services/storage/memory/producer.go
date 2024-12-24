package memory

import (
	"encoding/json"
	"os"

	"github.com/vadicheck/shorturl/internal/models"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, permission)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteURL(url *models.URL) error {
	return p.encoder.Encode(&url)
}

func (p *Producer) Close() error {
	return p.file.Close()
}
