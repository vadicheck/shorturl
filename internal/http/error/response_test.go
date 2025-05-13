package error

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{
			name:       "Bad Request",
			statusCode: http.StatusBadRequest,
			message:    "Invalid request",
		},
		{
			name:       "Internal Server Error",
			statusCode: http.StatusInternalServerError,
			message:    "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			RespondWithError(w, tt.statusCode, tt.message)
			result := w.Result()
			defer func() {
				if err := result.Body.Close(); err != nil {
					log.Printf("failed to close body: %v", err)
				}
			}()

			assert.Equal(t, tt.statusCode, result.StatusCode)
			assert.Equal(t, "application/json", result.Header.Get("Content-Type"))

			var resp shorten.ResponseError
			err := json.NewDecoder(result.Body).Decode(&resp)

			assert.NoError(t, err)
			assert.Equal(t, tt.message, resp.Error)
		})
	}
}
