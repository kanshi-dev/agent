package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kanshi-dev/agent/internal/app"
	"github.com/kanshi-dev/agent/internal/config"
)

func main() {

	cfg := config.DefaultConfig()
	if err := config.LoadFromEnv(&cfg); err != nil {
		log.Printf("invalid configuration: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		log.Printf("kanshi-agent failed: %v", err)
		os.Exit(1)
	}
}
