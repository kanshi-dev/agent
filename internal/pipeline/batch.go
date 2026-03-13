package pipeline

import "github.com/kanshi-dev/agent/internal/collect"

// Batch provides an in-memory buffer for collected metric points.
type Batch struct {
	points []collect.Point
}

// Add appends points to the batch.
func (b *Batch) Add(points []collect.Point) {
	b.points = append(b.points, points...)
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
