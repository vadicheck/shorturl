package urls

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

type URLStorage interface {
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)
}

func New(ctx context.Context, storage URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(string(constants.XUserID))

		if userID == "" {
			slog.Error("userID is empty")
			http.Error(w, "userID is empty", http.StatusBadRequest)
			return
		}

		slog.Info(fmt.Sprintf("userID requested: %s", userID))

		mURLs, err := storage.GetUserURLs(ctx, userID)
		if err != nil {
			slog.Error(
				fmt.Sprintf("Failed to get user urls. userID: %s, err: %s", userID, err),
			)
			http.Error(w, "Failed to get urls", http.StatusInternalServerError)
			return
		}

		if len(mURLs) == 0 {
			slog.Info(fmt.Sprintf("No URLs found for userID: %s", userID))
			http.Error(w, "No URLs found", http.StatusNoContent)
			return
		}

		response := make([]shorten.UserURLResponse, 0)

		for _, url := range mURLs {
			response = append(response, shorten.UserURLResponse{
				ShortURL:    config.Config.BaseURL + "/" + url.Code,
				OriginalURL: url.URL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("error encoding response", sl.Err(err))
			httpError.RespondWithError(w, http.StatusInternalServerError, "Failed encoding response")
			return
		}
	}
}
