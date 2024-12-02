package get

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"log/slog"
	"net/http"
	"strings"
)

func New(ctx context.Context, storage storage.URLStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		// Костыль, пока не решен вопрос с получением id из path в тестах
		if id == "" {
			id = strings.Trim(req.URL.String(), "/")
		}

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
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
