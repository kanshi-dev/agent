package transport

import (
	"context"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/identity"
	ingest "github.com/kanshi-dev/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LogSender implements the Sender interface using gRPC.
type LogSender struct {
	client  ingest.IngestServiceClient
	agentID string
	apiKey  string
}

// New creates a new gRPC-based Sender.
func New(coreAddr, agendID, apiKey string) (*LogSender, error) {
	conn, err := grpc.NewClient(coreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return &LogSender{
		client:  ingest.NewIngestServiceClient(conn),
		agentID: agendID,
		apiKey:  apiKey,
	}, nil
}

func (s *LogSender) withAuth(ctx context.Context) context.Context {
	if s.apiKey == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, "x-api-key", s.apiKey)
}

// IsAuthError returns true when an error represents an authentication/authorization failure.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	st := status.Code(err)
	return st == codes.Unauthenticated || st == codes.PermissionDenied
}

// Send transmits a batch of collected points to the core service.
func (s *LogSender) Send(ctx context.Context, batch []collect.Point) error {
	points := make([]*ingest.Point, 0, len(batch))

	for _, p := range batch {
		points = append(points, &ingest.Point{
			Name:              p.Name,
			Value:             p.Value,
			TimestampUnixNano: p.Timestamp.UnixNano(),
			Tags:              p.Tags,
		})
	}

	_, err := s.client.IngestBatch(s.withAuth(ctx), &ingest.Batch{
		AgentId: s.agentID,
		Points:  points,
	})

	return err
}

// ReportAgent sends system information to the core service.
func (s *LogSender) ReportAgent(ctx context.Context, info *identity.SystemInfo) error {
	_, err := s.client.ReportAgent(s.withAuth(ctx), &ingest.AgentReport{
		AgentId:     s.agentID,
		Hostname:    info.Hostname,
		Os:          info.OS,
		Platform:    info.Platform,
		Arch:        info.Arch,
		CpuCores:    info.CpuCores,
		TotalMemory: info.TotalMemory,
		Version:     info.Version,
		DiskSize:    info.DiskSize,
	})
	return err
}
