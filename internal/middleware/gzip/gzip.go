package gzip

import (
	"net/http"
	"strings"

	"github.com/vadicheck/shorturl/pkg/compress"
)

func New(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		contentType := r.Header.Get("Content-Type")
		acceptEncoding := r.Header.Get("Accept-Encoding")

		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip && isCompressibleContentType(contentType) {
			cw := compress.NewCompressWriter(w)
			ow = cw
			defer cw.Close()
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
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	}
}

func isCompressibleContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/html")
}
