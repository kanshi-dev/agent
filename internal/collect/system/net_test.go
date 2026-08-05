package system

import (
	"context"
	"testing"
	"time"

	psnet "github.com/shirou/gopsutil/v4/net"
)

func TestNetCollectorRatesAndReset(t *testing.T) {
	base := time.Unix(100, 0)
	now := base
	stats := psnet.IOCountersStat{BytesSent: 100, BytesRecv: 200}
	c := &NetCollector{
		now:  func() time.Time { return now },
		read: func(context.Context, bool) ([]psnet.IOCountersStat, error) { return []psnet.IOCountersStat{stats}, nil },
	}

	points, err := c.Collect(context.Background())
	if err != nil || len(points) != 0 {
		t.Fatalf("baseline = %v, %v; want no points", points, err)
	}

	now = base.Add(2 * time.Second)
	stats.BytesSent, stats.BytesRecv = 500, 800
	points, err = c.Collect(context.Background())
	if err != nil || len(points) != 2 {
		t.Fatalf("rates = %v, %v; want two points", points, err)
	}
	if points[0].Value != 200 || points[1].Value != 300 {
		t.Fatalf("rates = %v, %v; want 200, 300", points[0].Value, points[1].Value)
	}

	now = base.Add(3 * time.Second)
	stats.BytesSent, stats.BytesRecv = 10, 20
	points, err = c.Collect(context.Background())
	if err != nil || len(points) != 0 {
		t.Fatalf("reset = %v, %v; want no points", points, err)
	}
}
