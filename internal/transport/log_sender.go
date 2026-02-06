package transport

import (
	"context"
	"log"

	"github.com/kanshi-dev/agent/internal/collect"
)

type LogSender struct{}

func (LogSender) Send(ctx context.Context, batch []collect.Point) error {
	log.Printf("sending batch: %v", len(batch))
	return nil
}
