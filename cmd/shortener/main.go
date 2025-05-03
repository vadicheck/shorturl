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
	"os/signal"
	"syscall"
	"time"

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
	buildDate    = ""
	buildCommit  = "Short URL YP"
)

// main is the entry point of the application.
//
// Workflow:
//  1. Create a context signal.NotifyContext.
//  2. Initialize the HTTP application using `app.New(ctx)`.
//  3. Start the server with `httpApp.Run()` and handle potential errors.
//  4. Wait for system signals (`os.Interrupt`, `syscall.SIGTERM`).
//  5. Shut down the server when a signal is received or the context is canceled.
//
// If the application starts successfully, it logs `"app is ready"`.
// When the server shuts down, it logs `"Server Exited Properly"`.
func main() {
	printBuildInfo()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	httpApp := app.New(ctx)

	httpServer, err := httpApp.Run()
	if err != nil {
		log.Panic(fmt.Errorf("http server can't start: %w", err))
	}

	slog.Info("app is ready")

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	} else {
		slog.Info("server exited properly")
	}
}

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
