package batch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/internal/validator"
)

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

func TestNew(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name    string
		want    want
		err     error
		request []shorten.CreateBatchURLRequest
	}{
		{
			name: "Valid Batch Request",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
			err: nil,
			request: []shorten.CreateBatchURLRequest{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
				{CorrelationID: "2", OriginalURL: "https://google.com"},
			},
		},
		{
			name: "Validation error",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
			},
			err:     nil,
			request: []shorten.CreateBatchURLRequest{},
		},
		{
			name: "Invalid JSON Body",
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
			},
			err:     nil,
			request: nil,
		},
	}

	ctx := context.Background()
	tempFile, err := os.CreateTemp("", "tempfile-*.json")
	require.NoError(t, err)
	defer func() {
		if errRemove := os.Remove(tempFile.Name()); errRemove != nil {
			log.Printf("failed to remove file: %v", errRemove)
		}
	}()

	storage, err := memory.New(tempFile.Name())
	require.NoError(t, err)

	service := urlservice.New(storage)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader

			if tt.request == nil {
				reqBody = errorReader{}
			} else {
				data, err := json.Marshal(tt.request)
				require.NoError(t, err)
				reqBody = io.NopCloser(strings.NewReader(string(data)))
			}

			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", reqBody)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(string(constants.XUserID), uuid.New().String())

			w := httptest.NewRecorder()
			handler := New(ctx, service, &validator.Validator{})
			handler(w, req)

			result := w.Result()
			defer func() {
				if err := result.Body.Close(); err != nil {
					log.Printf("failed to close body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}
