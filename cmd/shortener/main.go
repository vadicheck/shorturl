package main

import (
	"context"
	"fmt"
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
		panic(fmt.Errorf("http server can't start %w", err))
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	fmt.Println("app is ready")
	select {
	case v := <-exit:
		fmt.Printf("signal.Notify: %v\n\n", v)
	case done := <-ctx.Done():
		fmt.Println(fmt.Errorf("ctx.Done: %v", done))
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Server Exited Properly")
}
