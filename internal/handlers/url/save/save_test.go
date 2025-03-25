package save

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
)

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

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
			name: "simple test",
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
		{
			name: "Invalid body",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
				response:    "Failed to read request body",
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
			var reqContent io.Reader

			if tt.name == "Invalid body" {
				reqContent = errorReader{}
			} else {
				reqContent = strings.NewReader(tt.request.url)
			}

			req := httptest.NewRequest(http.MethodPost, "/", reqContent)
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

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

			for code, url := range tt.urls {
				_, err = storage.SaveURL(ctx, code, url.URL, url.UserID)
				require.NoError(t, err)
			}

			req.Header.Set(string(constants.XUserID), uuid.New().String())

			New(ctx, urlservice.New(storage))(w, req)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			defer func() {
				if err := result.Body.Close(); err != nil {
					slog.Error(fmt.Sprintf("failed to close body: %v", err))
				}
			}()

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
