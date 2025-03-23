package get

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
)

func TestNew(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    string
	}
	tests := []struct {
		name string
		code string
		want want
		urls map[string]models.URL
	}{
		{
			name: "simple test #1",
			code: "code",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusTemporaryRedirect,
				response:    "https://practicum.yandex.ru/",
			},
			urls: map[string]models.URL{
				"code": {
					ID:     1,
					Code:   "code",
					URL:    "https://practicum.yandex.ru/",
					UserID: uuid.New().String(),
				},
			},
		},
		{
			name: "id is empty",
			code: "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				response:    "",
			},
			urls: map[string]models.URL{
				"code": {
					ID:     1,
					Code:   "code",
					URL:    "https://practicum.yandex.ru/",
					UserID: uuid.New().String(),
				},
			},
		},
		{
			name: "url not found",
			code: "nonexistent",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				response:    "",
			},
			urls: map[string]models.URL{
				"code": {
					ID:     1,
					Code:   "code",
					URL:    "https://practicum.yandex.ru/",
					UserID: uuid.New().String(),
				},
			},
		},
		{
			name: "url delete",
			code: "delete",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusGone,
				response:    "",
			},
			urls: map[string]models.URL{
				"delete": {
					ID:     1,
					Code:   "code",
					URL:    "https://practicum.yandex.ru/",
					UserID: uuid.New().String(),
				},
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.code, nil)
			w := httptest.NewRecorder()

			tempFile, err := os.CreateTemp("", "tempfile-*.json")
			if err != nil {
				require.NoError(t, err)
			}
			defer tempFile.Close()

			storage, err := memory.New(tempFile.Name())
			require.NoError(t, err)

			for code, url := range tt.urls {
				_, err = storage.SaveURL(ctx, code, url.URL, url.UserID)
				require.NoError(t, err)

				if tt.code == "delete" {
					err = storage.DeleteShortURLs(ctx, []string{code}, url.UserID)
					require.NoError(t, err)
				}
			}

			req.SetPathValue("id", tt.code)

			New(ctx, storage)(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
