package system

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kanshi-dev/agent/internal/collect"
	"github.com/shirou/gopsutil/v4/process"
)

type processObservation struct {
	pid      int32
	name     string
	cpuTime  *float64
	rssBytes *uint64
}

type processKey struct {
	pid  int32
	name string
}

type ProcessCollector struct {
	topN       int
	previous   map[processKey]float64
	previousAt time.Time
	now        func() time.Time
	read       func(context.Context) ([]processObservation, int, error)
}

func NewProcessCollector(topN int) *ProcessCollector { return &ProcessCollector{topN: topN} }

func (*ProcessCollector) Name() string { return "process" }

func (c *ProcessCollector) Collect(ctx context.Context) ([]collect.Point, error) {
	read := c.read
	if read == nil {
		read = readProcesses
	}
	now := time.Now
	if c.now != nil {
		now = c.now
	}

	observations, count, err := read(ctx)
	if err != nil {
		return nil, err
	}
	at := now()
	current := make(map[processKey]float64, len(observations))
	type sample struct {
		key processKey
		cpu *float64
		rss *uint64
	}
	samples := make([]sample, 0, len(observations))
	for _, observation := range observations {
		key := processKey{pid: observation.pid, name: normalizeProcessName(observation.name)}
		var cpuPercent *float64
		if observation.cpuTime != nil {
			current[key] = *observation.cpuTime
			if previous, ok := c.previous[key]; ok && at.After(c.previousAt) && *observation.cpuTime >= previous {
				value := (*observation.cpuTime - previous) / at.Sub(c.previousAt).Seconds() * 100
				cpuPercent = &value
			}
		}
		samples = append(samples, sample{key: key, cpu: cpuPercent, rss: observation.rssBytes})
	}
	c.previous, c.previousAt = current, at

	selected := make(map[processKey]sample, c.topN*2)
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].cpu == nil {
			return false
		}
		return samples[j].cpu == nil || *samples[i].cpu > *samples[j].cpu
	})
	for i := 0; i < len(samples) && i < c.topN && samples[i].cpu != nil; i++ {
		selected[samples[i].key] = samples[i]
	}
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].rss == nil {
			return false
		}
		return samples[j].rss == nil || *samples[i].rss > *samples[j].rss
	})
	for i := 0; i < len(samples) && i < c.topN && samples[i].rss != nil; i++ {
		selected[samples[i].key] = samples[i]
	}

	points := []collect.Point{{Name: "process.count", Value: float64(count), Timestamp: at}}
	for _, sample := range selected {
		tags := []string{"pid=" + strconv.FormatInt(int64(sample.key.pid), 10), "process=" + sample.key.name}
		if sample.cpu != nil {
			points = append(points, collect.Point{Name: "process.cpu_percent", Value: *sample.cpu, Timestamp: at, Tags: tags})
		}
		if sample.rss != nil {
			points = append(points, collect.Point{Name: "process.memory_rss_bytes", Value: float64(*sample.rss), Timestamp: at, Tags: tags})
		}
	}
	return points, nil
}

func readProcesses(ctx context.Context) ([]processObservation, int, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	observations := make([]processObservation, 0, len(processes))
	for _, proc := range processes {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		name, err := proc.NameWithContext(ctx)
		if err != nil {
			continue
		}
		observation := processObservation{pid: proc.Pid, name: name}
		if times, err := proc.TimesWithContext(ctx); err == nil {
			total := times.Total()
			observation.cpuTime = &total
		}
		if memory, err := proc.MemoryInfoWithContext(ctx); err == nil {
			observation.rssBytes = &memory.RSS
		}
		if observation.cpuTime != nil || observation.rssBytes != nil {
			observations = append(observations, observation)
		}
	}
	return observations, len(processes), nil
}

func normalizeProcessName(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return "unknown"
	}
	for len(name) > 240 {
		_, size := utf8.DecodeLastRuneInString(name)
		name = name[:len(name)-size]
	}
	return name
}
