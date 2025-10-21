package parse

import (
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestComplexNumberParsing tests scientific notation and edge cases
func TestComplexNumberParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    []core.Value
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "scientific notation e+",
			input:       "1.5e+2",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("150"), 1)},
			desc:        "Should parse positive exponent with explicit +",
		},
		{
			name:        "scientific notation e-",
			input:       "2.5e-3",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("0.0025"), 1)},
			desc:        "Should parse negative exponent",
		},
		{
			name:        "scientific notation E+",
			input:       "1.23E+4",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("12300"), 2)},
			desc:        "Should parse uppercase E with positive exponent",
		},
		{
			name:        "scientific notation E-",
			input:       "6.022E-23",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("6.022e-23"), 3)},
			desc:        "Should parse scientific notation with very small exponent",
		},
		{
			name:        "large exponent",
			input:       "1e100",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("1e+100"), 0)},
			desc:        "Should handle very large exponents",
		},
		{
			name:        "zero with exponent",
			input:       "0e5",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("0"), 0)},
			desc:        "Should parse zero with exponent",
		},
		{
			name:        "decimal with large exponent",
			input:       "1.23456e10",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("12345600000"), 5)},
			desc:        "Should handle decimal with large exponent",
		},
		{
			name:        "invalid exponent format",
			input:       "1.5e",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("1.5"), 1), value.WordVal("e")},
			checkResult: nil,
			desc:        "Should tokenize incomplete exponent as separate tokens",
		},
		{
			name:        "exponent without digit",
			input:       "1.5ee",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("1.5"), 1), value.WordVal("ee")},
			checkResult: nil,
			desc:        "Should tokenize invalid exponent as separate tokens",
		},
		{
			name:        "multiple decimal points",
			input:       "1.2.3",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				// This gets parsed as a path: 1.2.3
				if vals[0].GetType() != value.TypePath {
					t.Errorf("Expected path, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse multiple decimal points as path",
		},
		{
			name:        "exponent without digit",
			input:       "1.5ee",
			expectError: false,                                                  // Parser may be lenient
			expected:    []core.Value{value.DecimalVal(decimalValue("1.5"), 1)}, // May parse as 1.5
			desc:        "Should handle invalid exponent format gracefully",
		},
		{
			name:        "multiple decimal points",
			input:       "1.2.3",
			expectError: false,
			expected:    []core.Value{value.DecimalVal(decimalValue("1.2"), 1)}, // May parse as 1.2
			desc:        "Should handle multiple decimal points gracefully",
		},
		{
			name:        "number at end of input",
			input:       "value 42",
			expectError: false,
			expected:    []core.Value{value.WordVal("value"), value.IntVal(42)},
			checkResult: nil,
			desc:        "Should parse number at end of input",
		},
		{
			name:        "number at start of input",
			input:       "123 rest",
			expectError: false,
			expected:    []core.Value{value.IntVal(123), value.WordVal("rest")},
			checkResult: nil,
			desc:        "Should parse number at start of input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.desc)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.desc, err)
				return
			}

			if len(vals) != len(tt.expected) {
				t.Errorf("Expected %d values, got %d for %s", len(tt.expected), len(vals), tt.desc)
				return
			}

			for i, expected := range tt.expected {
				// For decimals, compare string representations since Equals might not work for all cases
				if expected.GetType() == value.TypeDecimal && vals[i].GetType() == value.TypeDecimal {
					expectedDec, _ := value.AsDecimal(expected)
					actualDec, _ := value.AsDecimal(vals[i])
					if expectedDec.String() != actualDec.String() {
						t.Errorf("Decimal value %d mismatch for %s: expected %s, got %s", i, tt.desc, expectedDec.String(), actualDec.String())
					}
				} else if !vals[i].Equals(expected) {
					t.Errorf("Value %d mismatch for %s: expected %v, got %v", i, tt.desc, expected, vals[i])
				}
			}
		})
	}
}

// Helper function to create decimal values for testing
func decimalValue(s string) *decimal.Big {
	d := new(decimal.Big)
	d.SetString(s)
	return d
}

// TestAdvancedPathTokenization tests various advanced path tokenization scenarios
func TestAdvancedPathTokenization(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "path starting with number",
			input:       "1.field",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypePath {
					t.Errorf("Expected path, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse paths starting with numbers",
		},
		{
			name:        "decimal path",
			input:       "42.0.field",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypePath {
					t.Errorf("Expected path, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse paths with decimal numbers",
		},
		{
			name:        "complex nested path",
			input:       "user.address.city.name",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				path, ok := value.AsPath(vals[0])
				if !ok {
					t.Errorf("Expected path value")
					return
				}
				if len(path.Segments) != 4 {
					t.Errorf("Expected 4 segments, got %d", len(path.Segments))
				}
			},
			desc: "Should parse deeply nested paths",
		},
		{
			name:        "path with special characters",
			input:       "obj.field_name",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypePath {
					t.Errorf("Expected path, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should handle underscores in path segments",
		},
		{
			name:        "set-path with number",
			input:       "1.field:",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeSetWord {
					t.Errorf("Expected set-word, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse set-paths starting with numbers",
		},
		{
			name:        "regular path vs set-path",
			input:       "obj.field obj.field:",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Errorf("Expected 2 values, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypePath {
					t.Errorf("First value should be path, got %s", value.TypeToString(vals[0].GetType()))
				}
				if vals[1].GetType() != value.TypeSetWord {
					t.Errorf("Second value should be set-word, got %s", value.TypeToString(vals[1].GetType()))
				}
			},
			desc: "Should distinguish between paths and set-paths",
		},
		{
			name:        "path with numbers and words",
			input:       "data.0.name.1",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				path, ok := value.AsPath(vals[0])
				if !ok {
					t.Errorf("Expected path value")
					return
				}
				if len(path.Segments) != 4 {
					t.Errorf("Expected 4 segments, got %d", len(path.Segments))
				}
				// Check segment types
				expectedTypes := []value.PathSegmentType{
					value.PathSegmentWord,
					value.PathSegmentIndex,
					value.PathSegmentWord,
					value.PathSegmentIndex,
				}
				for i, expectedType := range expectedTypes {
					if path.Segments[i].Type != expectedType {
						t.Errorf("Segment %d: expected %v, got %v", i, expectedType, path.Segments[i].Type)
					}
				}
			},
			desc: "Should handle mixed number and word segments",
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

// TestWordVariants tests various word types and special keywords
func TestWordVariants(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "refinement word",
			input:       "--flag",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeWord {
					t.Errorf("Expected word, got %s", value.TypeToString(vals[0].GetType()))
				}
				word, ok := value.AsWord(vals[0])
				if !ok || word != "--flag" {
					t.Errorf("Expected '--flag', got %q", word)
				}
			},
			desc: "Should parse refinement words starting with --",
		},
		{
			name:        "refinement with value",
			input:       "--option value",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Errorf("Expected 2 values, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeWord {
					t.Errorf("Expected word, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse refinement words with values",
		},
		{
			name:        "get-word",
			input:       ":variable",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeGetWord {
					t.Errorf("Expected get-word, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse get-words starting with :",
		},
		{
			name:        "lit-word",
			input:       "'literal",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeLitWord {
					t.Errorf("Expected lit-word, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse lit-words starting with '",
		},
		{
			name:        "datatype literal",
			input:       "integer!",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeDatatype {
					t.Errorf("Expected datatype, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse datatype literals ending with !",
		},
		{
			name:        "true keyword",
			input:       "true",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeLogic {
					t.Errorf("Expected logic, got %s", value.TypeToString(vals[0].GetType()))
				}
				logic, ok := value.AsLogic(vals[0])
				if !ok || !logic {
					t.Errorf("Expected true, got %v", logic)
				}
			},
			desc: "Should parse true as logic value",
		},
		{
			name:        "false keyword",
			input:       "false",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeLogic {
					t.Errorf("Expected logic, got %s", value.TypeToString(vals[0].GetType()))
				}
				logic, ok := value.AsLogic(vals[0])
				if !ok || logic {
					t.Errorf("Expected false, got %v", logic)
				}
			},
			desc: "Should parse false as logic value",
		},
		{
			name:        "none keyword",
			input:       "none",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeNone {
					t.Errorf("Expected none, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse none as none value",
		},
		{
			name:        "set-word",
			input:       "variable:",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeSetWord {
					t.Errorf("Expected set-word, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse set-words ending with :",
		},
		{
			name:        "regular word",
			input:       "variable",
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
			desc: "Should parse regular words",
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

// TestOperatorTokenization tests various operator tokenization scenarios
func TestOperatorTokenization(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "less than or equal",
			input:       "<= 5",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Errorf("Expected 2 values, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeWord {
					t.Errorf("Expected word, got %s", value.TypeToString(vals[0].GetType()))
				}
				word, ok := value.AsWord(vals[0])
				if !ok || word != "<=" {
					t.Errorf("Expected '<=', got %q", word)
				}
			},
			desc: "Should parse <= operator",
		},
		{
			name:        "greater than or equal",
			input:       ">= 10",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Errorf("Expected 2 values, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeWord {
					t.Errorf("Expected word, got %s", value.TypeToString(vals[0].GetType()))
				}
				word, ok := value.AsWord(vals[0])
				if !ok || word != ">=" {
					t.Errorf("Expected '>=', got %q", word)
				}
			},
			desc: "Should parse >= operator",
		},
		{
			name:        "not equal",
			input:       "<> 0",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Errorf("Expected 2 values, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeWord {
					t.Errorf("Expected word, got %s", value.TypeToString(vals[0].GetType()))
				}
				word, ok := value.AsWord(vals[0])
				if !ok || word != "<>" {
					t.Errorf("Expected '<>', got %q", word)
				}
			},
			desc: "Should parse <> operator",
		},
		{
			name:        "single character operators",
			input:       "+ - * / < > =",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				// This will be parsed as an expression, not separate tokens
				if len(vals) != 1 {
					t.Errorf("Expected 1 value (expression), got %d", len(vals))
					return
				}
				// The result should be a complex expression
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren (expression), got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse operators as expression",
		},
		{
			name:        "operators in expression",
			input:       "3 + 4 * 2",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value (parsed expression), got %d", len(vals))
					return
				}
				// The expression should be parsed as ((+ 3 4) * 2)
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren (expression), got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse operators in mathematical expressions",
		},
		{
			name:        "comparison operators",
			input:       "x < y and y > z",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value (parsed expression), got %d", len(vals))
					return
				}
			},
			desc: "Should parse comparison and logical operators",
		},
		{
			name:        "operator precedence test",
			input:       "1 + 2 * 3",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				// Should parse as (+ 1 (* 2 3)) due to left-to-right evaluation
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should demonstrate left-to-right operator evaluation",
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
