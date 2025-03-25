package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/handlers/url/batch"
	deleteurl "github.com/vadicheck/shorturl/internal/handlers/url/delete"
	geturl "github.com/vadicheck/shorturl/internal/handlers/url/get"
	"github.com/vadicheck/shorturl/internal/handlers/url/ping"
	saveurl "github.com/vadicheck/shorturl/internal/handlers/url/save"
	"github.com/vadicheck/shorturl/internal/handlers/url/shorten"
	"github.com/vadicheck/shorturl/internal/handlers/url/urls"
	mwcookie "github.com/vadicheck/shorturl/internal/middleware/cookie"
	"github.com/vadicheck/shorturl/internal/middleware/gzip"
	middlewarelogger "github.com/vadicheck/shorturl/internal/middleware/logger"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/storage/postgres"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/internal/validator"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

const (
	readTimeout  = 10
	writeTimeout = 10
	idleTimeout  = 15
)

type App struct {
	router        *chi.Mux
	serverAddress string
}

func (a *App) Run() (*http.Server, error) {
	server := &http.Server{
		Addr:         config.Config.ServerAddress,
		Handler:      a.router,
		ReadTimeout:  readTimeout * time.Second,
		WriteTimeout: writeTimeout * time.Second,
		IdleTimeout:  idleTimeout * time.Second,
	}

	slog.Info(fmt.Sprintf("Server starting: %s", a.serverAddress))

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting server", sl.Err(err))
		}
	}()

	return server, nil
}

func New(ctx context.Context) *App {
	config.ParseFlags()

	var err error
	var storage urlservice.URLStorage

	if config.Config.DatabaseDsn != "" {
		storage, err = postgres.New(config.Config.DatabaseDsn)
		if err != nil {
			log.Panic(err)
		}
		slog.Info("Storage: postgres")
	} else {
		storage, err = memory.New(config.Config.FileStoragePath)
		if err != nil {
			log.Panic(err)
		}
		slog.Info("Storage: memory")
	}

	urlService := urlservice.New(storage)
	shortenValidator := validator.New()

	r := chi.NewRouter()

	r.Use(gzip.New())
	r.Use(mwcookie.New())
	r.Use(middlewarelogger.New())

	r.Get("/{id}", geturl.New(ctx, storage))
	r.Get("/ping", ping.New(ctx, storage))
	r.Get("/api/user/urls", urls.New(ctx, storage))
	r.Post("/", saveurl.New(ctx, urlService))
	r.Post("/api/shorten", shorten.New(ctx, urlService))
	r.Post("/api/shorten/batch", batch.New(ctx, urlService, shortenValidator))
	r.Delete("/api/user/urls", deleteurl.New(ctx, urlService, shortenValidator))

	if config.Config.AppEnv == "dev" {
		r.Mount("/debug", middleware.Profiler())
	}

	return &App{
		router:        r,
		serverAddress: config.Config.ServerAddress,
	}
}
