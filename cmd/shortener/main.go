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

func main() {
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
