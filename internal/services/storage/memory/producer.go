package memory

import (
	"encoding/json"
	"io"

	"github.com/vadicheck/shorturl/internal/models"
)

type Producer struct {
	writer  *io.Writer
	encoder *json.Encoder
}

func NewProducer(writer io.Writer) (*Producer, error) {
	return &Producer{
		writer:  &writer,
		encoder: json.NewEncoder(writer),
	}, nil
}

func (p *Producer) WriteURL(url *models.URL) error {
	return p.encoder.Encode(&url)
}
