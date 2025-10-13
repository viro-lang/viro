package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC011_DecimalArithmetic validates Feature 002 - User Story 1
// Success Criteria SC-011: Decimal arithmetic operations maintain precision
func TestSC011_DecimalArithmetic(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	tests := []struct {
		name     string
		input    string
		expected string
		setup    []string
	}{
		{
			name:     "Decimal literal",
			input:    `19.99`,
			expected: "19.99",
		},
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
			name:     "Decimal literal multiplication with integer",
			input:    `total`,
			expected: "59.97",
			setup: []string{
				`price: 19.99`,
				`qty: 3`,
				`total: price * qty`,
			},
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
			name:     "Sqrt function with literal",
			input:    `sqrt 4.0`,
			expected: "2",
		},
		{
			name:     "Sqrt function",
			input:    `sqrt x`,
			expected: "2",
			setup:    []string{`x: decimal "4.0"`},
		},
		{
			name:     "Scientific notation",
			input:    `1.5e2`,
			expected: "1.5E+2", // Decimal library may normalize to exponential form
		},
		{
			name:     "Negative decimal",
			input:    `-3.14`,
			expected: "-3.14",
		},
		{
			name:     "Ceil function",
			input:    `ceil 3.14`,
			expected: "4",
		},
		{
			name:     "Floor function",
			input:    `floor 3.99`,
			expected: "3",
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
