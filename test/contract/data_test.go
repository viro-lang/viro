// Package contract validates data natives per contracts/data.md
package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
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
		expected core.Value
		check    string
		wantErr  bool
	}{
		{
			name:     "set integer value",
			input:    "set 'x 10\nx",
			expected: value.NewIntVal(10),
			check:    "x",
			wantErr:  false,
		},
		{
			name:     "set expression result",
			input:    "set 'y (+ 3 4)\ny",
			expected: value.NewIntVal(7),
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
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "get returns bound value",
			input:    "x: 10\nget 'x",
			expected: value.NewIntVal(10),
			wantErr:  false,
		},
		{
			name:     "get after set",
			input:    "set 'name \"Alice\"\nget 'name",
			expected: value.NewStrVal("Alice"),
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
		expected core.Value
	}{
		{
			name:     "integer type",
			input:    "type? 42",
			expected: value.NewWordVal("integer!"),
		},
		{
			name:     "string type",
			input:    "type? \"hello\"",
			expected: value.NewWordVal("string!"),
		},
		{
			name:     "logic true type",
			input:    "type? true",
			expected: value.NewWordVal("logic!"),
		},
		{
			name:     "logic false type",
			input:    "type? false",
			expected: value.NewWordVal("logic!"),
		},
		{
			name:     "none type",
			input:    "type? none",
			expected: value.NewWordVal("none!"),
		},
		{
			name:     "block type",
			input:    "type? []",
			expected: value.NewWordVal("block!"),
		},
		{
			name:     "word type",
			input:    "type? 'x",
			expected: value.NewWordVal("word!"),
		},
		{
			name:     "set-word type",
			input:    "type? (first [x:])",
			expected: value.NewWordVal("set-word!"),
		},
		{
			name:     "get-word type",
			input:    "type? (first [:x])",
			expected: value.NewWordVal("get-word!"),
		},
		{
			name:     "lit-word type",
			input:    "type? (first ['x])",
			expected: value.NewWordVal("lit-word!"),
		},
		{
			name:     "get-word fetches value",
			input:    "x: 20\ntype? :x",
			expected: value.NewWordVal("integer!"),
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
		expected core.Value
	}{
		{
			name:     "form block removes brackets",
			input:    "form [1 2 3]",
			expected: value.NewStrVal("1 2 3"),
		},
		{
			name:     "form string removes quotes",
			input:    "form \"hello\"",
			expected: value.NewStrVal("hello"),
		},
		{
			name:     "form integer",
			input:    "form 42",
			expected: value.NewStrVal("42"),
		},
		{
			name:     "form logic true",
			input:    "form true",
			expected: value.NewStrVal("true"),
		},
		{
			name:     "form logic false",
			input:    "form false",
			expected: value.NewStrVal("false"),
		},
		{
			name:     "form none",
			input:    "form none",
			expected: value.NewStrVal("none"),
		},
		{
			name:     "form word",
			input:    "form 'x",
			expected: value.NewStrVal("x"),
		},
		{
			name:     "form empty block",
			input:    "form []",
			expected: value.NewStrVal(""),
		},
		{
			name:     "form nested block",
			input:    "form [a [b c] d]",
			expected: value.NewStrVal("a b c d"),
		},
		{
			name:     "form empty object",
			input:    "form (make object! [])",
			expected: value.NewStrVal(""),
		},
		{
			name:     "form object with fields",
			input:    "form (make object! [x: 10 y: \"hello\"])",
			expected: value.NewStrVal("x: 10\ny: hello"),
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
// Contract: mold value → string! code-readable representation
// - Blocks: with outer brackets
// - Strings: with quotes
// - Other types: standard representation
func TestData_Mold(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
	}{
		{
			name:     "mold block includes brackets",
			input:    "mold [1 2 3]",
			expected: value.NewStrVal("[1 2 3]"),
		},
		{
			name:     "mold string includes quotes",
			input:    "mold \"hello\"",
			expected: value.NewStrVal("\"hello\""),
		},
		{
			name:     "mold integer",
			input:    "mold 42",
			expected: value.NewStrVal("42"),
		},
		{
			name:     "mold logic true",
			input:    "mold true",
			expected: value.NewStrVal("true"),
		},
		{
			name:     "mold logic false",
			input:    "mold false",
			expected: value.NewStrVal("false"),
		},
		{
			name:     "mold none",
			input:    "mold none",
			expected: value.NewStrVal("none"),
		},
		{
			name:     "mold word",
			input:    "mold 'x",
			expected: value.NewStrVal("x"),
		},
		{
			name:     "mold empty block",
			input:    "mold []",
			expected: value.NewStrVal("[]"),
		},
		{
			name:     "mold nested block",
			input:    "mold [a [b c] d]",
			expected: value.NewStrVal("[a [b c] d]"),
		},
		{
			name:     "mold empty object",
			input:    "mold (make object! [])",
			expected: value.NewStrVal("make object! []"),
		},
		{
			name:     "mold object with fields",
			input:    "mold (make object! [x: 10 y: \"hello\"])",
			expected: value.NewStrVal("make object! [x: 10 y: \"hello\"]"),
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

// TestData_Reduce validates the 'reduce' native.
//
// Contract: reduce block → block! containing evaluated elements
// - Takes a block (not evaluated)
// - Evaluates each element individually
// - Returns new block with evaluation results
// - Preserves element order
func TestData_Reduce(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "reduce literals",
			input:    "reduce [1 2 3]",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
			wantErr:  false,
		},
		{
			name:     "reduce expressions",
			input:    "reduce [1 + 2 3 * 4]",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(3), value.NewIntVal(12)}),
			wantErr:  false,
		},
		{
			name:     "reduce with variables",
			input:    "x: 10\ny: 20\nreduce [x y + 5]",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(10), value.NewIntVal(25)}),
			wantErr:  false,
		},
		{
			name:     "reduce empty block",
			input:    "reduce []",
			expected: value.NewBlockVal([]core.Value{}),
			wantErr:  false,
		},
		{
			name:     "reduce mixed types",
			input:    "reduce [42 \"hello\" true none]",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(42), value.NewStrVal("hello"), value.NewLogicVal(true), value.NewNoneVal()}),
			wantErr:  false,
		},
		{
			name:  "reduce nested blocks",
			input: "reduce [[1 2] [3 4]]",
			expected: value.NewBlockVal([]core.Value{
				value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}),
				value.NewBlockVal([]core.Value{value.NewIntVal(3), value.NewIntVal(4)}),
			}),
			wantErr: false,
		},
		{
			name:     "reduce with function calls",
			input:    "reduce [(type? 42) (type? \"test\")]",
			expected: value.NewBlockVal([]core.Value{value.NewWordVal("integer!"), value.NewWordVal("string!")}),
			wantErr:  false,
		},
		{
			name:     "reduce non-block argument",
			input:    "reduce 42",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:    "reduce undefined variable",
			input:   "reduce [undefined-var]",
			wantErr: true,
		},
		{
			name:    "reduce with evaluation error",
			input:   "reduce [1 / 0]",
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
