package ping

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vadicheck/shorturl/internal/services/storage/memory"
)

// mockStorage с эмуляцией ошибки
type mockStorage struct{}

func (m *mockStorage) PingContext(ctx context.Context) error {
	return errors.New("database connection failed")
}

func TestNew(t *testing.T) {
	tempFile, err := os.CreateTemp("", "tempfile-*.json")
	if err != nil {
		require.NoError(t, err)
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close temp file: %v", err))
		}
	}()

	storage, err := memory.New(tempFile.Name())
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		storage    URLStorage
		statusCode int
	}{
		{
			name:       "success",
			storage:    storage,
			statusCode: http.StatusOK,
		},
		{
			name:       "failure",
			storage:    &mockStorage{},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			New(ctx, tt.storage)(w, req)

			result := w.Result()
			defer func() {
				if err := result.Body.Close(); err != nil {
					slog.Error(fmt.Sprintf("failed to close body: %v", err))
				}
			}()

			assert.Equal(t, tt.statusCode, result.StatusCode)
		})
	}
}
