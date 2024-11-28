package save

import (
	"context"
	surl "github.com/vadicheck/shorturl/internal/services/url"
	"github.com/vadicheck/shorturl/internal/services/validators/url"
	"io"
	"log"
	"net/http"
)

func New(ctx context.Context, service surl.Service) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//if req.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		//	log.Println("Unsupported Content-Type")
		//	http.Error(res, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
		//	return
		//}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Println("Error reading body: ", err)
			http.Error(res, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		log.Printf("Received request body: %s", body)

		reqURL := string(body)

		_, err = url.IsValid(reqURL)
		if err != nil {
			log.Println("URL is invalid: ", err)
			http.Error(res, "URL is invalid", http.StatusBadRequest)
			return
		}

		code, err := service.Create(ctx, reqURL)
		if err != nil {
			log.Println("Error saving the record: ", err)
			http.Error(res, "Failed to save the record", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(code))
	}
}
