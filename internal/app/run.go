package app

import (
	"context"
	"log"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/identity"
	"github.com/kanshi-dev/agent/internal/pipeline"
	"github.com/kanshi-dev/agent/internal/registry"
	"github.com/kanshi-dev/agent/internal/transport"
)

// Run starts the agent's main collection and transmission loop.
// It initializes collectors, creates a transport, and manages the collection intervals.
func Run(ctx context.Context, cfg config.Config) error {
	log.Printf("kanshi-agent starting: core=%s interval=%s batchMax=%d flushEvery=%s tags=%d",
		cfg.CoreAddr, cfg.Interval, cfg.BatchMax, cfg.FlushEvery, len(cfg.HostTags),
	)

	collectors := registry.Enabled()
	batch := &pipeline.Batch{}

	//Generate agent ID
	agentID, err := identity.LoadOrCreateAgentID()
	if err != nil {
		return err
	}

	sender, err := transport.New(cfg.CoreAddr, agentID)
	if err != nil {
		return err
	}

	//Send agent info
	info, err := identity.Collect("0.1.0")
	if err != nil {
		return err
	}

	if err := sender.ReportAgent(ctx, info); err != nil {
		return err
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
