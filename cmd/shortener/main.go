// Main package of the application.
//
// The `main` package starts the HTTP server for the URL shortening service.
// It manages the application's lifecycle, handles system signals,
// and ensures the server shuts down gracefully.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/vadicheck/shorturl/internal/app"
	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/grpcserver"
	"github.com/vadicheck/shorturl/internal/grpcserver/handlers"
	pb "github.com/vadicheck/shorturl/internal/proto/v1"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/storage/postgres"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
	"github.com/vadicheck/shorturl/internal/validator"
)

// These variables can be set at build time using -ldflags
//
//	go build -ldflags "\
//	 -X 'main.buildVersion=1.0.3' \
//	 -X 'main.buildDate=$(date +%Y-%m-%d)' \
//	 -X 'main.buildCommit=$(git rev-parse --short HEAD)'" \
//	 -o bin/shortener ./cmd/shortener
var (
	buildVersion = "1.0.0"
	buildDate    = ""
	buildCommit  = "Short URL YP"
)

// shutdownSeconds defines the maximum number of seconds
// allowed for graceful shutdown of the application.
const shutdownSeconds = 5

// main is the application entry point.
//
// It performs the following steps:
//  1. Parses configuration flags.
//  2. Initializes the URL storage (PostgreSQL or in-memory).
//  3. Creates the main HTTP application.
//  4. Starts the HTTP server.
//  5. Starts the gRPC server with authorization interceptor.
//  6. Waits for OS shutdown signals.
//  7. Gracefully shuts down both servers.
func main() {
	printBuildInfo()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	config.ParseFlags()

	var err error
	var storage urlservice.URLStorage

	if config.Config.DatabaseDsn != "" {
		if storage, err = postgres.New(config.Config.DatabaseDsn); err != nil {
			log.Panic(err)
		}
		slog.Info("Storage: postgres")
	} else {
		if storage, err = memory.New(config.Config.FileStoragePath); err != nil {
			log.Panic(err)
		}
		slog.Info("Storage: memory")
	}

	shortenValidator := validator.New()
	urlService := urlservice.New(storage)

	// HTTP
	httpApp := app.New(
		ctx,
		storage,
		shortenValidator,
		urlService,
	)

	httpServer, err := httpApp.Run()
	if err != nil {
		log.Panic(fmt.Errorf("http server can't start: %w", err))
	}

	// gRPC
	listen, err := net.Listen("tcp", config.Config.GRPCAddress)
	if err != nil {
		log.Fatal(err)
	}

	server := grpcserver.NewGRPCServer(handlers.NewServer(
		storage,
		shortenValidator,
		urlService,
	))

	// Добавляем Unary интерцепторы
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		grpcserver.AuthUnaryInterceptor(map[string]bool{
			"/internal.proto.v1.ShortURL/Batch":   true,
			"/internal.proto.v1.ShortURL/Delete":  true,
			"/internal.proto.v1.ShortURL/Shorten": true,
			"/internal.proto.v1.ShortURL/GetURLs": true,
		}),
		grpcserver.LoggingInterceptor(),
		grpcserver.GzipUnaryInterceptor(),
	))
	pb.RegisterShortURLServer(grpcServer, server)
	reflection.Register(grpcServer)

	go func() {
		if grpcErr := grpcServer.Serve(listen); err != nil {
			log.Panic(fmt.Errorf("grpc server can't start: %w", grpcErr))
		}
	}()

	slog.Info("app is ready")

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownSeconds*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if errShutdown := httpServer.Shutdown(shutdownCtx); errShutdown != nil {
		slog.Error("server shutdown error", "error", errShutdown)
	} else {
		slog.Info("server exited properly")
	}
}

// printBuildInfo logs the build version, date, and commit hash.
// These values can be set at build time via ldflags.
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "1.0.0"
	}
	if buildDate == "" {
		buildDate = time.Now().Format("2006-01-02")
	}
	if buildCommit == "" {
		buildCommit = "Short URL YP"
	}

	slog.Info(fmt.Sprintf("Build version: %s", buildVersion))
	slog.Info(fmt.Sprintf("Build date: %s", buildDate))
	slog.Info(fmt.Sprintf("Build commit: %s", buildCommit))
}
