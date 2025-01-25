package ping

import (
	"context"
	"net/http"
	"time"
)

type URLStorage interface {
	PingContext(ctx context.Context) error
}

func New(ctx context.Context, storage URLStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		res.Header().Set("Content-Type", "application/json")

		if err := storage.PingContext(ctx); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		}

		res.WriteHeader(http.StatusOK)
	}
}
