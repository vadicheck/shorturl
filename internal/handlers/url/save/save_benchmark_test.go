package save

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
)

func BenchmarkNew(b *testing.B) {
	tempFile, err := os.CreateTemp("", "tempfile-*.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		if errClose := tempFile.Close(); errClose != nil {
			log.Printf("failed to close temp file: %v", errClose)
		}
	}()

	storage, err := memory.New(tempFile.Name())
	if err != nil {
		panic(err)
	}

	handler := New(context.Background(), urlservice.New(storage))

	requestBody := []byte("https://example.com")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set(string(constants.XUserID), uuid.New().String())

		w := httptest.NewRecorder()

		handler(w, req)
	}
}
