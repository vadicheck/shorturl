package main

import (
	"context"
	"flag"
	geturl "github.com/vadicheck/shorturl/internal/handlers/url/get"
	saveurl "github.com/vadicheck/shorturl/internal/handlers/url/save"
	"github.com/vadicheck/shorturl/internal/services/storage/sqlite"
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

	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	urlService := url.Service{
		Storage: storage,
	}

	ctx := context.Background()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{id}", geturl.New(ctx, storage))
	mux.HandleFunc("POST /", saveurl.New(ctx, urlService))

	log.Println("Server started")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
