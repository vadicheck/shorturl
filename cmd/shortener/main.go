package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/vadicheck/shorturl/internal/config"
	geturl "github.com/vadicheck/shorturl/internal/handlers/url/get"
	saveurl "github.com/vadicheck/shorturl/internal/handlers/url/save"
	"github.com/vadicheck/shorturl/internal/models"
	stor "github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/storage/sqlite"
	"github.com/vadicheck/shorturl/internal/services/url"
	"log"
	"net/http"
)

func main() {
	config.ParseFlags()

	var err error
	var storage stor.URLStorage

	if config.Config.StoragePath == "" {
		storage, err = memory.New(map[string]models.URL{})
		if err != nil {
			panic(err)
		}
	} else {
		storage, err = sqlite.New(config.Config.StoragePath)
		if err != nil {
			panic(err)
		}
	}

	urlService := url.Service{
		Storage: storage,
	}

	ctx := context.Background()

	r := chi.NewRouter()

	r.Get("/{id}", geturl.New(ctx, storage))
	r.Post("/", saveurl.New(ctx, urlService))

	log.Println("Server started:", config.Config.A)

	err = http.ListenAndServe(":"+config.Config.A, r)
	if err != nil {
		panic(err)
	}
}
