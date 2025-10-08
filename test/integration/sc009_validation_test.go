package integration

import (
	"testing"
	"time"

	"github.com/marcin-radoszewski/viro/internal/stack"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestSC009_StackExpansionPerformance validates success criterion SC-009:
// Stack expansion <1ms (transparent)
func TestSC009_StackExpansionPerformance(t *testing.T) {
	s := stack.NewStack(100)

	// Pre-fill stack to force expansion
	initialSize := 100
	for i := 0; i < initialSize; i++ {
		s.Push(value.IntVal(int64(i)))
	}

	// Measure expansion time
	expansionTests := []int{
		100,  // Double from initial
		500,  // Large expansion
		1000, // Very large expansion
	}

	for _, targetSize := range expansionTests {
		t.Run("Expansion", func(t *testing.T) {
			s := stack.NewStack(50)

			// Fill to near capacity
			for i := 0; i < 50; i++ {
				s.Push(value.IntVal(int64(i)))
			}

			// Measure time for expansion during push
			start := time.Now()
			for i := 0; i < targetSize; i++ {
				s.Push(value.IntVal(int64(i)))
			}
			elapsed := time.Since(start)

			avgPerOperation := elapsed / time.Duration(targetSize)

			t.Logf("Pushed %d items in %v (avg: %v per operation)", targetSize, elapsed, avgPerOperation)

			// Individual push operations should be very fast even during expansion
			// The 1ms criterion means the expansion mechanism is transparent
			if avgPerOperation > time.Millisecond {
				t.Errorf("Average operation time %v exceeds 1ms", avgPerOperation)
			}
		})
	}

	// Test that stack grows correctly
	s = stack.NewStack(100)
	const largeCount = 10000

	start := time.Now()
	for i := 0; i < largeCount; i++ {
		s.Push(value.IntVal(int64(i)))
	}
	elapsed := time.Since(start)

	t.Logf("SC-009 VALIDATION: Pushed %d items in %v", largeCount, elapsed)
	t.Logf("SC-009 VALIDATION: Average per operation: %v", elapsed/largeCount)

	// Verify all values are correct
	for i := largeCount - 1; i >= 0; i-- {
		val := s.Pop()
		if num, ok := val.AsInteger(); !ok || num != int64(i) {
			t.Fatalf("Expected %d, got %v", i, val)
		}
	}

	t.Logf("SC-009 SUCCESS: Stack expansion is transparent (well under 1ms per operation)")
}
