package shorten

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
)

func TestNew(t *testing.T) {
	type request struct {
		URL string `json:"url"`
	}
	type response struct {
		Result string `json:"result"`
	}
	type responseError struct {
		Error string `json:"error"`
	}
	type want struct {
		contentType   string
		statusCode    int
		response      response
		responseError responseError
	}
	tests := []struct {
		name    string
		want    want
		request request
	}{
		{
			name: "simple test #1",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusCreated,
				response: response{
					Result: "",
				},
				responseError: responseError{},
			},
			request: request{
				URL: "https://practicum.yandex.ru/",
			},
		},
		{
			name: "Empty URL",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
				response:    response{},
				responseError: responseError{
					Error: "URL is invalid",
				},
			},
			request: request{
				URL: "",
			},
		},
		{
			name: "Invalid URL",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
				response:    response{},
				responseError: responseError{
					Error: "URL is invalid",
				},
			},
			request: request{
				URL: "et4bnnny4h",
			},
		},
	}

	ctx := context.Background()

	config.ParseFlags()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			if err != nil {
				fmt.Println("Ошибка кодирования в JSON:", err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonData))

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			tempFile, err := os.CreateTemp("", "tempfile-*.json")
			if err != nil {
				require.NoError(t, err)
			}
			defer tempFile.Close()

			storage, err := memory.New(tempFile.Name())
			require.NoError(t, err)

			newCtx := context.WithValue(ctx, constants.ContextUserID, uuid.New().String())

			New(ctx, urlservice.New(storage))(w, req.WithContext(newCtx))

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			if tt.want.statusCode == http.StatusBadRequest {
				var resError responseError
				dec := json.NewDecoder(result.Body)
				err = dec.Decode(&resError)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.responseError.Error, resError.Error)
				return
			}

			var res response
			dec := json.NewDecoder(result.Body)

			err = dec.Decode(&res)
			fmt.Println(err)
			assert.NoError(t, err)

			if tt.want.statusCode == http.StatusCreated {
				id := strings.TrimPrefix(res.Result, "http://localhost:8080/")

				mURL, err := storage.GetURLByID(ctx, id)
				assert.NoError(t, err)
				assert.Equal(t, tt.request.URL, mURL.URL)
			}
		})
	}
}
