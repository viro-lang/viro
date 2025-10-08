// Package contract validates data natives per contracts/data.md
package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
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
			vals, err := parse.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			e := eval.NewEvaluator()
			result, evalErr := e.Do_Blk(vals)

			if tt.wantErr {
				if evalErr == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}

			if evalErr != nil {
				t.Fatalf("Unexpected error: %v", evalErr)
			}

			if !result.Equals(tt.expected) {
				t.Fatalf("Expected result %v, got %v", tt.expected, result)
			}

			if tt.check != "" {
				if len(e.Frames) == 0 {
					t.Fatalf("no frames available to verify binding")
				}
				bound, ok := e.Frames[len(e.Frames)-1].Get(tt.check)
				if !ok {
					t.Fatalf("expected word %s to be bound", tt.check)
				}
				if !bound.Equals(tt.expected) {
					t.Fatalf("binding for %s = %v, want %v", tt.check, bound, tt.expected)
				}
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
			vals, err := parse.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			e := eval.NewEvaluator()
			result, evalErr := e.Do_Blk(vals)

			if tt.wantErr {
				if evalErr == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}

			if evalErr != nil {
				t.Fatalf("Unexpected error: %v", evalErr)
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
			vals, err := parse.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			e := eval.NewEvaluator()
			result, evalErr := e.Do_Blk(vals)
			if evalErr != nil {
				t.Fatalf("Unexpected error: %v", evalErr)
			}

			if !result.Equals(tt.expected) {
				t.Fatalf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
