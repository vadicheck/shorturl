package get

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Storage struct {
	urls map[string]models.URL
}

func NewStorage(urls map[string]models.URL) (*Storage, error) {
	return &Storage{urls}, nil
}

func (s *Storage) SaveURL(ctx context.Context, code string, url string) (int64, error) {
	return 0, nil
}

func (s *Storage) GetURLByID(ctx context.Context, code string) (models.URL, error) {
	url, ok := s.urls[code]
	if !ok {
		return models.URL{}, nil
	}

	return url, nil
}

func (s *Storage) GetURLByURL(ctx context.Context, url string) (models.URL, error) {
	return models.URL{}, nil
}

func TestNew(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    string
	}
	tests := []struct {
		name    string
		request string
		want    want
		urls    map[string]models.URL
	}{
		{
			name:    "simple test #1",
			request: "/code",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusTemporaryRedirect,
				response:    "https://practicum.yandex.ru/",
			},
			urls: map[string]models.URL{
				"code": {
					ID:   1,
					Code: "code",
					URL:  "https://practicum.yandex.ru/",
				},
			},
		},
		{
			name:    "id is empty",
			request: "/",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				response:    "",
			},
			urls: map[string]models.URL{
				"code": {
					ID:   1,
					Code: "code",
					URL:  "https://practicum.yandex.ru/",
				},
			},
		},
		{
			name:    "url not found",
			request: "/nonexistent",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				response:    "",
			},
			urls: map[string]models.URL{
				"code": {
					ID:   1,
					Code: "code",
					URL:  "https://practicum.yandex.ru/",
				},
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			storage, err := memory.New(tt.urls)
			require.NoError(t, err)

			New(ctx, storage)(w, req)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
