package transport

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/identity"
	ingest "github.com/kanshi-dev/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type testIngestServer struct {
	ingest.UnimplementedIngestServiceServer
}

func (testIngestServer) ReportAgent(context.Context, *ingest.AgentReport) (*ingest.Ack, error) {
	return &ingest.Ack{}, nil
}

func TestConnectionsAndCancellation(t *testing.T) {
	cert, caFile := testCertificate(t)

	tests := []struct {
		name       string
		serverOpts []grpc.ServerOption
		configure  func(*config.Config)
	}{
		{name: "plaintext"},
		{
			name:       "TLS",
			serverOpts: []grpc.ServerOption{grpc.Creds(credentials.NewServerTLSFromCert(&cert))},
			configure: func(cfg *config.Config) {
				cfg.TLS = true
				cfg.TLSCAFile = caFile
				cfg.TLSServerName = "kanshi.test"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			server := grpc.NewServer(tt.serverOpts...)
			ingest.RegisterIngestServiceServer(server, testIngestServer{})
			go server.Serve(listener)
			t.Cleanup(server.Stop)

			cfg := config.DefaultConfig()
			cfg.CoreAddr = listener.Addr().String()
			if tt.configure != nil {
				tt.configure(&cfg)
			}
			sender, err := New(cfg, "agent")
			if err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if _, err := sender.ReportAgent(ctx, &identity.SystemInfo{}); err != nil {
				t.Fatalf("connect: %v", err)
			}

			cancelled, stop := context.WithCancel(context.Background())
			stop()
			if _, err := sender.ReportAgent(cancelled, &identity.SystemInfo{}); status.Code(err) != codes.Canceled {
				t.Fatalf("expected cancellation, got %v", err)
			}
		})
	}
}

func testCertificate(t *testing.T) (tls.Certificate, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "kanshi.test"},
		DNSNames:     []string{"kanshi.test"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return cert, writeTemp(t, "ca.pem", certPEM)
}

func writeTemp(t *testing.T, name string, content []byte) string {
	t.Helper()
	path := t.TempDir() + "/" + name
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
