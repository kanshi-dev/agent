package transport

import (
	"context"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/identity"
	ingest "github.com/kanshi-dev/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LogSender struct {
	client  ingest.IngestServiceClient
	agentID string
}

func New(coreAddr, agendID string) (*LogSender, error) {
	conn, err := grpc.NewClient(coreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return &LogSender{
		client:  ingest.NewIngestServiceClient(conn),
		agentID: agendID,
	}, nil
}

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

	_, err := s.client.IngestBatch(ctx, &ingest.Batch{
		AgentId: s.agentID,
		Points:  points,
	})

	return err
}

func (s *LogSender) ReportAgent(ctx context.Context, info *identity.SystemInfo) error {
	_, err := s.client.ReportAgent(ctx, &ingest.AgentReport{
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
