package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC011_DecimalArithmetic validates Feature 002 - User Story 1
// Success Criteria SC-011: Decimal arithmetic operations maintain precision
func TestSC011_DecimalArithmetic(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	tests := []struct {
		name     string
		input    string
		expected string
		setup    []string
	}{
		{
			name:     "Decimal constructor from string",
			input:    `decimal "19.99"`,
			expected: "19.99",
		},
		{
			name:     "Decimal constructor from integer",
			input:    `x`,
			expected: "42",
			setup:    []string{`x: decimal 42`},
		},
		{
			name:     "Decimal multiplication with integer promotion",
			input:    `total`,
			expected: "59.97",
			setup: []string{
				`price: decimal "19.99"`,
				`qty: 3`,
				`total: price * qty`,
			},
		},
		{
			name:     "Sqrt function",
			input:    `sqrt x`,
			expected: "2",
			setup:    []string{`x: decimal "4.0"`},
		},
		{
			name:     "Round function",
			input:    `round x 2`,
			expected: "3.14",
			setup:    []string{`x: decimal "3.14159"`},
		},
		{
			name:     "Ceil function",
			input:    `ceil x`,
			expected: "4",
			setup:    []string{`x: decimal "3.14"`},
		},
		{
			name:     "Floor function",
			input:    `floor x`,
			expected: "3",
			setup:    []string{`x: decimal "3.99"`},
		},
		{
			name:     "Truncate function",
			input:    `truncate x`,
			expected: "-3",
			setup:    []string{`x: decimal "-3.99"`},
		},
		{
			name:     "Pow function",
			input:    `result`,
			expected: "8",
			setup: []string{
				`base: decimal "2.0"`,
				`exp: decimal "3.0"`,
				`result: pow base exp`,
			},
		},
		{
			name:     "Backward compatibility - integer arithmetic",
			input:    `a * b`,
			expected: "15",
			setup: []string{
				`a: 5`,
				`b: 3`,
			},
		},
	}

	passedTests := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup commands if any
			for _, setupCmd := range tt.setup {
				out.Reset()
				loop.EvalLineForTest(setupCmd)
			}

			// Execute test
			out.Reset()
			loop.EvalLineForTest(tt.input)
			result := strings.TrimSpace(out.String())

			if !strings.Contains(result, tt.expected) {
				t.Errorf("%s: expected to contain %q, got %q", tt.name, tt.expected, result)
			} else {
				passedTests++
				t.Logf("SC-011 PASS: %s", tt.name)
			}
		})
	}

	t.Logf("SC-011 SUCCESS: %d/%d decimal arithmetic tests passed", passedTests, len(tests))
}
