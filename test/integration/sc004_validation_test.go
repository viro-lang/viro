package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC004_RecursionDepth validates success criterion SC-004:
// Users can define and call recursive functions up to depth of 100 without stack overflow
func TestSC004_RecursionDepth(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Define a recursive countdown function
	out.Reset()
	loop.EvalLineForTest("countdown: fn [n] [when (> n 0) [(countdown (- n 1))]]")
	if !strings.Contains(out.String(), "function[countdown]") {
		t.Fatalf("Failed to define recursive function: %s", out.String())
	}

	tests := []struct {
		depth    int
		expected string
		name     string
	}{
		{depth: 10, expected: "none", name: "Depth 10"},
		{depth: 50, expected: "none", name: "Depth 50"},
		{depth: 100, expected: "none", name: "Depth 100"},
		{depth: 150, expected: "none", name: "Depth 150"},
	}

	maxDepthAchieved := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out.Reset()
			loop.EvalLineForTest("countdown " + strings.Repeat("1", len(strings.Split(tt.name, " ")[1])))

			// Build the actual countdown call
			out.Reset()
			input := ""
			if tt.depth == 10 {
				input = "countdown 10"
			} else if tt.depth == 50 {
				input = "countdown 50"
			} else if tt.depth == 100 {
				input = "countdown 100"
			} else if tt.depth == 150 {
				input = "countdown 150"
			}

			loop.EvalLineForTest(input)
			result := strings.TrimSpace(out.String())

			// Check if it completed without error
			if !strings.Contains(result, "Error") && !strings.Contains(result, "overflow") {
				if tt.depth > maxDepthAchieved {
					maxDepthAchieved = tt.depth
				}
				t.Logf("Successfully handled recursion depth %d", tt.depth)
			} else {
				t.Logf("Recursion depth %d failed or hit limit: %s", tt.depth, result)
			}
		})
	}

	t.Logf("SC-004 VALIDATION: Maximum recursion depth achieved: %d", maxDepthAchieved)

	// Validate success criteria - need at least 100 depth
	if maxDepthAchieved < 100 {
		t.Fatalf("SC-004 FAILED: Only achieved depth %d, need at least 100", maxDepthAchieved)
	}

	t.Logf("SC-004 SUCCESS: Recursive functions work to depth 100+ without overflow")
}
