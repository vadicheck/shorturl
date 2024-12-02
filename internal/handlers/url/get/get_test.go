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
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
