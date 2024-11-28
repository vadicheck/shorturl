package memory

import (
	"context"
	"github.com/vadicheck/shorturl/internal/models"
)

type Storage struct {
	urls map[string]models.Url
}

func New(urls map[string]models.Url) (*Storage, error) {
	return &Storage{urls}, nil
}

func (s *Storage) SaveUrl(ctx context.Context, code string, url string) (int64, error) {
	id := int64(len(s.urls) + 1)

	mUrl := models.Url{
		ID:   id,
		Code: code,
		Url:  url,
	}

	s.urls[code] = mUrl

	return id, nil
}

func (s *Storage) GetUrlById(ctx context.Context, code string) (models.Url, error) {
	url, ok := s.urls[code]
	if !ok {
		return models.Url{}, nil
	}

	return url, nil
}

func (s *Storage) GetUrlByUrl(ctx context.Context, url string) (models.Url, error) {
	for _, u := range s.urls {
		if u.Url == url {
			return u, nil
		}
	}

	return models.Url{}, nil
}
