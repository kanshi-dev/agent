package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/kanshi-dev/agent/internal/config"
	"github.com/kanshi-dev/agent/internal/logger"
	"github.com/kanshi-dev/agent/internal/transport"
	ingest "github.com/kanshi-dev/agent/proto"
)

const maxProfileArtifact = 10 << 20

type profileWorker struct {
	ctx     context.Context
	cfg     config.Config
	agentID string
	log     *logger.StdLogger
	jobs    chan *ingest.ProfileCommand
	mu      sync.Mutex
	current string
	last    string
}

func newProfileWorker(ctx context.Context, cfg config.Config, agentID string, logg *logger.StdLogger) *profileWorker {
	w := &profileWorker{ctx: ctx, cfg: cfg, agentID: agentID, log: logg, jobs: make(chan *ingest.ProfileCommand, 1)}
	go w.run()
	return w
}

func (w *profileWorker) Submit(command *ingest.ProfileCommand) {
	if command == nil || command.GetCaptureId() == "" {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if command.GetCaptureId() == w.current || command.GetCaptureId() == w.last {
		return
	}
	select {
	case w.jobs <- command:
		w.current = command.GetCaptureId()
	default:
	}
}

func (w *profileWorker) run() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case command := <-w.jobs:
			w.handle(command)
			w.mu.Lock()
			w.last, w.current = command.GetCaptureId(), ""
			w.mu.Unlock()
		}
	}
}

func (w *profileWorker) handle(command *ingest.ProfileCommand) {
	expires := time.Unix(0, command.GetExpiresUnixNano())
	if !expires.After(time.Now()) {
		return
	}
	artifact, filename, captureErr := captureProfile(w.ctx, w.cfg, command)
	upload := &ingest.ProfileUpload{CaptureId: command.GetCaptureId(), AgentId: w.agentID}
	if captureErr != nil {
		upload.Error = captureErr.Error()
	} else {
		upload.Artifact = artifact
		upload.Filename = filename
		upload.ContentType = "application/octet-stream"
	}

	for time.Now().Before(expires) {
		if w.ctx.Err() != nil {
			return
		}
		sender, err := transport.New(w.cfg, w.agentID)
		if err == nil {
			ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
			err = sender.UploadProfile(ctx, upload)
			cancel()
		}
		if err == nil {
			return
		}
		w.log.Error("profile upload failed: %v", err)
		timer := time.NewTimer(min(2*time.Second, time.Until(expires)))
		select {
		case <-w.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func captureProfile(ctx context.Context, cfg config.Config, command *ingest.ProfileCommand) ([]byte, string, error) {
	base, err := resolveProfileTarget(ctx, cfg, command.GetTargetName())
	if err != nil {
		return nil, "", err
	}
	path := command.GetProfileType()
	query := ""
	switch path {
	case "cpu":
		path, query = "profile", "seconds="+strconv.Itoa(int(command.GetDurationSeconds()))
	case "trace":
		query = "seconds=" + strconv.Itoa(int(command.GetDurationSeconds()))
	case "heap", "allocs", "goroutine", "mutex", "block", "threadcreate":
	default:
		return nil, "", errors.New("unsupported profile type")
	}
	u, _ := url.Parse(base)
	u.Path = "/debug/pprof/" + path
	u.RawQuery = query
	timeout := 5*time.Second + time.Duration(command.GetDurationSeconds())*time.Second
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, _ := http.NewRequestWithContext(requestCtx, http.MethodGet, u.String(), nil)
	resp, err := profileHTTPClient().Do(req)
	if err != nil {
		return nil, "", errors.New("profile request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("profile endpoint returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxProfileArtifact+1))
	if err != nil {
		return nil, "", errors.New("profile response failed")
	}
	if len(data) > maxProfileArtifact {
		return nil, "", errors.New("profile artifact exceeds 10 MiB")
	}
	ext := ".pb.gz"
	if command.GetProfileType() == "trace" {
		ext = ".trace"
	}
	return data, command.GetCaptureId() + "-" + command.GetProfileType() + ext, nil
}

func resolveProfileTarget(ctx context.Context, cfg config.Config, name string) (string, error) {
	for _, target := range cfg.PprofTargets {
		if target.Name == name {
			return target.URL, nil
		}
	}
	for _, target := range cfg.PprofDiscovery {
		if target.Name != name {
			continue
		}
		for port := target.StartPort; port <= target.EndPort; port++ {
			base := target.Scheme + "://" + net.JoinHostPort(target.Host, strconv.Itoa(port))
			probeCtx, cancel := context.WithTimeout(ctx, time.Second)
			req, _ := http.NewRequestWithContext(probeCtx, http.MethodGet, base+"/debug/pprof/", nil)
			resp, err := profileHTTPClient().Do(req)
			cancel()
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return base, nil
				}
			}
		}
		return "", errors.New("no pprof endpoint found in approved discovery range")
	}
	return "", errors.New("profile target is not configured")
}

func profileHTTPClient() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}
