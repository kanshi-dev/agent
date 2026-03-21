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

	var sender transport.Sender
	for {
		sender, err = transport.New(cfg.CoreAddr, agentID)
		if err == nil {
			// gRPC NewClient is lazy, so we still need to check if we can reach it
			// or just proceed and let ReportAgent fail and trigger retry.
			// Let's proceed to ReportAgent.
			break
		}

		log.Printf("failed to connect to core %s: %v. Retrying in 5s...", cfg.CoreAddr, err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	//Send agent info
	info, err := identity.Collect("0.1.0")
	if err != nil {
		return err
	}

	for {
		if err := sender.ReportAgent(ctx, info); err != nil {
			log.Printf("failed to report agent: %v. Retrying in 5s...", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		}
		break
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
