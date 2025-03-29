// Package get provides a handler for retrieving a URL by its ID.
package get

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/models"
)

// URLStorage defines the interface for accessing URL data in the storage system.
type URLStorage interface {
	GetURLByID(ctx context.Context, code string) (models.URL, error)
}

// New creates a new handler function for retrieving a URL by its ID.
//
// It extracts the URL ID from the request path, retrieves the corresponding URL from
// the storage, and returns a redirect response based on the URL's status.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - storage: The URL storage service used to retrieve the URL by ID.
//
// Returns:
// - An HTTP handler function that processes requests for retrieving a URL by its ID.
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
