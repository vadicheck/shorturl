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
	SaveURL(ctx context.Context, code string, url string) (int64, error)
	SaveBatchURL(ctx context.Context, dto *[]repository.BatchURLDto) (*[]repository.BatchURL, error)
	GetURLByID(ctx context.Context, code string) (models.URL, error)
	GetURLByURL(ctx context.Context, url string) (models.URL, error)
}

const defaultCodeLength = 10

func (s *Service) Create(ctx context.Context, sourceURL string) (string, error) {
	code, err := s.generateCode(ctx)
	if err != nil {
		return "", err
	}

	_, err = s.storage.SaveURL(ctx, code, sourceURL)

	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) CreateBatch(
	ctx context.Context,
	request []shorten.CreateBatchURLRequest,
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

	batch, err := s.storage.SaveBatchURL(ctx, &dto)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (s *Service) generateCode(ctx context.Context) (string, error) {
	var code string
	isUnique := false

	for !isUnique {
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

	return code, nil
}
