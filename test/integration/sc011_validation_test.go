package integration

import (
	"bytes"
	"io"
	"os"
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

	// Redirect stdout to capture print output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

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
			input:    `result`,
			expected: "60",
			setup: []string{
				`a: 20`,
				`b: 3.0`,
				`result: a * b`,
			},
		},
		{
			name:     "Decimal division",
			input:    `result`,
			expected: "6.666666666666666666666666666666667",
			setup: []string{
				`a: 20.0`,
				`b: 3`,
				`result: a / b`,
			},
		},
		{
			name:     "Decimal addition",
			input:    `result`,
			expected: "23",
			setup: []string{
				`a: 20.0`,
				`b: 3`,
				`result: a + b`,
			},
		},
		{
			name:     "Decimal subtraction",
			input:    `result`,
			expected: "17",
			setup: []string{
				`a: 20.0`,
				`b: 3`,
				`result: a - b`,
			},
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
			// Run setup commands if any (discard their output)
			for _, setupCmd := range tt.setup {
				out.Reset()
				loop.EvalLineForTest(setupCmd)
				// Drain any output from setup
				w.Close()
				var drain bytes.Buffer
				io.Copy(&drain, r)
				r, w, _ = os.Pipe()
				os.Stdout = w
			}

			// Execute test command
			out.Reset()
			loop.EvalLineForTest(tt.input)

			// Read captured stdout from the test command only
			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			result := strings.TrimSpace(buf.String())

			if !strings.Contains(result, tt.expected) {
				t.Errorf("%s: expected to contain %q, got %q", tt.name, tt.expected, result)
			} else {
				passedTests++
				t.Logf("SC-011 PASS: %s", tt.name)
			}

			// Reset pipe for next test
			r, w, _ = os.Pipe()
			os.Stdout = w
		})
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	t.Logf("SC-011 SUCCESS: %d/%d decimal arithmetic tests passed", passedTests, len(tests))
}
