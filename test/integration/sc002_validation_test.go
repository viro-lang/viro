package integration

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC002_MemoryStability validates success criterion SC-002:
// REPL session remains stable for continuous operation exceeding 1000 evaluation
// cycles without memory leaks or crashes
func TestSC002_MemoryStability(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Force garbage collection and get baseline memory
	runtime.GC()
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Test expressions that exercise different parts of the system
	testExpressions := []string{
		"42",                       // Literals
		"x: 10",                    // Assignment
		"x",                        // Word lookup
		"3 + 4",                    // Arithmetic
		"[1 2 3]",                  // Blocks
		"(5 * 2)",                  // Parens
		"data: [1 2 3]",            // Series creation
		"first data",               // Series access
		"append data 4",            // Series modification
		"when true [42]",           // Control flow
		"if false [1] [2]",         // Conditionals
		"square: fn [n] [(* n n)]", // Function definition
		"square 5",                 // Function call
		"3 < 5",                    // Comparisons
		"and true false",           // Logic
		"type? 42",                 // Type queries
	}

	const targetCycles = 1000
	successfulCycles := 0

	// Run 1000+ evaluation cycles
	for i := 0; i < targetCycles; i++ {
		expr := testExpressions[i%len(testExpressions)]
		out.Reset()
		loop.EvalLineForTest(expr)

		// Check that evaluation didn't crash
		if out.Len() > 0 {
			successfulCycles++
		}

		// Periodic GC to test for leaks
		if i%100 == 0 {
			runtime.GC()
		}
	}

	// Final GC and memory check
	runtime.GC()
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory growth
	memGrowthBytes := int64(memStatsAfter.Alloc) - int64(memStatsBefore.Alloc)
	memGrowthMB := float64(memGrowthBytes) / (1024 * 1024)

	t.Logf("SC-002 VALIDATION: Completed %d evaluation cycles", successfulCycles)
	t.Logf("SC-002 VALIDATION: Memory before: %.2f MB", float64(memStatsBefore.Alloc)/(1024*1024))
	t.Logf("SC-002 VALIDATION: Memory after: %.2f MB", float64(memStatsAfter.Alloc)/(1024*1024))
	t.Logf("SC-002 VALIDATION: Memory growth: %.2f MB", memGrowthMB)

	// Validate success criteria
	if successfulCycles < targetCycles {
		t.Fatalf("SC-002 FAILED: Only %d cycles completed successfully, need %d", successfulCycles, targetCycles)
	}

	// Memory growth should be reasonable (less than 50MB for 1000 cycles is acceptable)
	if memGrowthMB > 50 {
		t.Errorf("SC-002 WARNING: Significant memory growth detected: %.2f MB", memGrowthMB)
	}

	t.Logf("SC-002 SUCCESS: REPL stable for 1000+ evaluation cycles without crashes")
}
