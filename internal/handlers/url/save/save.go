package save

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
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

		code, err := service.Create(ctx, reqURL)
		if err != nil {
			slog.Error(fmt.Sprintf("Error saving the record: %s", err))
			http.Error(res, "Failed to save the record", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)

		_, err = res.Write([]byte(config.Config.BaseURL + "/" + code))

		if err != nil {
			slog.Error(fmt.Sprintf("Error writing response: %s", err))
			return
		}
	}
}
