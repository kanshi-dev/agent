package registry

import (
	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/collect/system"
	"github.com/kanshi-dev/agent/internal/config"
)

// Enabled returns a slice of all metric collectors enabled for this agent.
func Enabled(cfg config.Config) []collect.Collector {
	collectors := []collect.Collector{
		system.CPUCollector{},
		system.MemCollector{},
		system.DiskCollector{},
		&system.NetCollector{},
	}
	if cfg.ProcessMetrics {
		collectors = append(collectors, system.NewProcessCollector(cfg.ProcessTopN))
	}
	return collectors
}
