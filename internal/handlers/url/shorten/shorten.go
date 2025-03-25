// Package shorten provides a handler for creating a shortened URL from a long URL.
package shorten

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
	"github.com/vadicheck/shorturl/pkg/validators/url"
)

// New creates a new handler function for shortening a URL.
//
// This function processes the incoming request, validates the URL, and attempts to create a shortened URL.
// If the URL is invalid, it returns a bad request status with an error message.
// If the URL already exists, it returns a conflict status with the existing shortened URL.
// On success, it returns the newly created shortened URL with an HTTP status of 201 Created.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - service: The URL service used to create the shortened URL.
//
// Returns:
// - An HTTP handler function that processes the URL shortening request and returns the result.
func New(ctx context.Context, service *urlservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request shorten.CreateURLRequest

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			httpError.RespondWithError(w, http.StatusInternalServerError, "Invalid JSON body")
			return
		}

		_, err := url.IsValid(request.URL)
		if err != nil {
			httpError.RespondWithError(w, http.StatusBadRequest, "URL is invalid")
			return
		}

		httpStatus := http.StatusCreated
		response := shorten.CreateURLResponse{}

		slog.Info(fmt.Sprintf("userID requested (save.go): %s", r.Header.Get(string(constants.XUserID))))

		code, err := service.Create(ctx, request.URL, r.Header.Get(string(constants.XUserID)))
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("error encoding response", sl.Err(err))
			httpError.RespondWithError(w, http.StatusInternalServerError, "Failed encoding response")
			return
		}
	}
}
