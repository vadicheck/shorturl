package save

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/url"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    string
	}
	type request struct {
		url string
	}
	tests := []struct {
		name    string
		want    want
		request request
		urls    map[string]models.URL
	}{
		{
			name: "simple test #1",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				response:    "first",
			},
			request: request{
				url: "https://practicum.yandex.ru/",
			},
			urls: map[string]models.URL{},
		},
		{
			name: "Empty URL",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				response:    "URL is invalid",
			},
			request: request{
				url: "",
			},
			urls: map[string]models.URL{},
		},
		{
			name: "Invalid URL",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				response:    "URL is invalid",
			},
			request: request{
				url: "et4bnnny4h",
			},
			urls: map[string]models.URL{},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request.url))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			storage, err := memory.New(tt.urls)
			require.NoError(t, err)

			urlService := url.Service{
				Storage: storage,
			}

			New(ctx, urlService)(w, req)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			defer result.Body.Close()
			resBody, err := io.ReadAll(result.Body)

			if tt.want.statusCode == http.StatusCreated {
				mURL, err := storage.GetUrlById(ctx, string(resBody))
				assert.NoError(t, err)
				assert.Equal(t, tt.request.url, mURL.URL)
			}
		})
	}
}
