package app

import (
	"context"
	"testing"
	"time"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/identity"
	"github.com/kanshi-dev/agent/internal/logger"
	"github.com/kanshi-dev/agent/internal/pipeline"
	"github.com/kanshi-dev/agent/internal/transport"
	ingest "github.com/kanshi-dev/agent/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authFailSender struct{}

func (authFailSender) Send(context.Context, []collect.Point) (*ingest.ProfileCommand, error) {
	return nil, status.Error(codes.Unauthenticated, "no")
}

func (authFailSender) ReportAgent(context.Context, *identity.SystemInfo) (*ingest.ProfileCommand, error) {
	return nil, nil
}
func (authFailSender) UploadProfile(context.Context, *ingest.ProfileUpload) error { return nil }

func TestSendBatchStopsDuringRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	batch := &pipeline.Batch{}
	batch.Add([]collect.Point{{Name: "cpu.used_percent"}})
	var sender transport.Sender = authFailSender{}

	done := make(chan struct{})
	go func() {
		sendBatch(ctx, batch, &sender, config.DefaultConfig(), "agent", logger.New(logger.ERROR), nil)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("sendBatch did not stop after cancellation")
	}
}
