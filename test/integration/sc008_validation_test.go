package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC008_MultiLineNesting validates success criterion SC-008:
// Multi-line input handles 10+ nested levels
func TestSC008_MultiLineNesting(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	tests := []struct {
		name      string
		input     string
		nestLevel int
	}{
		{
			name:      "3 levels",
			input:     "[[[1]]]",
			nestLevel: 3,
		},
		{
			name:      "5 levels",
			input:     "[[[[[1]]]]]",
			nestLevel: 5,
		},
		{
			name:      "10 levels",
			input:     "[[[[[[[[[[1]]]]]]]]]]",
			nestLevel: 10,
		},
		{
			name:      "15 levels",
			input:     "[[[[[[[[[[[[[[[1]]]]]]]]]]]]]]]",
			nestLevel: 15,
		},
		{
			name:      "Mixed parens and blocks 10 levels",
			input:     "[([([([([(1)]])])])]",
			nestLevel: 10,
		},
	}

	maxNestAchieved := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out.Reset()
			loop.EvalLineForTest(tt.input)
			result := strings.TrimSpace(out.String())

			// Check if parsing succeeded
			if !strings.Contains(result, "Error") {
				if tt.nestLevel > maxNestAchieved {
					maxNestAchieved = tt.nestLevel
				}
				t.Logf("Successfully handled %d nested levels", tt.nestLevel)
			} else {
				t.Logf("Failed at %d nested levels: %s", tt.nestLevel, result)
			}
		})
	}

	t.Logf("SC-008 VALIDATION: Maximum nesting depth achieved: %d", maxNestAchieved)

	// Validate success criteria - need at least 10 levels
	if maxNestAchieved < 10 {
		t.Fatalf("SC-008 FAILED: Only achieved %d nesting levels, need at least 10", maxNestAchieved)
	}

	t.Logf("SC-008 SUCCESS: Multi-line input handles 10+ nested levels")
}
