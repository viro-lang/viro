package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestToInteger_FromInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
	}{
		{
			name:     "positive integer",
			input:    "to-integer 42",
			expected: value.NewIntVal(42),
		},
		{
			name:     "negative integer",
			input:    "to-integer -15",
			expected: value.NewIntVal(-15),
		},
		{
			name:     "zero",
			input:    "to-integer 0",
			expected: value.NewIntVal(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Equals(tt.expected) {
				t.Fatalf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToInteger_FromDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
	}{
		{
			name:     "positive decimal truncated",
			input:    "to-integer 3.7",
			expected: value.NewIntVal(3),
		},
		{
			name:     "negative decimal truncated",
			input:    "to-integer -2.9",
			expected: value.NewIntVal(-2),
		},
		{
			name:     "decimal zero",
			input:    "to-integer 0.0",
			expected: value.NewIntVal(0),
		},
		{
			name:     "large decimal",
			input:    "to-integer 999.999",
			expected: value.NewIntVal(999),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Equals(tt.expected) {
				t.Fatalf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToInteger_FromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "positive string",
			input:    `to-integer "123"`,
			expected: value.NewIntVal(123),
			wantErr:  false,
		},
		{
			name:     "negative string",
			input:    `to-integer "-456"`,
			expected: value.NewIntVal(-456),
			wantErr:  false,
		},
		{
			name:     "zero string",
			input:    `to-integer "0"`,
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:    "invalid string",
			input:   `to-integer "abc"`,
			wantErr: true,
		},
		{
			name:    "decimal string",
			input:   `to-integer "12.34"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !result.Equals(tt.expected) {
				t.Fatalf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToDecimal_FromInteger(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "positive integer",
			input: "to-decimal 42",
		},
		{
			name:  "negative integer",
			input: "to-decimal -15",
		},
		{
			name:  "zero",
			input: "to-decimal 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.GetType() != value.TypeDecimal {
				t.Fatalf("Expected decimal type, got %v", result.GetType())
			}
		})
	}
}

func TestToDecimal_FromDecimal(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "positive decimal",
			input: "to-decimal 3.7",
		},
		{
			name:  "negative decimal",
			input: "to-decimal -2.9",
		},
		{
			name:  "decimal zero",
			input: "to-decimal 0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.GetType() != value.TypeDecimal {
				t.Fatalf("Expected decimal type, got %v", result.GetType())
			}
		})
	}
}

func TestToDecimal_FromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid decimal string",
			input:   `to-decimal "12.34"`,
			wantErr: false,
		},
		{
			name:    "integer string",
			input:   `to-decimal "42"`,
			wantErr: false,
		},
		{
			name:    "negative decimal string",
			input:   `to-decimal "-3.14"`,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   `to-decimal "abc"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.GetType() != value.TypeDecimal {
				t.Fatalf("Expected decimal type, got %v", result.GetType())
			}
		})
	}
}

func TestToString_FromVariousTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "integer to string",
			input:    "to-string 42",
			expected: "42",
		},
		{
			name:     "negative integer to string",
			input:    "to-string -15",
			expected: "-15",
		},
		{
			name:     "decimal to string",
			input:    "to-string 3.70",
			expected: "3.70",
		},
		{
			name:     "string to string",
			input:    `to-string "hello"`,
			expected: "hello",
		},
		{
			name:     "block to string",
			input:    "to-string [1 2 3]",
			expected: "1 2 3",
		},
		{
			name:     "word to string",
			input:    "to-string 'test",
			expected: "test",
		},
		{
			name:     "none to string",
			input:    "to-string none",
			expected: "none",
		},
		{
			name:     "object to string",
			input:    "to-string make object! [name: \"test\"]",
			expected: "name: test",
		},
		{
			name:     "empty binary to string",
			input:    "to-string #{}",
			expected: "",
		},
		{
			name:     "single byte binary to string",
			input:    "to-string #{41}",
			expected: "41",
		},
		{
			name:     "multiple bytes binary to string",
			input:    "to-string #{41 42 43}",
			expected: "414243",
		},
		{
			name:     "uppercase hex in binary to string",
			input:    "to-string #{DE AD BE EF}",
			expected: "DEADBEEF",
		},
		{
			name:     "mixed case hex in binary to string",
			input:    "to-string #{01 2A 3F}",
			expected: "012A3F",
		},
		{
			name:     "zero bytes in binary to string",
			input:    "to-string #{00 00}",
			expected: "0000",
		},
		{
			name:     "large binary to string",
			input:    "to-string #{41 42 43 44 45 46 47 48 49 4A 4B 4C 4D 4E 4F 50 51 52 53 54 55 56 57 58 59 5A}",
			expected: "4142434445464748494A4B4C4D4E4F505152535455565758595A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.GetType() != value.TypeString {
				t.Fatalf("Expected string type, got %v", result.GetType())
			}
			strVal, _ := value.AsStringValue(result)
			if strVal.String() != tt.expected {
				t.Fatalf("Expected %q, got %q", tt.expected, strVal.String())
			}
		})
	}
}

func TestConversionChain(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType core.ValueType
	}{
		{
			name:         "string to integer to decimal",
			input:        `to-decimal (to-integer "42")`,
			expectedType: value.TypeDecimal,
		},
		{
			name:         "integer to string to integer",
			input:        `to-integer (to-string 123)`,
			expectedType: value.TypeInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.GetType() != tt.expectedType {
				t.Fatalf("Expected type %v, got %v", tt.expectedType, result.GetType())
			}
		})
	}
}

func TestConversionErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "to-integer with invalid type",
			input: "to-integer [1 2 3]",
		},
		{
			name:  "to-decimal with invalid type",
			input: "to-decimal [1 2 3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.input)
			if err == nil {
				t.Fatalf("Expected error but got none")
			}
		})
	}
}
