package app

import (
	"context"
	"math/rand"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/identity"
	"github.com/kanshi-dev/agent/internal/logger"
	"github.com/kanshi-dev/agent/internal/pipeline"
	"github.com/kanshi-dev/agent/internal/registry"
	"github.com/kanshi-dev/agent/internal/transport"
)

var version = "dev"

func Run(ctx context.Context, cfg config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	lvl := logger.ParseLevel(cfg.LogLevel)
	logg := logger.New(lvl)

	logg.Info("kanshi-agent starting: core=%s interval=%s batchMax=%d flushEvery=%s tags=%d",
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
	info, err := identity.Collect(version)
	if err != nil {
		return err
	}

	for {
		sender, err = transport.New(cfg, agentID)
		if err == nil {
			ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = sender.ReportAgent(ctxTimeout, info)
			cancel()

			if err == nil {
				break
			}

			if transport.IsAuthError(err) {
				logg.Error("authentication failed while reporting agent (check KANSHI_API_KEY): %v", err)
			} else {
				logg.Error("report failed: %v", err)
			}
		} else {
			logg.Error("connect failed: %v", err)
		}

		if !sleepWithJitter(ctx, 5*time.Second) {
			logg.Info("kanshi-agent shutting down")
			return nil
		}
	}

	ticker := time.NewTicker(cfg.Interval)
	flushTicker := time.NewTicker(cfg.FlushEvery)
	defer ticker.Stop()
	defer flushTicker.Stop()

	for {
		select {

		case <-ctx.Done():
			sendBatch(ctx, batch, &sender, cfg, agentID, logg)
			logg.Info("kanshi-agent shutting down")
			return nil

		case <-flushTicker.C:
			sendBatch(ctx, batch, &sender, cfg, agentID, logg)

		case <-ticker.C:
			for _, c := range collectors {
				points, err := c.Collect(ctx)
				if err != nil {
					logg.Error("failed to collect %s: %v", c.Name(), err)
					continue
				}

				batch.Add(points)

				if batch.Len() >= cfg.BatchMax {
					sendBatch(ctx, batch, &sender, cfg, agentID, logg)
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
	logg *logger.StdLogger,
) {
	payload := batch.Flush()

	if len(payload) == 0 {
		return
	}

	drop := func() {
		logg.Warn("dropping %d metric points during shutdown", len(payload))
	}

	for {
		select {
		case <-ctx.Done():
			drop()
			return
		default:
		}

		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := (*sender).Send(ctxTimeout, payload)
		cancel()

		if err == nil {
			return
		}

		if transport.IsAuthError(err) {
			logg.Error("authentication failed while sending batch (check KANSHI_API_KEY): %v", err)
			if !sleepWithJitter(ctx, 10*time.Second) {
				drop()
				return
			}
			continue
		}

		logg.Error("send failed: %v", err)

		for {
			select {
			case <-ctx.Done():
				drop()
				return
			default:
			}

			newSender, err := transport.New(cfg, agentID)
			if err == nil {
				*sender = newSender
				break
			}

			logg.Error("reconnect failed: %v", err)
			if !sleepWithJitter(ctx, 5*time.Second) {
				drop()
				return
			}
		}

		// retry the same payload (NO DATA LOSS)
		if !sleepWithJitter(ctx, 2*time.Second) {
			drop()
			return
		}
	}
}

func sleepWithJitter(ctx context.Context, base time.Duration) bool {
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	timer := time.NewTimer(base + jitter)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
