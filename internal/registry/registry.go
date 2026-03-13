package registry

import (
	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/collect/system"
)

// Enabled returns a slice of all metric collectors enabled for this agent.
func Enabled() []collect.Collector {
	return []collect.Collector{
		system.CPUCollector{},
		system.MemCollector{},
		system.DiskCollector{},
	}
}
