package get

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/models"
)

type URLStorage interface {
	GetURLByID(ctx context.Context, code string) (models.URL, error)
}

func New(ctx context.Context, storage URLStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")

		if id == "" {
			slog.Error("id is empty")
			http.Error(res, "id is empty", http.StatusBadRequest)
			return
		}

		slog.Info(fmt.Sprintf("id requested: %s", id))

		mURL, err := storage.GetURLByID(ctx, id)
		if err != nil {
			slog.Error(
				fmt.Sprintf("Failed to get url by id. id: %s, err: %s", id, err),
			)
			http.Error(res, "Failed to get url", http.StatusInternalServerError)
			return
		}

		if mURL.ID == 0 {
			slog.Error(
				fmt.Sprintf("URL not found. id: %s", id),
			)
			http.Error(res, "URL not found", http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Location", mURL.URL)

		if mURL.IsDeleted {
			res.WriteHeader(http.StatusGone)
		} else {
			res.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}
