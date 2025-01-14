package error

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vadicheck/shorturl/internal/models/shorten"
)

func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(shorten.NewError(message)); encodeErr != nil {
		slog.Error(fmt.Sprintf("cannot encode response JSON body: %s", encodeErr))
	}
}
