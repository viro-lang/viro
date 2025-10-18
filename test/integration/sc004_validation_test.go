package integration

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC004_RecursionDepth validates success criterion SC-004:
// Users can define and call recursive functions up to depth of 100 without stack overflow
func TestSC004_RecursionDepth(t *testing.T) {
	evaluator := NewTestEvaluator()
	var errOut bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &errOut)

	// Define a recursive countdown function
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	errOut.Reset()
	loop.EvalLineForTest("countdown: fn [n] [when (> n 0) [(countdown (- n 1))]]")
	w.Close()
	output, _ := io.ReadAll(r)
	result := strings.TrimSpace(string(output))
	os.Stdout = oldStdout

	if !strings.Contains(result, "function[countdown]") {
		t.Fatalf("Failed to define recursive function: %s", result)
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
			// Capture stdout for the countdown call
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			errOut.Reset()
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
			w.Close()
			_, _ = io.ReadAll(r)
			os.Stdout = oldStdout

			// Check if it completed without error (no error in errOut)
			if errOut.Len() == 0 {
				if tt.depth > maxDepthAchieved {
					maxDepthAchieved = tt.depth
				}
				t.Logf("Successfully handled recursion depth %d", tt.depth)
			} else {
				t.Logf("Recursion depth %d failed or hit limit: %s", tt.depth, errOut.String())
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
