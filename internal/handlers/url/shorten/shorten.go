package shorten

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
	"github.com/vadicheck/shorturl/pkg/validators/url"
)

func New(ctx context.Context, service *urlservice.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var request shorten.CreateURLRequest

		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&request); err != nil {
			httpError.RespondWithError(res, http.StatusInternalServerError, "Invalid JSON body")
			return
		}

		_, err := url.IsValid(request.URL)
		if err != nil {
			httpError.RespondWithError(res, http.StatusBadRequest, "URL is invalid")
			return
		}

		code, err := service.Create(ctx, request.URL)
		if err != nil {
			slog.Error(fmt.Sprintf("Error saving the record: %s", err))
			httpError.RespondWithError(res, http.StatusBadRequest, "Failed to save the record")
			return
		}

		response := shorten.CreateURLResponse{
			Result: config.Config.BaseURL + "/" + code,
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(res).Encode(response); err != nil {
			slog.Error("error encoding response", sl.Err(err))
			httpError.RespondWithError(res, http.StatusInternalServerError, "Failed encoding response")
			return
		}
	}
}
