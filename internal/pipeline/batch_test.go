package pipeline

import (
	"testing"

	"github.com/kanshi-dev/agent/internal/collect"
)

func TestBatchSizeLimit(t *testing.T) {
	b := NewBatch(3) // Batch size of 3

	// Add exactly 3 points - should all be added
	points := []collect.Point{
		{Name: "test1", Value: 1.0},
		{Name: "test2", Value: 2.0},
		{Name: "test3", Value: 3.0},
	}
	b.Add(points)
	if got := b.Len(); got != 3 {
		t.Errorf("Expected length 3, got %d", got)
	}

	// Try to add more - should be dropped since batch is full
	morePoints := []collect.Point{
		{Name: "test4", Value: 4.0},
		{Name: "test5", Value: 5.0},
	}
	b.Add(morePoints)
	if got := b.Len(); got != 3 {
		t.Errorf("Expected length 3 (batch full), got %d", got)
	}

	// Flush and verify we get the original 3 points
	flushed := b.Flush()
	if got := len(flushed); got != 3 {
		t.Errorf("Expected 3 points on flush, got %d", got)
	}
	if got := b.Len(); got != 0 {
		t.Errorf("Expected length 0 after flush, got %d", got)
	}

	// Add fewer points than capacity
	fewPoints := []collect.Point{
		{Name: "test6", Value: 6.0},
		{Name: "test7", Value: 7.0},
	}
	b.Add(fewPoints)
	if got := b.Len(); got != 2 {
		t.Errorf("Expected length 2, got %d", got)
	}

	// Add more than remaining capacity - should only fill to capacity
	lotsOfPoints := []collect.Point{
		{Name: "test8", Value: 8.0},
		{Name: "test9", Value: 9.0},
		{Name: "test10", Value: 10.0},
		{Name: "test11", Value: 11.0},
	}
	b.Add(lotsOfPoints)
	if got := b.Len(); got != 3 {
		t.Errorf("Expected length 3 (filled to capacity), got %d", got)
	}

	// Verify the contents are what we expect (first 2 + first 1 from lotsOfPoints)
	flushed = b.Flush()
	if got := len(flushed); got != 3 {
		t.Errorf("Expected 3 points on flush, got %d", got)
	}
	// Check first two are from fewPoints
	if flushed[0].Name != "test6" || flushed[0].Value != 6.0 {
		t.Errorf("Expected first point to be test6=6.0, got %v=%v", flushed[0].Name, flushed[0].Value)
	}
	if flushed[1].Name != "test7" || flushed[1].Value != 7.0 {
		t.Errorf("Expected second point to be test7=7.0, got %v=%v", flushed[1].Name, flushed[1].Value)
	}
	// Check third is from lotsOfPoints (should be test8=8.0 since we take first available)
	if flushed[2].Name != "test8" || flushed[2].Value != 8.0 {
		t.Errorf("Expected third point to be test8=8.0, got %v=%v", flushed[2].Name, flushed[2].Value)
	}
}

func TestBatchZeroSize(t *testing.T) {
	b := NewBatch(0) // Batch size of 0

	points := []collect.Point{
		{Name: "test1", Value: 1.0},
		{Name: "test2", Value: 2.0},
	}
	b.Add(points)
	if got := b.Len(); got != 0 {
		t.Errorf("Expected length 0 for zero-size batch, got %d", got)
	}
}

func TestBatchNegativeSize(t *testing.T) {
	// Negative size should behave like zero size
	b := NewBatch(-1)

	points := []collect.Point{
		{Name: "test1", Value: 1.0},
	}
	b.Add(points)
	if got := b.Len(); got != 0 {
		t.Errorf("Expected length 0 for negative-size batch, got %d", got)
	}
}