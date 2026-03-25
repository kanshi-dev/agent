package app

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/identity"
	"github.com/kanshi-dev/agent/internal/pipeline"
	"github.com/kanshi-dev/agent/internal/registry"
	"github.com/kanshi-dev/agent/internal/transport"
)

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

	// --- CONNECT + REPORT (combined retry) ---
	var sender transport.Sender
	info, err := identity.Collect("0.1.0")
	if err != nil {
		return err
	}

	for {
		sender, err = transport.New(cfg.CoreAddr, agentID)
		if err == nil {
			ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = sender.ReportAgent(ctxTimeout, info)
			cancel()

			if err == nil {
				break
			}

			log.Printf("report failed: %v", err)
		} else {
			log.Printf("connect failed: %v", err)
		}

		sleepWithJitter(5 * time.Second)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	ticker := time.NewTicker(cfg.Interval)
	flushTicker := time.NewTicker(cfg.FlushEvery)
	defer ticker.Stop()
	defer flushTicker.Stop()

	for {
		select {

		case <-ctx.Done():
			sendBatch(ctx, batch, &sender, cfg, agentID)
			log.Printf("kanshi-agent shutting down")
			return nil

		case <-flushTicker.C:
			sendBatch(ctx, batch, &sender, cfg, agentID)

		case <-ticker.C:
			for _, c := range collectors {
				points, err := c.Collect(ctx)
				if err != nil {
					log.Printf("failed to collect %s: %v", c.Name(), err)
					continue
				}

				batch.Add(points)

				if batch.Len() >= cfg.BatchMax {
					sendBatch(ctx, batch, &sender, cfg, agentID)
				}
			}
		}
	}
}

func sendBatch(
	ctx context.Context,
	batch *pipeline.Batch,
	sender *transport.Sender,
	cfg config.Config,
	agentID string,
) {
	payload := batch.Flush()

	if len(payload) == 0 {
		return
	}

	for {
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := (*sender).Send(ctxTimeout, payload)
		cancel()

		if err == nil {
			return
		}

		log.Printf("send failed: %v", err)

		for {
			newSender, err := transport.New(cfg.CoreAddr, agentID)
			if err == nil {
				*sender = newSender
				break
			}

			log.Printf("reconnect failed: %v", err)
			sleepWithJitter(5 * time.Second)
		}

		// retry same payload (NO DATA LOSS)
		sleepWithJitter(2 * time.Second)
	}
}

func sleepWithJitter(base time.Duration) {
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(base + jitter)
}
