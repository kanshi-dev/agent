package app

import (
	"context"
	"log"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/pipeline"
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
	batch := pipeline.Batch{}

	ticker := time.NewTicker(cfg.Interval)
	flushTicker := time.NewTicker(cfg.FlushEvery)
	defer flushTicker.Stop()
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			log.Printf("kanshi-agent shutting down")
			return nil

		case <-flushTicker.C:
			flushed := batch.Flush()
			if len(flushed) > 0 {
				log.Printf("flushed batch size = %d", len(flushed))
			}

		case <-ticker.C:
			for _, c := range collectors {
				points, err := c.Collect(ctx)
				if err != nil {
					log.Printf("failed to collect %s: %v", c.Name(), err)
					continue
				}

				// Add points to batch
				batch.Add(points)

				// Flush batch if it's full'
				if batch.Len() >= cfg.BatchMax {
					batch.Flush()
				}
			}
		}
	}
}
