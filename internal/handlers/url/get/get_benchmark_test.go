package get

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
)

const (
	userID = "da9da41c-8f65-4ed6-abea-d58f50c41562"
	file   = "urls.txt"
)

func BenchmarkNew(b *testing.B) {
	tempFile, err := os.CreateTemp("", "tempfile-*.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			log.Printf("failed to close temp file: %v", err)
		}
	}()

	storage, err := memory.New(tempFile.Name())
	if err != nil {
		panic(err)
	}

	urls := make([]string, 0, 1598)
	file, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, string(scanner.Bytes()))
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	for key, url := range urls {
		if _, err := storage.SaveURL(ctx, strconv.Itoa(key), url, userID); err != nil {
			log.Fatal(err)
		}
	}

	handler := New(ctx, storage)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set(string(constants.XUserID), userID)
		req.SetPathValue("id", "1")

		w := httptest.NewRecorder()

		handler(w, req)
	}
}
