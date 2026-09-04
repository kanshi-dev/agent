package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/identity"
	ingest "github.com/kanshi-dev/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LogSender implements the Sender interface using gRPC.
type LogSender struct {
	client  ingest.IngestServiceClient
	agentID string
	apiKey  string
	targets []*ingest.ProfileTarget
}

// New creates a new gRPC-based Sender.
func New(cfg config.Config, agentID string) (*LogSender, error) {
	transportCredentials := credentials.TransportCredentials(insecure.NewCredentials())
	if cfg.TLS {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: cfg.TLSServerName,
		}
		if cfg.TLSCAFile != "" {
			roots, err := x509.SystemCertPool()
			if err != nil {
				return nil, fmt.Errorf("load system CA pool: %w", err)
			}
			pem, err := os.ReadFile(cfg.TLSCAFile)
			if err != nil {
				return nil, fmt.Errorf("read KANSHI_TLS_CA_FILE: %w", err)
			}
			if !roots.AppendCertsFromPEM(pem) {
				return nil, fmt.Errorf("KANSHI_TLS_CA_FILE contains no valid certificates")
			}
			tlsConfig.RootCAs = roots
		}
		transportCredentials = credentials.NewTLS(tlsConfig)
	}

	conn, err := grpc.NewClient(cfg.CoreAddr, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, err
	}

	targets := make([]*ingest.ProfileTarget, 0, len(cfg.ProfileTargets()))
	for _, target := range cfg.ProfileTargets() {
		targets = append(targets, &ingest.ProfileTarget{Name: target.Name, Discovered: target.Discovered})
	}
	return &LogSender{
		client:  ingest.NewIngestServiceClient(conn),
		agentID: agentID,
		apiKey:  cfg.APIKey,
		targets: targets,
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
func (s *LogSender) Send(ctx context.Context, batch []collect.Point) (*ingest.ProfileCommand, error) {
	points := make([]*ingest.Point, 0, len(batch))

	for _, p := range batch {
		points = append(points, &ingest.Point{
			Name:              p.Name,
			Value:             p.Value,
			TimestampUnixNano: p.Timestamp.UnixNano(),
			Tags:              p.Tags,
		})
	}

	ack, err := s.client.IngestBatch(s.withAuth(ctx), &ingest.Batch{
		AgentId: s.agentID,
		Points:  points,
	})

	if err != nil {
		return nil, err
	}
	return ack.GetProfileCommand(), nil
}

// ReportAgent sends system information to the core service.
func (s *LogSender) ReportAgent(ctx context.Context, info *identity.SystemInfo) (*ingest.ProfileCommand, error) {
	ack, err := s.client.ReportAgent(s.withAuth(ctx), &ingest.AgentReport{
		AgentId:        s.agentID,
		Hostname:       info.Hostname,
		Os:             info.OS,
		Platform:       info.Platform,
		Arch:           info.Arch,
		CpuCores:       info.CpuCores,
		TotalMemory:    info.TotalMemory,
		Version:        info.Version,
		DiskSize:       info.DiskSize,
		ProfileTargets: s.targets,
	})
	if err != nil {
		return nil, err
	}
	return ack.GetProfileCommand(), nil
}

func (s *LogSender) UploadProfile(ctx context.Context, upload *ingest.ProfileUpload) error {
	_, err := s.client.UploadProfile(s.withAuth(ctx), upload)
	return err
}
