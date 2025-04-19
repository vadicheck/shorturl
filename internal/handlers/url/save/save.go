// Package save provides a handler for saving a new URL and generating a shortened version.
package save

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/pkg/validators/url"
)

// New creates a new handler function for saving a URL and generating its shortened version.
//
// It processes the URL from the request body, validates it, and attempts to create a shortened URL.
// If the URL is already shortened, it returns a conflict status with the existing shortened URL.
// If the URL is invalid, it returns a bad request status.
// On successful creation, it returns the shortened URL with an HTTP status of 201 Created.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - service: The URL service used to create the shortened URL.
//
// Returns:
// - An HTTP handler function that processes the URL creation request and returns the result.
func New(ctx context.Context, service *urlservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error(fmt.Sprintf("Error reading body: %s", err))
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer func() {
			if errBodyClose := r.Body.Close(); errBodyClose != nil {
				slog.Error(fmt.Sprintf("failed to close body: %v", errBodyClose))
			}
		}()

		slog.Info(fmt.Sprintf("Received request body: %s", body))

		reqURL := string(body)

		_, err = url.IsValid(reqURL)
		if err != nil {
			slog.Error(fmt.Sprintf("URL is invalid: %s", err))
			http.Error(w, "URL is invalid", http.StatusBadRequest)
			return
		}

		httpStatus := http.StatusCreated
		response := shorten.CreateURLResponse{}

		slog.Info(fmt.Sprintf("userID requested (save.go): %s", r.Header.Get(string(constants.XUserID))))

		code, err := service.Create(ctx, reqURL, r.Header.Get(string(constants.XUserID)))
		if err != nil {
			var storageErr *storage.ExistsURLError

			if errors.As(err, &storageErr) {
				httpStatus = http.StatusConflict
				response.Result = config.Config.BaseURL + "/" + storageErr.ShortCode
			} else {
				httpError.RespondWithError(w, http.StatusInternalServerError, "Failed to create")
			}
		} else {
			response.Result = config.Config.BaseURL + "/" + code
		}

		slog.Info(fmt.Sprintf("Result (save.go): %s", response.Result))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(httpStatus)

		_, err = w.Write([]byte(response.Result))

		if err != nil {
			slog.Error(fmt.Sprintf("Error writing response: %s", err))
			return
		}
	}
}
