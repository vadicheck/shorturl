package shorten

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
	"github.com/vadicheck/shorturl/pkg/validators/url"
)

func New(ctx context.Context, service *urlservice.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var request shorten.Request

		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&request); err != nil {
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(res).Encode(shorten.NewError("Invalid JSON body")); err != nil {
				slog.Error(fmt.Sprintf("cannot encode response JSON body: %s", err))
			}
			return
		}

		_, err := url.IsValid(request.URL)
		if err != nil {
			slog.Error(fmt.Sprintf("URL is invalid: %s", err))

			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(res).Encode(shorten.NewError("URL is invalid")); err != nil {
				slog.Error(fmt.Sprintf("URL is invalid: %s", err))
			}
			return
		}

		code, err := service.Create(ctx, request.URL)
		if err != nil {
			slog.Error(fmt.Sprintf("Error saving the record: %s", err))

			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(res).Encode(shorten.NewError("Failed to save the record")); err != nil {
				slog.Error(fmt.Sprintf("Failed to save the record: %s", err))
			}
			return
		}

		response := shorten.Response{
			Result: config.Config.BaseURL + "/" + code,
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(res)
		if err := enc.Encode(response); err != nil {
			slog.Error("error encoding response", sl.Err(err))

			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(res).Encode(shorten.NewError("Failed encoding response")); err != nil {
				slog.Error(fmt.Sprintf("Failed encoding response: %s", err))
			}
			return
		}
	}
}
