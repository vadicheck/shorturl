package save

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/pkg/validators/url"
)

func New(ctx context.Context, service *urlservice.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			slog.Error(fmt.Sprintf("Error reading body: %s", err))
			http.Error(res, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		slog.Info(fmt.Sprintf("Received request body: %s", body))

		reqURL := string(body)

		_, err = url.IsValid(reqURL)
		if err != nil {
			slog.Error(fmt.Sprintf("URL is invalid: %s", err))
			http.Error(res, "URL is invalid", http.StatusBadRequest)
			return
		}

		httpStatus := http.StatusCreated
		response := shorten.CreateURLResponse{}

		code, err := service.Create(ctx, reqURL)
		if err != nil {
			var storageErr *storage.ExistsURLError

			if errors.As(err, &storageErr) {
				httpStatus = http.StatusConflict
				response.Result = config.Config.BaseURL + "/" + storageErr.ShortCode
			} else {
				httpError.RespondWithError(res, http.StatusInternalServerError, "Failed to create")
			}
		} else {
			response.Result = config.Config.BaseURL + "/" + code
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(httpStatus)

		_, err = res.Write([]byte(response.Result))

		if err != nil {
			slog.Error(fmt.Sprintf("Error writing response: %s", err))
			return
		}
	}
}
