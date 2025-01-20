package batch

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	reqValidator "github.com/vadicheck/shorturl/internal/validator"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

func New(
	ctx context.Context,
	service *urlservice.Service,
	validator reqValidator.CreateBatchURLValidator,
) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var request []shorten.CreateBatchURLRequest

		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&request); err != nil {
			httpError.RespondWithError(res, http.StatusInternalServerError, "Invalid JSON body")
			return
		}

		errs := validator.CreateBatchShortURL(&request)
		if len(errs.Errors) != 0 {
			httpError.RespondWithError(res, http.StatusBadRequest, errs.Error())
			return
		}

		batchURL, err := service.CreateBatch(ctx, request)
		if err != nil {
			httpError.RespondWithError(res, http.StatusBadRequest, err.Error())
			return
		}

		response := make([]shorten.CreateBatchURLResponse, 0)

		for _, url := range *batchURL {
			response = append(response, shorten.CreateBatchURLResponse{
				CorrelationID: url.CorrelationID,
				ShortURL:      config.Config.BaseURL + "/" + url.ShortCode,
			})
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
