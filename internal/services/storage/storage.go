package storage

import (
	"context"
	"errors"
	"github.com/vadicheck/shorturl/internal/models"
)

var (
	ErrUrlOrCodeExists = errors.New("url or code exists")
)

type UrlStorage interface {
	SaveUrl(ctx context.Context, code string, url string) (int64, error)
	GetUrlById(ctx context.Context, code string) (models.Url, error)
	GetUrlByUrl(ctx context.Context, url string) (models.Url, error)
}
