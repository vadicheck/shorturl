package main

import (
	"context"
	"flag"
	"github.com/go-chi/chi/v5"
	geturl "github.com/vadicheck/shorturl/internal/handlers/url/get"
	saveurl "github.com/vadicheck/shorturl/internal/handlers/url/save"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/url"
	"log"
	"net/http"
)

func main() {
	var storagePath string

	flag.StringVar(&storagePath, "storage-path", "./storage/database.db", "path to storage")
	flag.Parse()

	if storagePath == "" {
		log.Fatal("Storage path is required")
		return
	}

	//storage, err := sqlite.New(storagePath)
	//if err != nil {
	//	panic(err)
	//}

	storage, err := memory.New(map[string]models.URL{})
	if err != nil {
		panic(err)
	}

	urlService := url.Service{
		Storage: storage,
	}

	ctx := context.Background()

	r := chi.NewRouter()

	r.Get("/{id}", geturl.New(ctx, storage))
	r.Post("/", saveurl.New(ctx, urlService))

	log.Println("Server started")

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
