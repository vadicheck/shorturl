package save

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/constants"

	"github.com/vadicheck/shorturl/internal/config"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/pkg/validators/url"
)

func New(ctx context.Context, service *urlservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error(fmt.Sprintf("Error reading body: %s", err))
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

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

		userID := r.Context().Value(constants.ContextUserID).(string)

		code, err := service.Create(ctx, reqURL, userID)
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

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(httpStatus)

		_, err = w.Write([]byte(response.Result))

		if err != nil {
			slog.Error(fmt.Sprintf("Error writing response: %s", err))
			return
		}
	}
}
