package system

import (
	"context"
	"testing"
	"time"
)

func ptr[T any](value T) *T { return &value }

func TestProcessCollectorSamplesAndRankings(t *testing.T) {
	at := time.Unix(100, 0)
	observations := []processObservation{
		{pid: 1, name: " cpu  worker ", cpuTime: ptr(1.0), rssBytes: ptr(uint64(10))},
		{pid: 2, name: "memory", cpuTime: ptr(2.0), rssBytes: ptr(uint64(30))},
		{pid: 3, name: "both", cpuTime: ptr(3.0), rssBytes: ptr(uint64(20))},
	}
	c := &ProcessCollector{
		topN: 1,
		now:  func() time.Time { return at },
		read: func(context.Context) ([]processObservation, int, error) { return observations, 4, nil },
	}

	points, err := c.Collect(context.Background())
	if err != nil || len(points) != 2 || points[0].Name != "process.count" || points[0].Value != 4 {
		t.Fatalf("first sample = %v, %v", points, err)
	}

	at = at.Add(time.Second)
	*observations[0].cpuTime += 3
	*observations[1].cpuTime += 1
	*observations[2].cpuTime += 2
	points, err = c.Collect(context.Background())
	if err != nil || len(points) != 5 {
		t.Fatalf("second sample = %v, %v; want count plus two metrics for two ranked processes", points, err)
	}
	seen := map[string]bool{}
	for _, point := range points[1:] {
		seen[point.Tags[0]+"/"+point.Name] = true
	}
	if !seen["pid=1/process.cpu_percent"] || !seen["pid=2/process.memory_rss_bytes"] {
		t.Fatalf("independent rankings missing from %v", points)
	}
}

func TestProcessCollectorBoundsUnionAndDropsMissingProcesses(t *testing.T) {
	at := time.Unix(100, 0)
	observations := make([]processObservation, 50)
	for i := range observations {
		observations[i] = processObservation{pid: int32(i), name: "process", cpuTime: ptr(float64(i)), rssBytes: ptr(uint64(50 - i))}
	}
	c := &ProcessCollector{topN: 20, now: func() time.Time { return at }, read: func(context.Context) ([]processObservation, int, error) { return observations, len(observations), nil }}
	if _, err := c.Collect(context.Background()); err != nil {
		t.Fatal(err)
	}
	at = at.Add(time.Second)
	for i := range observations {
		*observations[i].cpuTime += float64(i)
	}
	points, err := c.Collect(context.Background())
	if err != nil || len(points) != 81 {
		t.Fatalf("bounded sample has %d points, error %v; want count plus 40 processes", len(points), err)
	}

	observations = []processObservation{{pid: 99, name: "rss only", rssBytes: ptr(uint64(5))}, {pid: 100, name: "inaccessible"}}
	at = at.Add(time.Second)
	points, err = c.Collect(context.Background())
	if err != nil || len(points) != 2 || points[1].Name != "process.memory_rss_bytes" {
		t.Fatalf("missing process sample = %v, %v", points, err)
	}
}

func TestReadProcessesHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := readProcesses(ctx); err == nil {
		t.Fatal("expected canceled process read to fail")
	}
}
