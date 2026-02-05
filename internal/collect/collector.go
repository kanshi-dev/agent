package collect

import (
	"context"
	"time"
)

// Point represents a single data point with a name, value, timestamp, and tags.
type Point struct {
	Name      string
	Value     float64
	TimeStamp time.Time
	Tags      []string
}

// Collector is an interface for collecting metrics.
type Collector interface {
	Name() string
	Collect(ctx context.Context) ([]Point, error)
}
