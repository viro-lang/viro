package parse

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestFormatFunction tests the Format function with various value types
func TestFormatFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    core.Value
		expected string
		desc     string
	}{
		{
			name:     "integer",
			input:    value.IntVal(42),
			expected: "42",
			desc:     "Should format integers correctly",
		},
		{
			name:     "negative integer",
			input:    value.IntVal(-123),
			expected: "-123",
			desc:     "Should format negative integers correctly",
		},
		{
			name:     "string",
			input:    value.StrVal("hello"),
			expected: "\"hello\"",
			desc:     "Should format strings with quotes",
		},
		{
			name:     "string with quotes",
			input:    value.StrVal("he\"llo"),
			expected: "\"he\"llo\"",
			desc:     "Should include quotes in strings",
		},
		{
			name:     "word",
			input:    value.WordVal("variable"),
			expected: "variable",
			desc:     "Should format words without quotes",
		},
		{
			name:     "set-word",
			input:    value.SetWordVal("variable"),
			expected: "variable:",
			desc:     "Should format set-words with colon",
		},
		{
			name:     "get-word",
			input:    value.GetWordVal("variable"),
			expected: ":variable",
			desc:     "Should format get-words with colon prefix",
		},
		{
			name:     "lit-word",
			input:    value.LitWordVal("variable"),
			expected: "'variable",
			desc:     "Should format lit-words with apostrophe",
		},
		{
			name:     "logic true",
			input:    value.LogicVal(true),
			expected: "true",
			desc:     "Should format true as 'true'",
		},
		{
			name:     "logic false",
			input:    value.LogicVal(false),
			expected: "false",
			desc:     "Should format false as 'false'",
		},
		{
			name:     "none",
			input:    value.NoneVal(),
			expected: "none",
			desc:     "Should format none as 'none'",
		},
		{
			name:     "datatype",
			input:    value.DatatypeVal("integer"),
			expected: "unknown",
			desc:     "Should format unknown types as 'unknown'",
		},
		{
			name:     "empty block",
			input:    value.BlockVal([]core.Value{}),
			expected: "[]",
			desc:     "Should format empty blocks",
		},
		{
			name:     "block with elements",
			input:    value.BlockVal([]core.Value{value.IntVal(1), value.IntVal(2), value.IntVal(3)}),
			expected: "[1 2 3]",
			desc:     "Should format blocks with elements",
		},
		{
			name:     "nested blocks",
			input:    value.BlockVal([]core.Value{value.BlockVal([]core.Value{value.IntVal(1)}), value.IntVal(2)}),
			expected: "[[1] 2]",
			desc:     "Should format nested blocks",
		},
		{
			name:     "empty paren",
			input:    value.ParenVal([]core.Value{}),
			expected: "()",
			desc:     "Should format empty parens",
		},
		{
			name:     "paren with elements",
			input:    value.ParenVal([]core.Value{value.IntVal(1), value.WordVal("+"), value.IntVal(2)}),
			expected: "(1 + 2)",
			desc:     "Should format parens with elements",
		},
		{
			name:     "mixed block and paren",
			input:    value.BlockVal([]core.Value{value.ParenVal([]core.Value{value.IntVal(1)}), value.IntVal(2)}),
			expected: "[(1) 2]",
			desc:     "Should format mixed blocks and parens",
		},
		{
			name: "simple path",
			input: func() core.Value {
				vals, _ := ParseEval("obj.field")
				return vals[0]
			}(),
			expected: "path[obj.field]",
			desc:     "Should format simple paths",
		},
		{
			name: "path with index",
			input: func() core.Value {
				vals, _ := ParseEval("array.5")
				return vals[0]
			}(),
			expected: "path[array.5]",
			desc:     "Should format paths with numeric indices",
		},
		{
			name: "complex path",
			input: func() core.Value {
				vals, _ := ParseEval("obj.sub.field")
				return vals[0]
			}(),
			expected: "path[obj.sub.field]",
			desc:     "Should format complex nested paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(tt.input)
			if result != tt.expected {
				t.Errorf("Format() = %q, expected %q for %s", result, tt.expected, tt.desc)
			}
		})
	}
}

// TestFormatEdgeCases tests edge cases and error conditions for Format function
func TestFormatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    core.Value
		expected string
		desc     string
	}{
		{
			name:     "string with newlines",
			input:    value.StrVal("line1\nline2"),
			expected: "\"line1\nline2\"",
			desc:     "Should preserve newlines in strings",
		},
		{
			name:     "string with tabs",
			input:    value.StrVal("col1\tcol2"),
			expected: "\"col1\tcol2\"",
			desc:     "Should preserve tabs in strings",
		},
		{
			name:     "string with backslash",
			input:    value.StrVal("path\\to\\file"),
			expected: "\"path\\to\\file\"",
			desc:     "Should preserve backslashes in strings",
		},
		{
			name: "large block",
			input: func() core.Value {
				elements := make([]core.Value, 10)
				for i := range 10 {
					elements[i] = value.IntVal(int64(i))
				}
				return value.BlockVal(elements)
			}(),
			expected: "[0 1 2 3 4 5 6 7 8 9]",
			desc:     "Should format large blocks correctly",
		},
		{
			name:     "deeply nested structures",
			input:    value.BlockVal([]core.Value{value.ParenVal([]core.Value{value.BlockVal([]core.Value{value.IntVal(1)})})}),
			expected: "[([1])]",
			desc:     "Should format deeply nested structures",
		},
		{
			name: "block with mixed types",
			input: value.BlockVal([]core.Value{
				value.IntVal(42),
				value.StrVal("hello"),
				value.WordVal("world"),
				value.LogicVal(true),
				value.NoneVal(),
			}),
			expected: "[42 \"hello\" world true none]",
			desc:     "Should format blocks with mixed value types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(tt.input)
			if result != tt.expected {
				t.Errorf("Format() = %q, expected %q for %s", result, tt.expected, tt.desc)
			}
		})
	}
}

// TestParseEvalFunction tests the ParseEval function
func TestParseEvalFunction(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "single expression",
			input:       "42 + 24",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse single expressions",
		},
		{
			name:        "simple value",
			input:       "hello",
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
			desc: "Should parse simple values",
		},
		{
			name:        "multiple statements",
			input:       "x: 1 y: 2 x + y",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				// ParseEval should handle multiple values like Parse does
				if len(vals) < 3 {
					t.Errorf("Expected at least 3 values, got %d", len(vals))
				}
			},
			desc: "Should handle multiple statements",
		},
		{
			name:        "error cases",
			input:       "[unclosed",
			expectError: true,
			checkResult: nil,
			desc:        "Should propagate parse errors",
		},
		{
			name:        "empty input",
			input:       "",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 0 {
					t.Errorf("Expected 0 values for empty input, got %d", len(vals))
				}
			},
			desc: "Should handle empty input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := ParseEval(tt.input)

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

// TestFormatParseRoundTrip tests that Format and Parse are compatible
func TestFormatParseRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input core.Value
		desc  string
	}{
		{
			name:  "integer",
			input: value.IntVal(42),
			desc:  "Should round-trip integers",
		},
		{
			name:  "string",
			input: value.StrVal("test"),
			desc:  "Should round-trip strings",
		},
		{
			name:  "word",
			input: value.WordVal("variable"),
			desc:  "Should round-trip words",
		},
		{
			name:  "block",
			input: value.BlockVal([]core.Value{value.IntVal(1), value.IntVal(2)}),
			desc:  "Should round-trip blocks",
		},
		{
			name:  "paren",
			input: value.ParenVal([]core.Value{value.IntVal(1), value.IntVal(2)}),
			desc:  "Should round-trip simple parens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Format the value
			formatted := Format(tt.input)

			// Parse it back
			parsed, err := ParseEval(formatted)
			if err != nil {
				t.Errorf("Failed to parse formatted value %q: %v", formatted, err)
				return
			}

			if len(parsed) != 1 {
				t.Errorf("Expected 1 parsed value, got %d", len(parsed))
				return
			}

			// Check if they match (this is a basic check - full equality might be complex)
			parsedFormatted := Format(parsed[0])
			if formatted != parsedFormatted {
				t.Errorf("Round-trip failed for %s: original %q != parsed %q", tt.desc, formatted, parsedFormatted)
			}
		})
	}
}
