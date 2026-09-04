package transport

import (
	"context"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/identity"
	ingest "github.com/kanshi-dev/agent/proto"
)

// Sender defines the interface for transmitting data to the core service.
type Sender interface {
	// Send transmits a batch of collected points.
	Send(ctx context.Context, batch []collect.Point) (*ingest.ProfileCommand, error)
	// ReportAgent sends system information about the agent host.
	ReportAgent(ctx context.Context, info *identity.SystemInfo) (*ingest.ProfileCommand, error)
	UploadProfile(ctx context.Context, upload *ingest.ProfileUpload) error
}
