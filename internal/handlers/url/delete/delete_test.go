package delete

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
	"sync"
	"testing"

	"github.com/gobuffalo/validate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
)

type mockValidator struct {
	errs error
}

func (m *mockValidator) DeleteShortURLs(_ *[]string) *validate.Errors {
	if m.errs != nil {
		errorsMap := make(map[string][]string)
		errorsMap["general"] = []string{m.errs.Error()}

		return &validate.Errors{
			Errors: errorsMap,
			Lock:   &sync.RWMutex{},
		}
	}
	return &validate.Errors{Errors: make(map[string][]string), Lock: &sync.RWMutex{}}
}

func TestNew(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    string
	}
	type request struct {
		urls []string
	}
	tests := []struct {
		name      string
		want      want
		request   request
		existing  map[string]string
		validator *mockValidator
	}{
		{
			name: "successful deletion",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusAccepted,
				response:    "null",
			},
			request: request{
				urls: []string{"first"},
			},
			existing: map[string]string{
				"first": "https://example.com",
			},
			validator: &mockValidator{},
		},
		{
			name: "invalid request body",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusInternalServerError,
				response:    `{"error":"Invalid JSON body"}`,
			},
			request: request{
				urls: nil,
			},
			existing:  nil,
			validator: &mockValidator{},
		},
		{
			name: "validation failed",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
				response:    `{"error":"validation error"}`,
			},
			request: request{
				urls: []string{"invalid"},
			},
			existing:  nil,
			validator: &mockValidator{errs: errors.New("validation error")},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqContent io.Reader

			if tt.request.urls != nil {
				body, _ := json.Marshal(tt.request.urls)
				reqContent = strings.NewReader(string(body))
			} else {
				reqContent = strings.NewReader("invalid")
			}

			req := httptest.NewRequest(http.MethodDelete, "/", reqContent)
			req.Header.Set("Content-Type", "application/json")

			tempFile, err := os.CreateTemp("", "tempfile-*.json")
			require.NoError(t, err)
			defer func() {
				if err := os.Remove(tempFile.Name()); err != nil {
					log.Printf("failed to remove file: %v", err)
				}
			}()

			storage, err := memory.New(tempFile.Name())
			require.NoError(t, err)

			for code, url := range tt.existing {
				_, err = storage.SaveURL(ctx, code, url, uuid.New().String())
				require.NoError(t, err)
			}

			req.Header.Set(string(constants.XUserID), uuid.New().String())

			w := httptest.NewRecorder()
			New(ctx, urlservice.New(storage), tt.validator)(w, req)

			result := w.Result()
			defer func() {
				if err := result.Body.Close(); err != nil {
					log.Printf("failed to close body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			resBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.JSONEq(t, tt.want.response, string(resBody))
		})
	}
}
