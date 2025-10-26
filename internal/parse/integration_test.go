package parse

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestComplexInputScenarios tests various complex input scenarios
func TestComplexInputScenarios(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "mixed whitespace",
			input:       "  42   +   24  ",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 3 {
					t.Errorf("Expected 3 values (flat sequence), got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeInteger {
					t.Errorf("Expected integer, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should handle mixed whitespace correctly",
		},
		{
			name:        "very long input",
			input:       generateLongInput(1000), // 1000 tokens
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) == 0 {
					t.Errorf("Expected some values, got 0")
				}
			},
			desc: "Should handle very long input strings",
		},
		{
			name:        "unicode identifiers",
			input:       "α: 42 β: α + 1 γ: β * 2",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) < 6 {
					t.Errorf("Expected at least 6 values, got %d", len(vals))
				}
			},
			desc: "Should handle Unicode characters in identifiers",
		},
		{
			name:        "mixed case keywords",
			input:       "True False NONE",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 3 {
					t.Errorf("Expected 3 values, got %d", len(vals))
					return
				}
				// Should be treated as regular words, not keywords
				for i, val := range vals {
					if val.GetType() != value.TypeWord {
						t.Errorf("Value %d should be word, got %s", i, value.TypeToString(val.GetType()))
					}
				}
			},
			desc: "Should handle mixed case keywords as regular words",
		},
		{
			name:        "deeply nested structures",
			input:       generateNestedBlocks(5), // 5 levels deep
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeBlock {
					t.Errorf("Expected block, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should handle deeply nested structures",
		},
		{
			name:        "complex expressions",
			input:       "result: ((a: 1 + 2) * (b: 3 + 4)) / (c: 5 - 1)",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				// This parses as multiple statements
				if len(vals) < 2 {
					t.Errorf("Expected at least 2 values, got %d", len(vals))
				}
			},
			desc: "Should handle complex nested expressions",
		},
		{
			name:        "mixed data types",
			input:       `42 "hello" true false none integer! [1 2 3] (4 5 6)`,
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) < 8 {
					t.Errorf("Expected at least 8 values, got %d", len(vals))
					return
				}
				// Check that we have the expected types somewhere in the result
				hasInteger := false
				hasString := false
				hasWord := false
				hasDatatype := false
				hasBlock := false
				hasParen := false

				for _, val := range vals {
					switch val.GetType() {
					case value.TypeInteger:
						hasInteger = true
					case value.TypeString:
						hasString = true
					case value.TypeWord:
						hasWord = true
					case value.TypeDatatype:
						hasDatatype = true
					case value.TypeBlock:
						hasBlock = true
					case value.TypeParen:
						hasParen = true
					}
				}

				if !hasInteger || !hasString || !hasWord || !hasDatatype || !hasBlock || !hasParen {
					t.Errorf("Missing expected types in parsed values")
				}
			},
			desc: "Should handle mixed data types in sequence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.desc)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.desc, err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, vals)
			}
		})
	}
}

// TestParserStateManagement tests parser state and edge cases
func TestParserStateManagement(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "empty input",
			input:       "",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 0 {
					t.Errorf("Expected 0 values for empty input, got %d", len(vals))
				}
			},
			desc: "Should handle empty input gracefully",
		},
		{
			name:        "only whitespace",
			input:       "   \t\n\r   ",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 0 {
					t.Errorf("Expected 0 values for whitespace-only input, got %d", len(vals))
				}
			},
			desc: "Should handle whitespace-only input",
		},
		{
			name:        "single character inputs",
			input:       "a",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeWord {
					t.Errorf("Expected word, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should handle single character inputs",
		},
		{
			name:        "maximum integer",
			input:       "9223372036854775807", // max int64
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeInteger {
					t.Errorf("Expected integer, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should handle maximum integer values",
		},
		{
			name:        "minimum integer",
			input:       "-9223372036854775808", // min int64
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeInteger {
					t.Errorf("Expected integer, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should handle minimum integer values",
		},
		{
			name:        "consecutive operators",
			input:       "1 + + 2",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				// This parses as some values (exact count depends on parser behavior)
				if len(vals) == 0 {
					t.Errorf("Expected some values, got 0")
				}
			},
			desc: "Should handle consecutive operators",
		},
		{
			name:        "mixed delimiters",
			input:       "[1 2 3] (4 5 6) {invalid}",
			expectError: true, // { } are not valid delimiters
			checkResult: nil,
			desc:        "Should reject invalid delimiters",
		},
		{
			name:        "unmatched delimiters",
			input:       "[1 2 3 (4 5 6",
			expectError: true,
			checkResult: nil,
			desc:        "Should detect unmatched delimiters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.desc)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.desc, err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, vals)
			}
		})
	}
}

// Helper functions for generating test data

// generateLongInput creates a string with n repeated tokens
func generateLongInput(n int) string {
	var parts []string
	for i := range n {
		parts = append(parts, "x")
		if i < n-1 {
			parts = append(parts, "+")
		}
	}
	return strings.Join(parts, " ")
}

// generateNestedBlocks creates a string with n levels of nested blocks
func generateNestedBlocks(depth int) string {
	if depth <= 0 {
		return "42"
	}
	return "[" + generateNestedBlocks(depth-1) + "]"
}
