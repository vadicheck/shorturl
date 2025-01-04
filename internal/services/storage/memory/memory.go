package memory

import (
	"context"
	"os"

	"github.com/vadicheck/shorturl/internal/models"
)

type Storage struct {
	producer *Producer
	consumer *Consumer
	urls     map[string]models.URL
}

func New(fileName string) (*Storage, error) {
	pFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, permission)
	if err != nil {
		return nil, err
	}
	producer, err := NewProducer(pFile)
	if err != nil {
		return nil, err
	}

	cFile, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, permission)
	if err != nil {
		return nil, err
	}

	consumer, err := NewConsumer(cFile)
	if err != nil {
		return nil, err
	}

	urls, err := consumer.Load()
	if err != nil {
		return nil, err
	}

	return &Storage{
		producer: producer,
		consumer: consumer,
		urls:     urls,
	}, nil
}

func (s *Storage) SaveURL(ctx context.Context, code, url string) (int64, error) {
	id := int64(len(s.urls) + 1)

	mURL := models.URL{
		ID:   id,
		Code: code,
		URL:  url,
	}

	s.urls[code] = mURL

	err := s.producer.WriteURL(&mURL)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Storage) GetURLByID(ctx context.Context, code string) (models.URL, error) {
	url, ok := s.urls[code]
	if !ok {
		return models.URL{}, nil
	}

	return url, nil
}

func (s *Storage) GetURLByURL(ctx context.Context, url string) (models.URL, error) {
	for _, u := range s.urls {
		if u.URL == url {
			return u, nil
		}
	}

	return models.URL{}, nil
}
