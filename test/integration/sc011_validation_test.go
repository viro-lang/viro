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
			name:     "Decimal addition",
			input:    `19.99 + 0.01`,
			expected: "20",
		},
		{
			name:     "Decimal subtraction",
			input:    `20.00 - 0.01`,
			expected: "19.99",
		},
		{
			name:     "Decimal multiplication",
			input:    `10.0 * 2.0`,
			expected: "20",
		},
		{
			name:     "Decimal division",
			input:    `20.0 / 2.0`,
			expected: "10",
		},
		{
			name:     "Decimal comparison equal",
			input:    `19.99 = 19.99`,
			expected: "true",
		},
		{
			name:     "Decimal comparison not equal",
			input:    `19.99 = 20.00`,
			expected: "false",
		},
		{
			name:     "Decimal comparison less than",
			input:    `19.99 < 20.00`,
			expected: "true",
		},
		{
			name:     "Decimal comparison greater than",
			input:    `20.00 > 19.99`,
			expected: "true",
		},
		{
			name:     "Decimal comparison less or equal",
			input:    `19.99 <= 19.99`,
			expected: "true",
		},
		{
			name:     "Decimal comparison greater or equal",
			input:    `20.00 >= 19.99`,
			expected: "true",
		},
	}

	passedTests := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup commands if any (discard their output)
			for _, setupCmd := range tt.setup {
				out.Reset()
				loop.EvalLineForTest(setupCmd)
				// Setup output is discarded
			}

			// Execute test command
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
