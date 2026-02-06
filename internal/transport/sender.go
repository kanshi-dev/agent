package transport

import (
	"context"

	"github.com/kanshi-dev/agent/internal/collect"
)

type Sender interface {
	Send(ctx context.Context, batch []collect.Point) error
}
