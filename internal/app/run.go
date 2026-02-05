package app

import (
	"context"
	"log"

	"github.com/kanshi-dev/agent/internal/config"
)

// Run
// This runs the agent with predefined configuration.
// /*
func Run(ctx context.Context, cfg config.Config) error {
	log.Printf("kanshi-agent starting: core=%s interval=%s batchMax=%d flushEvery=%s tags=%d",
		cfg.CoreAddr, cfg.Interval, cfg.BatchMax, cfg.FlushEvery, len(cfg.HostTags),
	)

	<-ctx.Done()

	log.Printf("kanshi-agent shutting down")
	return nil
}
