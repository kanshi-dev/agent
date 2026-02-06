package pipeline

import "github.com/kanshi-dev/agent/internal/collect"

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

// Flush returns the points in the batch and clears it.
func (b *Batch) Flush() []collect.Point {
	out := b.points
	b.points = nil
	return out
}

func (b *Batch) Snapshot() []collect.Point {
	return b.points
}

func (b *Batch) Clear() {
	b.points = nil
}
