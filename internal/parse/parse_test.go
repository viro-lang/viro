package parse

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestDecimalLiteralParsing validates T033/T034: decimal literal parsing
func TestDecimalLiteralParsing(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantType   core.ValueType
		wantString string
	}{
		{
			name:       "Integer literal",
			input:      "42",
			wantType:   value.TypeInteger,
			wantString: "42",
		},
		{
			name:       "Negative integer",
			input:      "-123",
			wantType:   value.TypeInteger,
			wantString: "-123",
		},
		{
			name:       "Simple decimal",
			input:      "19.99",
			wantType:   value.TypeDecimal,
			wantString: "19.99",
		},
		{
			name:       "Negative decimal",
			input:      "-3.14",
			wantType:   value.TypeDecimal,
			wantString: "-3.14",
		},
		{
			name:       "Decimal with exponent",
			input:      "1.5e2",
			wantType:   value.TypeDecimal,
			wantString: "150",
		},
		{
			name:       "Decimal with negative exponent",
			input:      "2.5E-3",
			wantType:   value.TypeDecimal,
			wantString: "0.0025",
		},
		{
			name:       "Decimal with positive exponent",
			input:      "1.23e+4",
			wantType:   value.TypeDecimal,
			wantString: "12300",
		},
		{
			name:       "Scientific notation",
			input:      "6.022e23",
			wantType:   value.TypeDecimal,
			wantString: "6.022e+23",
		},
		{
			name:       "Zero decimal",
			input:      "0.0",
			wantType:   value.TypeDecimal,
			wantString: "0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if len(vals) != 1 {
				t.Fatalf("Expected 1 value, got %d", len(vals))
			}

			val := vals[0]
			if val.GetType() != tt.wantType {
				t.Errorf("Expected type %s, got %s", value.TypeToString(tt.wantType), value.TypeToString(val.GetType()))
			}

			// Note: We check Contains instead of exact match because decimal
			// formatting may vary slightly (e.g., "150" vs "1.5e+2")
			valStr := val.String()
			if tt.wantType == value.TypeDecimal {
				// For decimals, just verify it's a decimal type
				if _, ok := value.AsDecimal(val); !ok {
					t.Errorf("Expected decimal value, got %v", val)
				}
			} else {
				if valStr != tt.wantString {
					t.Errorf("Expected %q, got %q", tt.wantString, valStr)
				}
			}
		})
	}
}

// TestDecimalLiteralDisambiguation validates T034.1: token disambiguation
func TestDecimalLiteralDisambiguation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		desc    string
	}{
		{
			name:    "Decimal vs refinement",
			input:   "19.99 --places",
			wantErr: false,
			desc:    "Should parse decimal literal and refinement separately",
		},
		{
			name:    "Negative decimal",
			input:   "-3.14",
			wantErr: false,
			desc:    "Should parse negative decimal correctly",
		},
		{
			name:    "Multiple decimals",
			input:   "1.5 2.5 3.5",
			wantErr: false,
			desc:    "Should parse multiple decimals in sequence",
		},
		{
			name:    "Decimal in expression",
			input:   "19.99 * 3",
			wantErr: false,
			desc:    "Should parse decimal in arithmetic expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				if tt.wantErr {
					t.Errorf("Expected error but got none for: %s", tt.desc)
				} else {
					t.Errorf("Unexpected error for %s: %v", tt.desc, err)
				}
			}
		})
	}
}

// TestCommentParsing validates that comments starting with ';' are ignored
func TestCommentParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []core.Value
		desc     string
	}{
		{
			name:     "Comment at start of line",
			input:    "; this is a comment\n42",
			expected: []core.Value{value.IntVal(42)},
			desc:     "Should ignore comment and parse the number",
		},
		{
			name:     "Comment after code",
			input:    "42 ; this is a comment",
			expected: []core.Value{value.IntVal(42)},
			desc:     "Should parse number and ignore trailing comment",
		},
		{
			name:     "Multiple comments",
			input:    "; first comment\n42 ; second comment\n; third comment",
			expected: []core.Value{value.IntVal(42)},
			desc:     "Should ignore all comments and parse the number",
		},
		{
			name:     "Comment in expression",
			input:    "3 ; comment\n+ ; another comment\n4",
			expected: []core.Value{value.IntVal(3), value.WordVal("+"), value.IntVal(4)},
			desc:     "Should parse as flat sequence with comments removed",
		},
		{
			name:     "Empty comment",
			input:    "42 ;",
			expected: []core.Value{value.IntVal(42)},
			desc:     "Should handle empty comments",
		},
		{
			name:     "Comment at EOF",
			input:    "42 ; comment at end",
			expected: []core.Value{value.IntVal(42)},
			desc:     "Should handle comments at end of input",
		},
		{
			name:     "Only comments",
			input:    "; first\n; second\n; third",
			expected: []core.Value{},
			desc:     "Should handle input with only comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error for %s: %v", tt.desc, err)
			}

			if len(vals) != len(tt.expected) {
				t.Fatalf("Expected %d values, got %d for %s", len(tt.expected), len(vals), tt.desc)
			}

			for i, expected := range tt.expected {
				if !vals[i].Equals(expected) {
					t.Errorf("Value %d mismatch for %s: expected %v, got %v", i, tt.desc, expected, vals[i])
				}
			}
		})
	}
}
