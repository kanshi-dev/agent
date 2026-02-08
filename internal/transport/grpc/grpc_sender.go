package grpc

import (
	"context"

	"github.com/kanshi-dev/agent/internal/collect"
	ingest "github.com/kanshi-dev/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcSender struct {
	client  ingest.IngestServiceClient
	agentID string
}

func NewGrpcSender(coreAddr, agendID string) (*GrpcSender, error) {
	conn, err := grpc.Dial(coreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return &GrpcSender{
		client:  ingest.NewIngestServiceClient(conn),
		agentID: agendID,
	}, nil
}

func (s *GrpcSender) Send(ctx context.Context, batch []collect.Point) error {
	points := make([]*ingest.Point, len(batch))
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
