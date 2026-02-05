package system

import (
	"context"
	"time"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CPUCollector struct{}

func (CPUCollector) Name() string {
	return "cpu"
}

func (CPUCollector) Collect(ctx context.Context) ([]collect.Point, error) {
	pct, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, err
	}

	if len(pct) == 0 {
		return nil, nil
	}

	return []collect.Point{
		{
			Name:      "cpu.percennt",
			Value:     pct[0],
			TimeStamp: time.Now(),
			Tags:      nil,
		},
	}, nil
}
