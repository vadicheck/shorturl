package batch

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
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
	return func(w http.ResponseWriter, r *http.Request) {
		var request []shorten.CreateBatchURLRequest

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&request); err != nil {
			httpError.RespondWithError(w, http.StatusInternalServerError, "Invalid JSON body")
			return
		}

		errs := validator.CreateBatchShortURL(&request)
		if len(errs.Errors) != 0 {
			httpError.RespondWithError(w, http.StatusBadRequest, errs.Error())
			return
		}

		userID := r.Context().Value(constants.ContextUserID).(string)

		batchURL, err := service.CreateBatch(ctx, request, userID)
		if err != nil {
			httpError.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		response := make([]shorten.CreateBatchURLResponse, 0)

		for _, url := range *batchURL {
			response = append(response, shorten.CreateBatchURLResponse{
				CorrelationID: url.CorrelationID,
				ShortURL:      config.Config.BaseURL + "/" + url.ShortCode,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("error encoding response", sl.Err(err))
			httpError.RespondWithError(w, http.StatusInternalServerError, "Failed encoding response")
			return
		}
	}
}
