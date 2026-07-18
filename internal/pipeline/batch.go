package pipeline

import "github.com/kanshi-dev/agent/internal/collect"

// Batch provides an in-memory buffer for collected metric points with a maximum size.
type Batch struct {
	points []collect.Point
	max    int
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
// If adding the points would exceed the maximum, excess points are dropped.
func (b *Batch) Add(points []collect.Point) {
	if len(b.points) >= b.max {
		// Batch is already at or over capacity, drop all new points
		return
	}

	available := b.max - len(b.points)
	if len(points) > available {
		// Not enough space for all points, take only what fits
		b.points = append(b.points, points[:available]...)
	} else {
		// Enough space for all points
		b.points = append(b.points, points...)
	}
}

// Len returns the number of points in the batch.
func (b *Batch) Len() int {
	return len(b.points)
}

// Flush returns and clears all points currently in the batch.
func (b *Batch) Flush() []collect.Point {
	out := b.points
	b.points = nil
	return out
}