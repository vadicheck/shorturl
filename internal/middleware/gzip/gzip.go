// Package gzip provides middleware for handling gzip compression and decompression.
// It checks the `Accept-Encoding` and `Content-Encoding` headers to decide whether to compress the
// response body or decompress the request body, respectively.
package gzip

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/vadicheck/shorturl/pkg/compress"
)

// New returns a middleware function that handles gzip compression and decompression.
//
// The middleware inspects the `Accept-Encoding` header of the incoming request. If it contains "gzip",
// and the `Content-Type` of the request is compressible, the response body will be compressed with gzip.
//
// It also inspects the `Content-Encoding` header of the incoming request. If it indicates that the request
// body is compressed (i.e., contains "gzip"), it will decompress the body before passing it to the next handler.
//
// The middleware works with content types that are compressible, such as `application/json` and `text/html`.
//
// Parameters:
//   - None (this is a middleware factory function)
//
// Returns:
//   - A middleware function that can be used with `http.Handle` or other HTTP routers.
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
