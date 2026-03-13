package system

import (
	"context"
	"time"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/shirou/gopsutil/v4/disk"
)

// DiskCollector gathers disk usage metrics.
type DiskCollector struct{}

func (DiskCollector) Name() string {
	return "disk"
}

func (DiskCollector) Collect(ctx context.Context) ([]collect.Point, error) {
	diskStats, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return nil, err
	}

	return []collect.Point{
		{
			Timestamp: time.Now(),
			Name:      "disk.used_percent",
			Value:     diskStats.UsedPercent,
			Tags:      nil,
		},
	}, nil

}
