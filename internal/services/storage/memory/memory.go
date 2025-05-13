// Package memory provides an in-memory storage solution for managing URL data.
// It implements the Storage interface for saving, retrieving, and deleting URLs,
// with functionality for batch processing and user-specific URL management.
package memory

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/repository"
	"github.com/vadicheck/shorturl/internal/services/storage"
)

// Storage is an in-memory implementation of a URL storage system.
// It supports saving, retrieving, and deleting individual and batch URLs.
// The storage is backed by file-based producers and consumers.
type Storage struct {
	// producer is responsible for writing URL data to the storage.
	producer *Producer

	// consumer is responsible for reading and loading URL data from storage.
	consumer *Consumer

	// urls is a map of stored URLs, keyed by their unique code.
	urls map[string]models.URL
}

// New creates and initializes a new in-memory URL storage instance.
// It opens files for both reading and writing URL data and creates the producer and consumer.
// It returns a pointer to the Storage instance and any error encountered during initialization.
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

// PingContext is a no-op method for compatibility with the storage interface.
// It returns nil, indicating that the storage is available.
func (s *Storage) PingContext(ctx context.Context) error {
	return nil
}

// SaveURL saves a new URL to the storage system.
// It checks for duplicates and returns an error if the URL already exists.
// It returns the ID of the newly saved URL and any error encountered during the process.
func (s *Storage) SaveURL(ctx context.Context, code, url, userID string) (int64, error) {
	id := int64(len(s.urls) + 1)

	// Check if the URL already exists
	for _, u := range s.urls {
		if u.URL == url {
			return 0, &storage.ExistsURLError{
				OriginalURL: url,
				ShortCode:   code,
				Err:         nil,
			}
		}
	}

	mURL := models.URL{
		ID:     id,
		Code:   code,
		URL:    url,
		UserID: userID,
	}

	// Add the new URL to the map
	s.urls[code] = mURL

	// Write the URL data using the producer
	err := s.producer.WriteURL(&mURL)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveBatchURL saves a batch of URLs to the storage system.
// It accepts a slice of BatchURLDto objects and saves each URL individually.
// It returns a slice of BatchURL objects containing the correlation ID and short code of each URL.
func (s *Storage) SaveBatchURL(
	ctx context.Context,
	dto *[]repository.BatchURLDto,
	userID string,
) (*[]repository.BatchURL, error) {
	entities := make([]repository.BatchURL, 0)

	// Save each URL in the batch
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

// GetURLByID retrieves a URL from the storage by its short code.
// It returns the URL if found, or an empty URL struct if not.
func (s *Storage) GetURLByID(ctx context.Context, code string) (models.URL, error) {
	url, ok := s.urls[code]
	if !ok {
		return models.URL{}, nil
	}

	return url, nil
}

// GetURLByURL retrieves a URL from the storage by its original URL.
// It returns the URL if found, or an empty URL struct if not.
func (s *Storage) GetURLByURL(ctx context.Context, url string) (models.URL, error) {
	for _, u := range s.urls {
		if u.URL == url {
			return u, nil
		}
	}

	return models.URL{}, nil
}

// GetUserURLs retrieves all URLs associated with a specific user by their userID.
// It returns a slice of URLs associated with the user.
func (s *Storage) GetUserURLs(ctx context.Context, userID string) ([]models.URL, error) {
	var urls []models.URL

	// Collect all URLs associated with the user
	for _, u := range s.urls {
		if u.UserID == userID {
			urls = append(urls, u)
		}
	}

	return urls, nil
}

// DeleteShortURLs deletes a batch of short URLs by their short codes and the userID.
// It marks the URLs as deleted by setting their IsDeleted flag to true.
func (s *Storage) DeleteShortURLs(ctx context.Context, urls []string, userID string) error {
	// Mark the specified URLs as deleted
	for code, url := range s.urls {
		if url.UserID == userID && slices.Contains(urls, url.Code) {
			url.IsDeleted = true
			s.urls[code] = url
		}
	}
	return nil
}

// GetCountURLs returns the total number of URLs stored in memory.
// It returns the count as an integer.
func (s *Storage) GetCountURLs(ctx context.Context) (int, error) {
	return len(s.urls), nil
}

// GetCountUsers returns the number of unique users who have stored URLs.
// It returns the count of distinct user IDs.
func (s *Storage) GetCountUsers(ctx context.Context) (int, error) {
	uniqueUserIDs := make(map[string]struct{})

	for _, url := range s.urls {
		uniqueUserIDs[url.UserID] = struct{}{}
	}

	return len(uniqueUserIDs), nil
}
