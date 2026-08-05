package system

import (
	"context"
	"time"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/shirou/gopsutil/v4/net"
)

type NetCollector struct {
	previous   *net.IOCountersStat
	previousAt time.Time
	now        func() time.Time
	read       func(context.Context, bool) ([]net.IOCountersStat, error)
}

func (*NetCollector) Name() string { return "net" }

func (c *NetCollector) Collect(ctx context.Context) ([]collect.Point, error) {
	read := c.read
	if read == nil {
		read = net.IOCountersWithContext
	}
	now := time.Now
	if c.now != nil {
		now = c.now
	}

	stats, err := read(ctx, false)
	if err != nil || len(stats) == 0 {
		return nil, err
	}
	current, at := stats[0], now()
	previous, previousAt := c.previous, c.previousAt
	c.previous, c.previousAt = &current, at
	if previous == nil || current.BytesSent < previous.BytesSent || current.BytesRecv < previous.BytesRecv || !at.After(previousAt) {
		return nil, nil
	}

	seconds := at.Sub(previousAt).Seconds()
	return []collect.Point{
		{Name: "net.bytes_sent_per_second", Value: float64(current.BytesSent-previous.BytesSent) / seconds, Timestamp: at},
		{Name: "net.bytes_recv_per_second", Value: float64(current.BytesRecv-previous.BytesRecv) / seconds, Timestamp: at},
	}, nil
}
