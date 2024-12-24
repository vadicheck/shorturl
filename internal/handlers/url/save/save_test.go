package save

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
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
		},
	}

	ctx := context.Background()

	config.ParseFlags()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request.url))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			tempFile, err := os.CreateTemp("", "tempfile-*.json")
			if err != nil {
				require.NoError(t, err)
			}
			defer tempFile.Close()

			storage, err := memory.New(tempFile.Name())
			require.NoError(t, err)

			New(ctx, urlservice.New(storage))(w, req)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			defer result.Body.Close()
			resBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			if tt.want.statusCode == http.StatusCreated {
				id := strings.TrimPrefix(string(resBody), "http://localhost:8080/")

				mURL, err := storage.GetURLByID(ctx, id)
				assert.NoError(t, err)
				assert.Equal(t, tt.request.url, mURL.URL)
			}
		})
	}
}
