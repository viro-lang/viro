// Package contract validates data natives per contracts/data.md
package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestData_Set validates the 'set' native.
//
// Contract: set 'word value
// - First argument must be word symbol (lit-word evaluates to word)
// - Second argument evaluated, bound to word in current frame
// - Returns the value that was set
func TestData_Set(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
		check    string
		wantErr  bool
	}{
		{
			name:     "set integer value",
			input:    "set 'x 10\nx",
			expected: value.IntVal(10),
			check:    "x",
			wantErr:  false,
		},
		{
			name:     "set expression result",
			input:    "set 'y (+ 3 4)\ny",
			expected: value.IntVal(7),
			check:    "y",
			wantErr:  false,
		},
		{
			name:    "set non-word errors",
			input:   "set 42 10",
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
				t.Fatalf("Expected result %v, got %v", tt.expected, result)
			}

		})
	}
}

// TestData_Get validates the 'get' native.
//
// Contract: get 'word
// - Returns value bound to word
// - Errors if argument not word or word unbound
func TestData_Get(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
		wantErr  bool
	}{
		{
			name:     "get returns bound value",
			input:    "x: 10\nget 'x",
			expected: value.IntVal(10),
			wantErr:  false,
		},
		{
			name:     "get after set",
			input:    "set 'name \"Alice\"\nget 'name",
			expected: value.StrVal("Alice"),
			wantErr:  false,
		},
		{
			name:    "get unbound word errors",
			input:   "get 'undefined",
			wantErr: true,
		},
		{
			name:    "get non-word errors",
			input:   "get 42",
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

// TestData_TypeQ validates the 'type?' native.
//
// Contract: type? value → word describing value type
func TestData_TypeQ(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
	}{
		{
			name:     "integer type",
			input:    "type? 42",
			expected: value.WordVal("integer!"),
		},
		{
			name:     "string type",
			input:    "type? \"hello\"",
			expected: value.WordVal("string!"),
		},
		{
			name:     "logic true type",
			input:    "type? true",
			expected: value.WordVal("logic!"),
		},
		{
			name:     "logic false type",
			input:    "type? false",
			expected: value.WordVal("logic!"),
		},
		{
			name:     "none type",
			input:    "type? none",
			expected: value.WordVal("none!"),
		},
		{
			name:     "block type",
			input:    "type? []",
			expected: value.WordVal("block!"),
		},
		{
			name:     "word type",
			input:    "type? 'x",
			expected: value.WordVal("word!"),
		},
		{
			name:     "set-word type",
			input:    "type? (first [x:])",
			expected: value.WordVal("set-word!"),
		},
		{
			name:     "get-word type",
			input:    "type? (first [:x])",
			expected: value.WordVal("get-word!"),
		},
		{
			name:     "lit-word type",
			input:    "type? (first ['x])",
			expected: value.WordVal("lit-word!"),
		},
		{
			name:     "get-word fetches value",
			input:    "x: 20\ntype? :x",
			expected: value.WordVal("integer!"),
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

// TestData_Form validates the 'form' native.
//
// Contract: form value → string! human-readable representation
// - Blocks: no outer brackets
// - Strings: no quotes
// - Other types: standard representation
func TestData_Form(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
	}{
		{
			name:     "form block removes brackets",
			input:    "form [1 2 3]",
			expected: value.StrVal("1 2 3"),
		},
		{
			name:     "form string removes quotes",
			input:    "form \"hello\"",
			expected: value.StrVal("hello"),
		},
		{
			name:     "form integer",
			input:    "form 42",
			expected: value.StrVal("42"),
		},
		{
			name:     "form logic true",
			input:    "form true",
			expected: value.StrVal("true"),
		},
		{
			name:     "form logic false",
			input:    "form false",
			expected: value.StrVal("false"),
		},
		{
			name:     "form none",
			input:    "form none",
			expected: value.StrVal("none"),
		},
		{
			name:     "form word",
			input:    "form 'x",
			expected: value.StrVal("x"),
		},
		{
			name:     "form empty block",
			input:    "form []",
			expected: value.StrVal(""),
		},
		{
			name:     "form nested block",
			input:    "form [a [b c] d]",
			expected: value.StrVal("a [b c] d"),
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

// TestData_Mold validates the 'mold' native.
//
// Contract: mold value → string! REBOL-readable representation
// - Blocks: with outer brackets
// - Strings: with quotes
// - Other types: standard representation
func TestData_Mold(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
	}{
		{
			name:     "mold block includes brackets",
			input:    "mold [1 2 3]",
			expected: value.StrVal("[1 2 3]"),
		},
		{
			name:     "mold string includes quotes",
			input:    "mold \"hello\"",
			expected: value.StrVal("\"hello\""),
		},
		{
			name:     "mold integer",
			input:    "mold 42",
			expected: value.StrVal("42"),
		},
		{
			name:     "mold logic true",
			input:    "mold true",
			expected: value.StrVal("true"),
		},
		{
			name:     "mold logic false",
			input:    "mold false",
			expected: value.StrVal("false"),
		},
		{
			name:     "mold none",
			input:    "mold none",
			expected: value.StrVal("none"),
		},
		{
			name:     "mold word",
			input:    "mold 'x",
			expected: value.StrVal("x"),
		},
		{
			name:     "mold empty block",
			input:    "mold []",
			expected: value.StrVal("[]"),
		},
		{
			name:     "mold nested block",
			input:    "mold [a [b c] d]",
			expected: value.StrVal("[a [b c] d]"),
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
