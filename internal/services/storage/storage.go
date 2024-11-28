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
	SaveURL(ctx context.Context, code string, url string) (int64, error)
	GetURLByID(ctx context.Context, code string) (models.URL, error)
	GetURLByURL(ctx context.Context, url string) (models.URL, error)
}
