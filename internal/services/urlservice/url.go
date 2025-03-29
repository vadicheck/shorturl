// Package urlservice provides services for managing URL shortening operations,
// including the creation of short URLs, batch creation, deletion, and URL generation.
package urlservice

import (
	"context"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/repository"
	"github.com/vadicheck/shorturl/pkg/random"
)

// Service provides the main URL shortening services, including creating short URLs,
// batch processing of short URLs, and deleting URLs.
type Service struct {
	storage URLStorage
}

// New creates a new instance of the Service with the provided URLStorage implementation.
func New(storage URLStorage) *Service {
	return &Service{storage}
}

// URLStorage is an interface that defines the storage operations needed by the URL service.
// It abstracts the interaction with different storage backends (e.g., memory, PostgreSQL).
type URLStorage interface {
	// PingContext checks the health of the storage.
	PingContext(ctx context.Context) error

	// SaveURL stores a new URL with the provided short code and user ID.
	SaveURL(ctx context.Context, code string, url string, userID string) (int64, error)

	// SaveBatchURL stores multiple URLs in a batch, associating them with the user ID.
	SaveBatchURL(ctx context.Context, dto *[]repository.BatchURLDto, userID string) (*[]repository.BatchURL, error)

	// GetURLByID retrieves a URL by its short code.
	GetURLByID(ctx context.Context, code string) (models.URL, error)

	// GetURLByURL retrieves a URL by its original URL.
	GetURLByURL(ctx context.Context, url string) (models.URL, error)

	// GetUserURLs retrieves all URLs associated with a specific user ID.
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)

	// DeleteShortURLs deletes multiple short URLs associated with the given user ID.
	DeleteShortURLs(ctx context.Context, urls []string, userID string) error
}

const defaultCodeLength = 10

// Create generates a new short code for the given source URL and saves it in the storage.
// Returns the short code or an error if the operation fails.
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

// CreateBatch generates multiple short codes for a batch of URLs and saves them in storage.
// Returns the batch of created short URLs or an error if the operation fails.
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

// Delete deletes multiple short URLs associated with the provided user ID.
// Returns an error if the operation fails.
func (s *Service) Delete(ctx context.Context, urls []string, userID string) error {
	return s.storage.DeleteShortURLs(ctx, urls, userID)
}

// generateCode generates a unique random code for a new short URL.
// It checks if the generated code already exists in the storage, and retries if necessary.
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
