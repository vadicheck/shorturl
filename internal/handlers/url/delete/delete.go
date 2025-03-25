// Package delete provides a handler for processing URL deletion requests.
package delete

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/constants"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	delValidator "github.com/vadicheck/shorturl/internal/validator"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

// New creates a new handler function for processing URL deletion requests.
//
// It reads a JSON body containing a list of URLs to be deleted, validates the data,
// and then asynchronously deletes the URLs. The response is returned immediately
// with a 202 Accepted status, and the deletion process continues in the background.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - service: The URL service used to delete the URLs.
// - validator: The validator used to validate the delete request data.
//
// Returns:
// - A handler function that processes HTTP requests for URL deletion.
func New(
	ctx context.Context,
	service *urlservice.Service,
	validator delValidator.DeleteURLsValidator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request []string

		slog.Info(fmt.Sprintf("userID requested (urls.go): %s", r.Header.Get(string(constants.XUserID))))

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			httpError.RespondWithError(w, http.StatusInternalServerError, "Invalid JSON body")
			return
		}

		errs := validator.DeleteShortURLs(&request)
		if len(errs.Errors) != 0 {
			httpError.RespondWithError(w, http.StatusBadRequest, errs.Error())
			return
		}

		closeCh := make(chan string)

		go func() {
			defer close(closeCh)

			if err := service.Delete(ctx, request, r.Header.Get(string(constants.XUserID))); err != nil {
				slog.Error("failed to delete URLs", sl.Err(err))
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		if err := json.NewEncoder(w).Encode(nil); err != nil {
			slog.Error("error encoding response", sl.Err(err))
			httpError.RespondWithError(w, http.StatusInternalServerError, "Failed encoding response")
			return
		}

		select {
		case <-closeCh:
		case <-ctx.Done():
			return
		}
	}
}
