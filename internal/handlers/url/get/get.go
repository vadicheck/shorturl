package get

import (
	"context"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"log"
	"net/http"
)

func New(ctx context.Context, storage storage.UrlStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")

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
