package pipeline

import "github.com/kanshi-dev/agent/internal/collect"

// Batch provides an in-memory buffer for collected metric points with a maximum size.
// When the buffer is full, the oldest points are dropped to make room for newer ones.
type Batch struct {
	points []collect.Point
	max    int
	// dropped is the number of points that have been dropped due to size limits.
	dropped int64
}

// NewBatch creates a new Batch with the specified maximum size.
// If max is less than 0, it is treated as 0.
func NewBatch(max int) *Batch {
	if max < 0 {
		max = 0
	}
	return &Batch{
		points: make([]collect.Point, 0, max),
		max:    max,
	}
}

// Add adds points to the batch, ensuring the batch does not exceed its maximum size.
// If adding the points would exceed the maximum, the oldest points are dropped first
// to keep the newest points.
// It returns the number of points dropped.
func (b *Batch) Add(points []collect.Point) int {
	if b.max == 0 {
		b.dropped += int64(len(points))
		return len(points)
	}
	if len(points) == 0 {
		return 0
	}
	combined := append(b.points, points...)
	if len(combined) > b.max {
		dropped := len(combined) - b.max
		b.dropped += int64(dropped)
		b.points = combined[len(combined)-b.max:]
		return dropped
	}
	b.points = combined
	return 0
}

// Len returns the number of points in the batch.
func (b *Batch) Len() int {
	return len(b.points)
}

// Flush returns and clears all points currently in the batch.
func (b *Batch) Flush() []collect.Point {
	points := b.points
	b.points = nil
	return points
}

// Dropped returns the number of points that have been dropped due to size limits
// since the batch was created or since the last call to DroppedReset.
func (b *Batch) Dropped() int64 {
	return b.dropped
}

// DroppedReset returns the number of points dropped since the last call and resets the counter.
func (b *Batch) DroppedReset() int64 {
	d := b.dropped
	b.dropped = 0
	return d
}