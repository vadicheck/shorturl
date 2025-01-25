package memory

import (
	"context"
	"fmt"
	"os"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/repository"
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

func (s *Storage) PingContext(ctx context.Context) error {
	return nil
}

func (s *Storage) SaveURL(ctx context.Context, code, url, userID string) (int64, error) {
	id := int64(len(s.urls) + 1)

	mURL := models.URL{
		ID:     id,
		Code:   code,
		URL:    url,
		UserID: userID,
	}

	s.urls[code] = mURL

	err := s.producer.WriteURL(&mURL)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Storage) SaveBatchURL(ctx context.Context, dto *[]repository.BatchURLDto, userID string) (*[]repository.BatchURL, error) {
	entities := make([]repository.BatchURL, 0)

	for _, urlDTO := range *dto {
		_, err := s.SaveURL(ctx, urlDTO.ShortCode, urlDTO.OriginalURL, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to save URL: %w", err)
		}

		entities = append(entities, repository.BatchURL{
			CorrelationID: urlDTO.CorrelationID,
			ShortCode:     urlDTO.ShortCode,
		})
	}

	return &entities, nil
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

func (s *Storage) GetUserURLs(ctx context.Context, userID string) (*[]models.URL, error) {
	var urls []models.URL

	for _, u := range s.urls {
		if u.UserID == userID {
			urls = append(urls, u)
		}
	}

	return &urls, nil
}
