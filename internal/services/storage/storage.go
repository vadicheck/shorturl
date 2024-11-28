package storage

import (
	"context"
	"errors"
	"github.com/vadicheck/shorturl/internal/models"
)

var (
	ErrURLOrCodeExists = errors.New("url or code exists")
)

type URLStorage interface {
	SaveUrl(ctx context.Context, code string, url string) (int64, error)
	GetUrlById(ctx context.Context, code string) (models.URL, error)
	GetUrlByUrl(ctx context.Context, url string) (models.URL, error)
}
