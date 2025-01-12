package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/vadicheck/shorturl/internal/config"
	geturl "github.com/vadicheck/shorturl/internal/handlers/url/get"
	"github.com/vadicheck/shorturl/internal/handlers/url/ping"
	saveurl "github.com/vadicheck/shorturl/internal/handlers/url/save"
	"github.com/vadicheck/shorturl/internal/handlers/url/shorten"
	"github.com/vadicheck/shorturl/internal/middleware/gzip"
	middlewarelogger "github.com/vadicheck/shorturl/internal/middleware/logger"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/storage/postgres"
	"github.com/vadicheck/shorturl/internal/services/storage/sqlite"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
)

type App struct {
	router        *chi.Mux
	serverAddress string
}

func (a *App) Run() error {
	server := &http.Server{
		Addr:         config.Config.ServerAddress,
		Handler:      a.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	slog.Info(fmt.Sprintf("Server starting: %s", a.serverAddress))

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("Error starting server")
		return err
	}

	return nil
}

func New() *App {
	config.ParseFlags()

	var err error
	var storage urlservice.URLStorage

	if config.Config.DatabaseDsn != "" {
		storage, err = postgres.New(config.Config.DatabaseDsn)
		if err != nil {
			log.Panic(err)
		}
		slog.Info("Storage: postgres")
	} else if config.Config.StoragePath != "" {
		storage, err = sqlite.New(config.Config.StoragePath)
		if err != nil {
			log.Panic(err)
		}
		slog.Info("Storage: sqlite")
	} else {
		storage, err = memory.New(config.Config.FileStoragePath)
		if err != nil {
			log.Panic(err)
		}
		slog.Info("Storage: memory")
	}

	urlService := urlservice.New(storage)

	ctx := context.Background()

	r := chi.NewRouter()

	r.Use(gzip.New())
	r.Use(middlewarelogger.New())

	r.Get("/{id}", geturl.New(ctx, storage))
	r.Get("/ping", ping.New(ctx, storage))
	r.Post("/", saveurl.New(ctx, urlService))
	r.Post("/api/shorten", shorten.New(ctx, urlService))

	return &App{
		router:        r,
		serverAddress: config.Config.ServerAddress,
	}
}
