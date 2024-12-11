package urlservice

import (
	"context"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/pkg/random"
)

type Service struct {
	storage URLStorage
}

func New(storage URLStorage) *Service {
	return &Service{storage}
}

type URLStorage interface {
	SaveURL(ctx context.Context, code string, url string) (int64, error)
	GetURLByID(ctx context.Context, code string) (models.URL, error)
	GetURLByURL(ctx context.Context, url string) (models.URL, error)
}

func (s *Service) Create(ctx context.Context, sourceURL string) (string, error) {
	mURL, err := s.storage.GetURLByURL(ctx, sourceURL)
	if err != nil {
		return "", err
	}
	if mURL.ID > 0 {
		return mURL.Code, nil
	}

	var code string
	isUnique := false

	for !isUnique {
		code = random.GenerateRandomString(10)

		mURL, err = s.storage.GetURLByID(ctx, code)
		if err != nil {
			return "", err
		}
		if mURL.ID == 0 {
			isUnique = true
		}
	}

	_, err = s.storage.SaveURL(ctx, code, sourceURL)

	if err != nil {
		return "", err
	}

	return code, nil
}
