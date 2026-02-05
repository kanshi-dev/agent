package app

import (
	"context"
	"log"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/registry"
)

// Run
// This runs the agent with predefined configuration.
// /*
func Run(ctx context.Context, cfg config.Config) error {
	log.Printf("kanshi-agent starting: core=%s interval=%s batchMax=%d flushEvery=%s tags=%d",
		cfg.CoreAddr, cfg.Interval, cfg.BatchMax, cfg.FlushEvery, len(cfg.HostTags),
	)

	collectors := registry.Enabled()

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			log.Printf("kanshi-agent shutting down")
			return nil

		case <-ticker.C:
			for _, c := range collectors {
				points, err := c.Collect(ctx)
				if err != nil {
					log.Printf("failed to collect %s: %v", c.Name(), err)
					continue
				}
				for _, p := range points {
					log.Printf("collected %s: %v", p.Name, p.Value)
				}
			}
		}
	}
}
