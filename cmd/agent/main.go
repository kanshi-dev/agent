package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kanshi-dev/agent/internal/app"
	"github.com/kanshi-dev/agent/internal/config"
)

func main() {
	cfg := config.DefaultConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		os.Exit(1)
	}
}
