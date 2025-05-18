// Package urls provides a handler for retrieving a user's URLs.
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

// URLStorage defines the methods for storing and retrieving URLs.
type URLStorage interface {
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)
}

// New creates a new handler function that retrieves a list of URLs for a specific user.
//
// This function processes the incoming request, retrieves the user's URLs from storage,
// and returns them in the response. If the user does not have any URLs, it returns a 204 No Content status.
// If there is an error retrieving the URLs, it returns a 500 Internal Server Error.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - storage: The URL storage service used to retrieve the user's URLs.
//
// Returns:
// - An HTTP handler function that processes the request and returns the list of URLs for the user.
func New(ctx context.Context, storage URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(string(constants.XUserID))
		slog.Info(fmt.Sprintf("userID requested (urls.go): %s", userID))

		if userID == "" {
			slog.Error("userID is empty")
			http.Error(w, "userID is empty", http.StatusBadRequest)
			return
		}

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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
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
