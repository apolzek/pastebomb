// Command api starts the pastebomb HTTP service.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/apolzek/pastebomb/internal/bootstrap"
)

func main() {
	// SIGINT and SIGTERM cancel the root context, which drains the HTTP
	// server and stops the background workers instead of killing them.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := bootstrap.Run(ctx); err != nil {
		slog.Error("startup failed", "err", err)
		os.Exit(1)
	}
}
