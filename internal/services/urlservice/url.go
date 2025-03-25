package urlservice

import (
	"context"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/repository"
	"github.com/vadicheck/shorturl/pkg/random"
)

type Service struct {
	storage URLStorage
}

func New(storage URLStorage) *Service {
	return &Service{storage}
}

type URLStorage interface {
	PingContext(ctx context.Context) error
	SaveURL(ctx context.Context, code string, url string, userID string) (int64, error)
	SaveBatchURL(ctx context.Context, dto *[]repository.BatchURLDto, userID string) (*[]repository.BatchURL, error)
	GetURLByID(ctx context.Context, code string) (models.URL, error)
	GetURLByURL(ctx context.Context, url string) (models.URL, error)
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)
	DeleteShortURLs(ctx context.Context, urls []string, userID string) error
}

const defaultCodeLength = 10

func (s *Service) Create(ctx context.Context, sourceURL, userID string) (string, error) {
	code, err := s.generateCode(ctx)
	if err != nil {
		return "", err
	}

	_, err = s.storage.SaveURL(ctx, code, sourceURL, userID)

	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) CreateBatch(
	ctx context.Context,
	request []shorten.CreateBatchURLRequest,
	userID string,
) (*[]repository.BatchURL, error) {
	dto := make([]repository.BatchURLDto, 0)

	for _, r := range request {
		code, err := s.generateCode(ctx)
		if err != nil {
			return nil, err
		}

		dto = append(dto, repository.BatchURLDto{
			CorrelationID: r.CorrelationID,
			OriginalURL:   r.OriginalURL,
			ShortCode:     code,
		})
	}

	batch, err := s.storage.SaveBatchURL(ctx, &dto, userID)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (s *Service) Delete(ctx context.Context, urls []string, userID string) error {
	return s.storage.DeleteShortURLs(ctx, urls, userID)
}

func (s *Service) generateCode(ctx context.Context) (string, error) {
	for {
		code, err := random.GenerateRandomString(defaultCodeLength)
		if err != nil {
			return "", err
		}

		mURL, err := s.storage.GetURLByID(ctx, code)
		if err != nil {
			return "", err
		}

		if mURL.ID == 0 {
			return code, nil
		}
	}
}
