package memory

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/repository"
)

// TestStorage_SaveURL tests the SaveURL method of the Storage.
func TestStorage_SaveURL(t *testing.T) {
	storage, err := getStorage(t)
	assert.NoError(t, err)

	ctx := context.Background()
	code := "abc123"
	url := "http://example.com"
	userID := "user1"

	id, err := storage.SaveURL(ctx, code, url, userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	storedURL, err := storage.GetURLByID(ctx, code)
	assert.NoError(t, err)
	assert.Equal(t, url, storedURL.URL)
}

// TestStorage_SaveBatchURL tests the SaveBatchURL method of the Storage.
func TestStorage_SaveBatchURL(t *testing.T) {
	storage, err := getStorage(t)
	assert.NoError(t, err)

	ctx := context.Background()
	batch := []repository.BatchURLDto{
		{CorrelationID: "batch1", ShortCode: "xyz123", OriginalURL: "http://example1.com"},
		{CorrelationID: "batch2", ShortCode: "xyz124", OriginalURL: "http://example2.com"},
	}
	userID := "user1"

	batchURLs, err := storage.SaveBatchURL(ctx, &batch, userID)
	assert.NoError(t, err)
	assert.Len(t, *batchURLs, 2)

	storedURL1, err := storage.GetURLByID(ctx, "xyz123")
	assert.NoError(t, err)
	assert.Equal(t, "http://example1.com", storedURL1.URL)

	storedURL2, err := storage.GetURLByID(ctx, "xyz124")
	assert.NoError(t, err)
	assert.Equal(t, "http://example2.com", storedURL2.URL)
}

// TestStorage_GetURLByID tests the GetURLByID method of the Storage.
func TestStorage_GetURLByID(t *testing.T) {
	storage, err := getStorage(t)
	assert.NoError(t, err)

	ctx := context.Background()
	code := "abc123"
	url := "http://example.com"
	userID := "user1"

	_, err = storage.SaveURL(ctx, code, url, userID)
	assert.NoError(t, err)

	storedURL, err := storage.GetURLByID(ctx, code)
	assert.NoError(t, err)
	assert.Equal(t, url, storedURL.URL)

	storedURL, err = storage.GetURLByID(ctx, "nonexistentcode")
	assert.NoError(t, err)
	assert.Equal(t, models.URL{}, storedURL) // Should return an empty URL
}

// TestStorage_GetURLByURL tests the GetURLByURL method of the Storage.
func TestStorage_GetURLByURL(t *testing.T) {
	storage, err := getStorage(t)
	assert.NoError(t, err)

	ctx := context.Background()
	code := "abc123"
	url := "http://example.com"
	userID := "user1"

	_, err = storage.SaveURL(ctx, code, url, userID)
	assert.NoError(t, err)

	storedURL, err := storage.GetURLByURL(ctx, url)
	assert.NoError(t, err)
	assert.Equal(t, url, storedURL.URL) // Should return the correct URL

	nonExistentURL := "http://nonexistent.com"
	storedURL, err = storage.GetURLByURL(ctx, nonExistentURL)
	assert.NoError(t, err)
	assert.Equal(t, models.URL{}, storedURL) // Should return an empty URL
}

// TestStorage_PingContext tests the PingContext method of the Storage.
func TestStorage_PingContext(t *testing.T) {
	storage, err := getStorage(t)
	assert.NoError(t, err)

	ctx := context.Background()

	err = storage.PingContext(ctx)
	assert.NoError(t, err)
}

// TestStorage_GetUserURLs tests the GetUserURLs method of the Storage.
func TestStorage_GetUserURLs(t *testing.T) {
	storage, err := getStorage(t)
	assert.NoError(t, err)

	ctx := context.Background()
	userID := "user1"
	url1 := "http://example1.com"
	url2 := "http://example2.com"

	_, err = storage.SaveURL(ctx, "abc123", url1, userID)
	assert.NoError(t, err)
	_, err = storage.SaveURL(ctx, "abc124", url2, userID)
	assert.NoError(t, err)

	userURLs, err := storage.GetUserURLs(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, userURLs, 2)

	userURLs, err = storage.GetUserURLs(ctx, "nonexistentuser")
	assert.NoError(t, err)
	assert.Len(t, userURLs, 0)
}

// TestStorage_DeleteShortURLs tests the DeleteShortURLs method of the Storage.
func TestStorage_DeleteShortURLs(t *testing.T) {
	storage, err := getStorage(t)
	assert.NoError(t, err)

	ctx := context.Background()
	userID := "user1"
	url1 := "http://example1.com"
	url2 := "http://example2.com"

	_, err = storage.SaveURL(ctx, "abc123", url1, userID)
	assert.NoError(t, err)
	_, err = storage.SaveURL(ctx, "abc124", url2, userID)
	assert.NoError(t, err)

	err = storage.DeleteShortURLs(ctx, []string{"abc123"}, userID)
	assert.NoError(t, err)

	storedURL, err := storage.GetURLByID(ctx, "abc123")
	assert.NoError(t, err)
	assert.True(t, storedURL.IsDeleted)

	storedURL, err = storage.GetURLByID(ctx, "abc124")
	assert.NoError(t, err)
	assert.False(t, storedURL.IsDeleted)
}

func getStorage(t *testing.T) (*Storage, error) {
	tempFile, err := os.CreateTemp("", "test_storage_*.txt")
	assert.NoError(t, err)
	defer func() {
		if errRemove := os.Remove(tempFile.Name()); errRemove != nil {
			log.Printf("failed to remove file: %v", errRemove)
		}
	}()

	return New(tempFile.Name())
}
