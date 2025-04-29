package urls

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
)

const (
	userOne = "da9da41c-8f65-4ed3-abea-d58f57c41562"
	userTwo = "da1da41c-8f69-4ed3-abea-d58f57c41409"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URL, error) {
	return []models.URL{}, errors.New("failed to get urls")
}

func TestNew(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    []shorten.UserURLResponse
	}
	tests := []struct {
		name   string
		userID string
		want   want
		urls   map[string]models.URL
	}{
		{
			name:   "empty user",
			userID: "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				response:    nil,
			},
		},
		{
			name:   "user urls",
			userID: userOne,
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				response: []shorten.UserURLResponse{
					{
						ShortURL:    "/practicum",
						OriginalURL: "https://practicum.yandex.ru/",
					},
				},
			},
			urls: map[string]models.URL{
				"practicum": {
					ID:     1,
					Code:   "practicum",
					URL:    "https://practicum.yandex.ru/",
					UserID: userOne,
				},
				"yandex": {
					ID:     1,
					Code:   "yandex",
					URL:    "https://ya.ru/",
					UserID: userTwo,
				},
			},
		},
		{
			name:   "empty urls",
			userID: userTwo,
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusNoContent,
				response:    nil,
			},
		},
		{
			name:   "Failed to get urls",
			userID: "error-uuid",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
				response:    nil,
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			w := httptest.NewRecorder()

			tempFile, err := os.CreateTemp("", "tempfile-*.json")
			if err != nil {
				require.NoError(t, err)
			}
			defer func() {
				if errClose := tempFile.Close(); errClose != nil {
					log.Printf("failed to close temp file: %v", errClose)
				}
			}()

			storage, err := memory.New(tempFile.Name())
			require.NoError(t, err)

			for _, url := range tt.urls {
				_, err = storage.SaveURL(ctx, url.Code, url.URL, url.UserID)
				require.NoError(t, err)
			}

			req.Header.Set(string(constants.XUserID), tt.userID)

			if tt.name == "Failed to get urls" {
				mockStorage := new(MockStorage)
				New(ctx, mockStorage)(w, req)
			} else {
				New(ctx, storage)(w, req)
			}

			result := w.Result()
			defer func() {
				if errClose := result.Body.Close(); errClose != nil {
					log.Printf("failed to close body: %v", errClose)
				}
			}()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			if tt.want.statusCode == http.StatusOK {
				response := make([]shorten.UserURLResponse, 0)
				dec := json.NewDecoder(result.Body)

				err = dec.Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, tt.want.response, response)
			}
		})
	}
}
