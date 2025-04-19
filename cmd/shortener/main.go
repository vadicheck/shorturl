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
	"os"
	"os/signal"
	"syscall"

	"github.com/vadicheck/shorturl/internal/app"
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
	buildDate    = "2025-04-19"
	buildCommit  = "Short URL YP"
)

// main is the entry point of the application.
//
// Workflow:
//  1. Create a context with cancellation (`context.WithCancel`).
//  2. Initialize the HTTP application using `app.New(ctx)`.
//  3. Start the server with `httpApp.Run()` and handle potential errors.
//  4. Wait for system signals (`os.Interrupt`, `syscall.SIGTERM`).
//  5. Shut down the server when a signal is received or the context is canceled.
//
// If the application starts successfully, it logs `"app is ready"`.
// When the server shuts down, it logs `"Server Exited Properly"`.
func main() {
	printBuildInfo()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpApp := app.New(ctx)

	httpServer, err := httpApp.Run()
	if err != nil {
		log.Panic(fmt.Errorf("http server can't start %w", err))
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	slog.Info("app is ready")
	select {
	case v := <-exit:
		slog.Info(fmt.Sprintf("signal.Notify: %v\n\n", v))
	case done := <-ctx.Done():
		slog.Info(fmt.Sprintf("ctx.Done: %v", done))
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Info(err.Error())
	}

	slog.Info("Server Exited Properly")
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	slog.Info(fmt.Sprintf("Build version: %s\n", buildVersion))
	slog.Info(fmt.Sprintf("Build date: %s\n", buildDate))
	slog.Info(fmt.Sprintf("Build commit: %s\n", buildCommit))
}
