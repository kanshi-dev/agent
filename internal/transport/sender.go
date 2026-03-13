package transport

import (
	"context"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/identity"
)

// Sender defines the interface for transmitting data to the core service.
type Sender interface {
	// Send transmits a batch of collected points.
	Send(ctx context.Context, batch []collect.Point) error
	// ReportAgent sends system information about the agent host.
	ReportAgent(ctx context.Context, info *identity.SystemInfo) error
}
