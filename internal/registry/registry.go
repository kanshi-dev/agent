package registry

import (
	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/kanshi-dev/agent/internal/collect/system"
)

func Enabled() []collect.Collector {
	return []collect.Collector{
		system.CPUCollector{},
		system.MemCollector{},
		system.DiskCollector{},
	}
}
