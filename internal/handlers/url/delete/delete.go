package delete

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/constants"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	delValidator "github.com/vadicheck/shorturl/internal/validator"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

func New(
	ctx context.Context,
	service *urlservice.Service,
	validator delValidator.DeleteURLsValidator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request []string

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

		userID := r.Context().Value(constants.ContextUserID).(string)

		go func() {
			if err := service.Delete(ctx, request, userID); err != nil {
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
	}
}
