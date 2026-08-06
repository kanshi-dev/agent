package registry

import (
	"testing"

	"github.com/kanshi-dev/agent/internal/config"
)

func TestProcessCollectorIsOptIn(t *testing.T) {
	cfg := config.DefaultConfig()
	if got := len(Enabled(cfg)); got != 4 {
		t.Fatalf("disabled collectors = %d; want 4", got)
	}
	cfg.ProcessMetrics = true
	if got := len(Enabled(cfg)); got != 5 {
		t.Fatalf("enabled collectors = %d; want 5", got)
	}
}
