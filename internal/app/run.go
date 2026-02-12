package app

import (
	"context"
	"log"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/pipeline"
	"github.com/kanshi-dev/agent/internal/registry"
	"github.com/kanshi-dev/agent/internal/transport"
)

// Run
// This runs the agent with predefined configuration.
// /*
func Run(ctx context.Context, cfg config.Config) error {
	log.Printf("kanshi-agent starting: core=%s interval=%s batchMax=%d flushEvery=%s tags=%d",
		cfg.CoreAddr, cfg.Interval, cfg.BatchMax, cfg.FlushEvery, len(cfg.HostTags),
	)

	collectors := registry.Enabled()
	batch := &pipeline.Batch{}
	sender, err := transport.NewLogSender(cfg.CoreAddr, "agent-1")
	if err != nil {
	}

	ticker := time.NewTicker(cfg.Interval)
	flushTicker := time.NewTicker(cfg.FlushEvery)
	defer flushTicker.Stop()
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			sendBatch(ctx, batch, sender)
			log.Printf("kanshi-agent shutting down")
			return nil

		case <-flushTicker.C:
			sendBatch(ctx, batch, sender)

		case <-ticker.C:
			for _, c := range collectors {
				points, err := c.Collect(ctx)
				if err != nil {
					log.Printf("failed to collect %s: %v", c.Name(), err)
					continue
				}

				// Add points to batch
				batch.Add(points)

				if batch.Len() >= cfg.BatchMax {
					sendBatch(ctx, batch, sender)
				}
			}
		}
	}
}

func sendBatch(ctx context.Context, batch *pipeline.Batch, sender transport.Sender) {
	payload := batch.Flush()

	if len(payload) == 0 {
		return
	}

	if err := sender.Send(ctx, payload); err != nil {
		log.Printf("failed to send batch: %v", err)
		return
	}
}
