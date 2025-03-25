package gzip

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/vadicheck/shorturl/pkg/compress"
)

func New() func(next http.Handler) http.Handler {
	slog.Info("gzip middleware enabled")

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ow := w

			contentType := r.Header.Get("Content-Type")
			acceptEncoding := r.Header.Get("Accept-Encoding")

			supportsGzip := strings.Contains(acceptEncoding, "gzip")

			if supportsGzip && isCompressibleContentType(contentType) {
				cw := compress.NewCompressWriter(w)
				ow = cw
				defer func() {
					if err := cw.Close(); err != nil {
						slog.Error(fmt.Sprintf("failed to close cw: %v", err))
					}
				}()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := compress.NewCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = cr
				defer func() {
					if err := cr.Close(); err != nil {
						slog.Error(fmt.Sprintf("failed to close cr: %v", err))
					}
				}()
			}

			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}

func isCompressibleContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/html")
}
