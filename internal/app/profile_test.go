package app

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/logger"
	ingest "github.com/kanshi-dev/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCaptureProfilePathsAndBounds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("seconds") == "30" && r.URL.Path != "/debug/pprof/profile" {
			t.Errorf("CPU path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer server.Close()
	cfg := config.DefaultConfig()
	cfg.PprofTargets = []config.PprofTarget{{Name: "target", URL: server.URL}}

	for _, typ := range []string{"cpu", "trace", "heap", "allocs", "goroutine", "mutex", "block", "threadcreate"} {
		duration := int32(0)
		if typ == "cpu" {
			duration = 30
		}
		if typ == "trace" {
			duration = 5
		}
		data, filename, err := captureProfile(context.Background(), cfg, &ingest.ProfileCommand{CaptureId: "id", TargetName: "target", ProfileType: typ, DurationSeconds: duration})
		if err != nil || len(data) == 0 || !strings.HasPrefix(filename, "id-"+typ) {
			t.Fatalf("%s: data=%q filename=%q err=%v", typ, data, filename, err)
		}
	}

	large := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = ioCopyN(w, maxProfileArtifact+1) }))
	defer large.Close()
	cfg.PprofTargets[0].URL = large.URL
	if _, _, err := captureProfile(context.Background(), cfg, &ingest.ProfileCommand{TargetName: "target", ProfileType: "heap"}); err == nil {
		t.Fatal("expected oversized profile to fail")
	}
}

func ioCopyN(w http.ResponseWriter, n int) (int, error) {
	written := 0
	chunk := strings.Repeat("x", 32*1024)
	for written < n {
		part := min(len(chunk), n-written)
		count, err := w.Write([]byte(chunk[:part]))
		written += count
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func TestResolveDiscoveredProfileTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/debug/pprof/" {
			t.Errorf("probe path = %s", r.URL.Path)
		}
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	port, _ := strconv.Atoi(u.Port())
	cfg := config.DefaultConfig()
	cfg.PprofDiscovery = []config.PprofDiscovery{{Name: "found", Scheme: "http", Host: u.Hostname(), StartPort: port, EndPort: port}}
	got, err := resolveProfileTarget(context.Background(), cfg, "found")
	if err != nil || got != server.URL {
		t.Fatalf("resolved %q, %v", got, err)
	}
}

func TestCaptureProfileCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	cfg := config.DefaultConfig()
	cfg.PprofTargets = []config.PprofTarget{{Name: "target", URL: server.URL}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := captureProfile(ctx, cfg, &ingest.ProfileCommand{TargetName: "target", ProfileType: "heap"}); err == nil {
		t.Fatal("expected canceled capture to fail")
	}
}

func TestProfileWorkerRetriesUploadWithoutRecapture(t *testing.T) {
	var captures atomic.Int32
	profileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		captures.Add(1)
		_, _ = w.Write([]byte("profile"))
	}))
	defer profileServer.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	uploads := &retryUploadServer{}
	ingest.RegisterIngestServiceServer(grpcServer, uploads)
	go grpcServer.Serve(listener)
	defer grpcServer.Stop()

	cfg := config.DefaultConfig()
	cfg.CoreAddr = listener.Addr().String()
	cfg.PprofTargets = []config.PprofTarget{{Name: "target", URL: profileServer.URL}}
	w := &profileWorker{ctx: context.Background(), cfg: cfg, agentID: "agent", log: logger.New(logger.ERROR)}
	w.handle(&ingest.ProfileCommand{CaptureId: "capture", TargetName: "target", ProfileType: "heap", ExpiresUnixNano: time.Now().Add(5 * time.Second).UnixNano()})
	if captures.Load() != 1 || uploads.calls.Load() != 2 {
		t.Fatalf("captures=%d uploads=%d", captures.Load(), uploads.calls.Load())
	}
}

type retryUploadServer struct {
	ingest.UnimplementedIngestServiceServer
	calls atomic.Int32
}

func (s *retryUploadServer) UploadProfile(context.Context, *ingest.ProfileUpload) (*ingest.Ack, error) {
	if s.calls.Add(1) == 1 {
		return nil, status.Error(codes.Unavailable, "retry")
	}
	return &ingest.Ack{Accepted: 1}, nil
}

func TestProfileWorkerIgnoresDuplicateCommands(t *testing.T) {
	w := &profileWorker{jobs: make(chan *ingest.ProfileCommand, 1)}
	command := &ingest.ProfileCommand{CaptureId: "same"}
	w.Submit(command)
	w.Submit(command)
	if len(w.jobs) != 1 {
		t.Fatalf("queued jobs = %d", len(w.jobs))
	}
	w.mu.Lock()
	w.current, w.last = "", "same"
	w.mu.Unlock()
	w.Submit(command)
	if len(w.jobs) != 1 {
		t.Fatalf("completed duplicate queued: %d", len(w.jobs))
	}
}
