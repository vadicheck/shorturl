package get

import (
	"context"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"log"
	"net/http"
	"strings"
)

func New(ctx context.Context, storage storage.UrlStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")

		// Костыль, пока не решен вопрос с получением id из path в тестах
		if id == "" {
			id = strings.Trim(req.URL.String(), "/")
		}

		if id == "" {
			log.Printf("id is empty")
			http.Error(res, "id is empty", http.StatusBadRequest)
			return
		}

		log.Printf("id requested: %s", id)

		mUrl, err := storage.GetUrlById(ctx, id)
		if err != nil {
			log.Printf("Failed to get url by id. id: %s, err: %s", id, err)
			http.Error(res, "Failed to get url", http.StatusInternalServerError)
			return
		}

		if mUrl.ID == 0 {
			log.Printf("URL not found. id: %s", id)
			http.Error(res, "URL not found", http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Location", mUrl.Url)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
