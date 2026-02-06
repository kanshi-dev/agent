package system

import (
	"context"
	"time"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/shirou/gopsutil/v4/mem"
)

type MemCollector struct{}

func (MemCollector) Name() string {
	return "mem"
}

func (MemCollector) Collect(ctx context.Context) ([]collect.Point, error) {
	memory, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return []collect.Point{
		{
			TimeStamp: time.Now(),
			Name:      "mem.used_percent",
			Value:     memory.UsedPercent,
			Tags:      nil,
		},
	}, nil
}
