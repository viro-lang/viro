package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC001_ExpressionTypesCoverage validates success criterion SC-001:
// Users can evaluate at least 20 different expression types correctly
// (literals, arithmetic, comparisons, control flow, series operations, function calls)
func TestSC001_ExpressionTypesCoverage(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	tests := []struct {
		name     string
		input    string
		expected string
		setup    []string // Optional setup commands to run first
	}{
		// 1. Integer literal
		{
			name:     "Integer literal",
			input:    "42",
			expected: "42",
		},
		// 2. String literal
		{
			name:     "String literal",
			input:    "\"hello\"",
			expected: "\"hello\"",
		},
		// 3. Logic literal (true)
		{
			name:     "Logic literal true",
			input:    "true",
			expected: "true",
		},
		// 4. Logic literal (false)
		{
			name:     "Logic literal false",
			input:    "false",
			expected: "false",
		},
		// 5. Set-word assignment
		{
			name:     "Set-word assignment",
			input:    "x: 10",
			expected: "10",
		},
		// 6. Word lookup
		{
			name:     "Word lookup",
			input:    "x",
			expected: "10",
			setup:    []string{"x: 10"},
		},
		// 7. Block literal
		{
			name:     "Block literal",
			input:    "[1 2 3]",
			expected: "[1 2 3]",
		},
		// 8. Paren evaluation
		{
			name:     "Paren evaluation",
			input:    "(1 + 2)",
			expected: "3",
		},
		// 9. Addition
		{
			name:     "Addition",
			input:    "3 + 4",
			expected: "7",
		},
		// 10. Subtraction
		{
			name:     "Subtraction",
			input:    "10 - 3",
			expected: "7",
		},
		// 11. Multiplication
		{
			name:     "Multiplication",
			input:    "6 * 7",
			expected: "42",
		},
		// 12. Division
		{
			name:     "Division",
			input:    "20 / 4",
			expected: "5",
		},
		// 13. Less than comparison
		{
			name:     "Less than comparison",
			input:    "3 < 5",
			expected: "true",
		},
		// 14. Greater than comparison
		{
			name:     "Greater than comparison",
			input:    "5 > 3",
			expected: "true",
		},
		// 15. Equal comparison
		{
			name:     "Equal comparison",
			input:    "5 = 5",
			expected: "true",
		},
		// 16. Not equal comparison
		{
			name:     "Not equal comparison",
			input:    "5 <> 3",
			expected: "true",
		},
		// 17. Logic AND
		{
			name:     "Logic AND",
			input:    "and true false",
			expected: "false",
		},
		// 18. Logic OR
		{
			name:     "Logic OR",
			input:    "or true false",
			expected: "true",
		},
		// 19. Logic NOT
		{
			name:     "Logic NOT",
			input:    "not false",
			expected: "true",
		},
		// 20. First operation
		{
			name:     "First series operation",
			input:    "first data",
			expected: "1",
			setup:    []string{"data: [1 2 3]"},
		},
		// 21. Last operation
		{
			name:     "Last series operation",
			input:    "last data",
			expected: "3",
			setup:    []string{"data: [1 2 3]"},
		},
		// 22. Length? operation
		{
			name:     "Length? series operation",
			input:    "length? data",
			expected: "3",
			setup:    []string{"data: [1 2 3]"},
		},
		// 23. Append operation
		{
			name:     "Append series operation",
			input:    "append data 4",
			expected: "[1 2 3 4]",
			setup:    []string{"data: [1 2 3]"},
		},
		// 24. Insert operation
		{
			name:     "Insert series operation",
			input:    "insert data 0",
			expected: "[0 1 2 3]",
			setup:    []string{"data: [1 2 3]"},
		},
		// 25. When control flow
		{
			name:     "When control flow",
			input:    "when true [42]",
			expected: "42",
		},
		// 26. If control flow
		{
			name:     "If control flow",
			input:    "if false [1] [2]",
			expected: "2",
		},
		// 27. Loop control flow
		{
			name:     "Loop control flow",
			input:    "counter",
			expected: "3",
			setup: []string{
				"counter: 0",
				"loop 3 [counter: (+ counter 1)]",
			},
		},
		// 28. Function definition
		{
			name:     "Function definition",
			input:    "square: fn [n] [(* n n)]",
			expected: "function[square]",
		},
		// 29. Function call
		{
			name:     "Function call",
			input:    "square 5",
			expected: "25",
			setup:    []string{"square: fn [n] [(* n n)]"},
		},
		// 30. Type? query
		{
			name:     "Type? query",
			input:    "type? 42",
			expected: "integer!",
		},
		// 31. Left-to-right arithmetic evaluation
		{
			name:     "Left-to-right arithmetic",
			input:    "3 + 4 * 2",
			expected: "14", // (3 + 4) * 2 = 7 * 2 = 14
		},
		// 32. Complex paren expression
		{
			name:     "Complex paren expression",
			input:    "((5 + 3) * 2)",
			expected: "16",
		},
		// 33. Get-word
		{
			name:     "Get-word evaluation",
			input:    ":y",
			expected: "20",
			setup:    []string{"y: 20"},
		},
	}

	expressionTypeCount := 0
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

			if result != tt.expected {
				t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, result)
			} else {
				passedTests++
			}
		})
		expressionTypeCount++
	}

	// Validate SC-001 criterion: at least 20 different expression types
	if expressionTypeCount < 20 {
		t.Fatalf("SC-001 FAILED: Only %d expression types tested, need at least 20", expressionTypeCount)
	}

	t.Logf("SC-001 VALIDATION: %d expression types tested successfully", expressionTypeCount)
	t.Logf("SC-001 VALIDATION: %d/%d tests passed", passedTests, expressionTypeCount)

	if passedTests < 20 {
		t.Fatalf("SC-001 FAILED: Only %d expression types working correctly, need at least 20", passedTests)
	}

	t.Logf("SC-001 SUCCESS: At least 20 different expression types evaluate correctly")
}
