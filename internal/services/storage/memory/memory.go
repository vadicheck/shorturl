package memory

import (
	"context"

	"github.com/vadicheck/shorturl/internal/models"
)

type Storage struct {
	urls map[string]models.URL
}

func New() (*Storage, error) {
	return &Storage{make(map[string]models.URL)}, nil
}

func (s *Storage) SaveURL(ctx context.Context, code, url string) (int64, error) {
	id := int64(len(s.urls) + 1)

	mURL := models.URL{
		ID:   id,
		Code: code,
		URL:  url,
	}

	s.urls[code] = mURL

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
