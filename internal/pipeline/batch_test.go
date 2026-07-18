package pipeline

import (
	"testing"

	"github.com/kanshi-dev/agent/internal/collect"
)

func TestBatchSizeLimitKeepsNewest(t *testing.T) {
	b := NewBatch(3) // Batch size of 3

	// Add exactly 3 points - should all be added
	points1 := []collect.Point{
		{Name: "test1", Value: 1.0},
		{Name: "test2", Value: 2.0},
		{Name: "test3", Value: 3.0},
	}
	if dropped := b.Add(points1); dropped != 0 {
		t.Errorf("Expected 0 drops when adding 3 to empty batch, got %d", dropped)
	}
	if got := b.Len(); got != 3 {
		t.Errorf("Expected length 3, got %d", got)
	}
	// To verify content without losing batch, flush and restore
	flushed := b.Flush()
	if len(flushed) != 3 {
		t.Errorf("Expected 3 points on flush, got %d", len(flushed))
	} else {
		if flushed[0].Name != "test1" || flushed[0].Value != 1.0 {
			t.Errorf("First point expected test1=1.0, got %v=%v", flushed[0].Name, flushed[0].Value)
		}
		if flushed[1].Name != "test2" || flushed[1].Value != 2.0 {
			t.Errorf("Second point expected test2=2.0, got %v=%v", flushed[1].Name, flushed[1].Value)
		}
		if flushed[2].Name != "test3" || flushed[2].Value != 3.0 {
			t.Errorf("Third point expected test3=3.0, got %v=%v", flushed[2].Name, flushed[2].Value)
		}
	}
	// Put them back
	if dropped := b.Add(flushed); dropped != 0 {
		t.Errorf("Expected 0 drops when re-adding flushed points, got %d", dropped)
	}
	if got := b.Len(); got != 3 {
		t.Errorf("Expected length 3 after re-adding, got %d", got)
	}

	// Add 2 more points; batch is full (3), so we should drop the oldest 2 to keep newest 3
	morePoints := []collect.Point{
		{Name: "test4", Value: 4.0},
		{Name: "test5", Value: 5.0},
	}
	if dropped := b.Add(morePoints); dropped != 2 {
		t.Errorf("Expected 2 points dropped when adding 2 to full batch (size 3), got %d", dropped)
	}
	if got := b.Len(); got != 3 {
		t.Errorf("Expected length 3 after adding 2 to full batch, got %d", got)
	}
	// Flush and verify we have the newest three: test3, test4, test5
	flushed = b.Flush()
	if len(flushed) != 3 {
		t.Errorf("Expected 3 points on flush, got %d", len(flushed))
	} else {
		if flushed[0].Name != "test3" || flushed[0].Value != 3.0 {
			t.Errorf("First point expected test3=3.0, got %v=%v", flushed[0].Name, flushed[0].Value)
		}
		if flushed[1].Name != "test4" || flushed[1].Value != 4.0 {
			t.Errorf("Second point expected test4=4.0, got %v=%v", flushed[1].Name, flushed[1].Value)
		}
		if flushed[2].Name != "test5" || flushed[2].Value != 5.0 {
			t.Errorf("Third point expected test5=5.0, got %v=%v", flushed[2].Name, flushed[2].Value)
		}
	}

	// Add 4 points; batch currently empty after flush, capacity 3, so we should keep last 3 (drop oldest 1)
	lotsOfPoints := []collect.Point{
		{Name: "test8", Value: 8.0},
		{Name: "test9", Value: 9.0},
		{Name: "test10", Value: 10.0},
		{Name: "test11", Value: 11.0},
	}
	if dropped := b.Add(lotsOfPoints); dropped != 1 {
		t.Errorf("Expected 1 drop when adding 4 to empty batch (size 3), got %d", dropped)
	}
	if got := b.Len(); got != 3 {
		t.Errorf("Expected length 3 after adding, got %d", got)
	}
	flushed = b.Flush()
	if len(flushed) != 3 {
		t.Errorf("Expected 3 points on flush, got %d", len(flushed))
	} else {
		if flushed[0].Name != "test9" || flushed[0].Value != 9.0 {
			t.Errorf("First point expected test9=9.0, got %v=%v", flushed[0].Name, flushed[0].Value)
		}
		if flushed[1].Name != "test10" || flushed[1].Value != 10.0 {
			t.Errorf("Second point expected test10=10.0, got %v=%v", flushed[1].Name, flushed[1].Value)
		}
		if flushed[2].Name != "test11" || flushed[2].Value != 11.0 {
			t.Errorf("Third point expected test11=11.0, got %v=%v", flushed[2].Name, flushed[2].Value)
		}
	}
}

func TestBatchZeroSize(t *testing.T) {
	b := NewBatch(0) // Batch size of 0

	points := []collect.Point{
		{Name: "test1", Value: 1.0},
		{Name: "test2", Value: 2.0},
	}
	if dropped := b.Add(points); dropped != 2 {
		t.Errorf("Expected 2 points dropped, got %d", dropped)
	}
	if got := b.Len(); got != 0 {
		t.Errorf("Expected length 0 for zero-size batch, got %d", got)
	}
	if dropped := b.Dropped(); dropped != 2 {
		t.Errorf("Expected dropped count 2, got %d", dropped)
	}
	// Reset drop count
	if dropped := b.DroppedReset(); dropped != 2 {
		t.Errorf("Expected dropped reset 2, got %d", dropped)
	}
	if dropped := b.Dropped(); dropped != 0 {
		t.Errorf("Expected dropped count 0 after reset, got %d", dropped)
	}
}

func TestBatchNegativeSize(t *testing.T) {
	// Negative size should behave like zero size
	b := NewBatch(-1)

	points := []collect.Point{
		{Name: "test1", Value: 1.0},
	}
	if dropped := b.Add(points); dropped != 1 {
		t.Errorf("Expected 1 point dropped, got %d", dropped)
	}
	if got := b.Len(); got != 0 {
		t.Errorf("Expected length 0 for negative-size batch, got %d", got)
	}
}

func TestBatchDroppedCounter(t *testing.T) {
	b := NewBatch(2)
	// Add 1, no drop
	if d := b.Add([]collect.Point{{Name: "a", Value: 1}}); d != 0 {
		t.Errorf("Expected 0 drop, got %d", d)
	}
	if d := b.Dropped(); d != 0 {
		t.Errorf("Expected dropped 0, got %d", d)
	}
	// Add another, total 2, still no drop
	if d := b.Add([]collect.Point{{Name: "b", Value: 2}}); d != 0 {
		t.Errorf("Expected 0 drop, got %d", d)
	}
	if d := b.Dropped(); d != 0 {
		t.Errorf("Expected dropped 0, got %d", d)
	}
	// Add third, should drop oldest (a)
	if d := b.Add([]collect.Point{{Name: "c", Value: 3}}); d != 1 {
		t.Errorf("Expected 1 drop, got %d", d)
	}
	if d := b.Dropped(); d != 1 {
		t.Errorf("Expected dropped 1, got %d", d)
	}
	// Flush and check contents
	flushed := b.Flush()
	if len(flushed) != 2 {
		t.Errorf("Expected 2 points after flush, got %d", len(flushed))
	}
	if flushed[0].Name != "b" || flushed[0].Value != 2.0 {
		t.Errorf("First point expected b=2.0, got %v=%v", flushed[0].Name, flushed[0].Value)
	}
	if flushed[1].Name != "c" || flushed[1].Value != 3.0 {
		t.Errorf("Second point expected c=3.0, got %v=%v", flushed[1].Name, flushed[1].Value)
	}
	// Dropped count should still be 1 (not reset by flush)
	if d := b.Dropped(); d != 1 {
		t.Errorf("Expected dropped 1 after flush, got %d", d)
	}
	// Reset
	if d := b.DroppedReset(); d != 1 {
		t.Errorf("Expected drop reset 1, got %d", d)
	}
	if d := b.Dropped(); d != 0 {
		t.Errorf("Expected dropped 0 after reset, got %d", d)
	}
}