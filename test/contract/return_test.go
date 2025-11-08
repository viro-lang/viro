// Package contract validates return native per contracts/control-flow.md
package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestReturn_Basic validates basic return functionality.
//
// Contract: return [value]
// - With value: returns the value from function
// - Without value: returns none from function
// - Works with all value types (int, string, block, object, etc.)
func TestReturn_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return integer value",
			input:    "fn: function [x] [return x + 10]\nfn 5",
			expected: value.NewIntVal(15),
			wantErr:  false,
		},
		{
			name:     "return no value returns none",
			input:    "fn: function [] [return]\nfn",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "return string value",
			input:    "fn: function [] [return \"hello\"]\nfn",
			expected: value.NewStrVal("hello"),
			wantErr:  false,
		},
		{
			name:     "return logic true value",
			input:    "fn: function [] [return true]\nfn",
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "return logic false value",
			input:    "fn: function [] [return false]\nfn",
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "return block value",
			input:    "fn: function [] [return [1 2 3]]\nfn",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
			wantErr:  false,
		},
		{
			name:     "return empty block",
			input:    "fn: function [] [return []]\nfn",
			expected: value.NewBlockVal([]core.Value{}),
			wantErr:  false,
		},
		{
			name:     "return on first line",
			input:    "fn: function [] [return 42\nprint \"never\"]\nfn",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "return on last line",
			input:    "fn: function [] [x: 10\nreturn x * 2]\nfn",
			expected: value.NewIntVal(20),
			wantErr:  false,
		},
		{
			name:     "return with expression evaluation",
			input:    "fn: function [a b] [return a + b * 2]\nfn 3 4",
			expected: value.NewIntVal(11),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestReturn_EarlyExit validates that return skips remaining code.
//
// Contract: return exits function immediately, skipping all remaining code.
func TestReturn_EarlyExit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return skips remaining code",
			input:    "fn: function [] [x: 0\nreturn 42\nx: 100]\nfn",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "return in conditional branch",
			input:    "fn: function [x] [when (> x 10) [return 1]\nreturn 0]\nfn 20",
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		{
			name:     "return in conditional branch - false case",
			input:    "fn: function [x] [when (> x 10) [return 1]\nreturn 0]\nfn 5",
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:     "return with counter to verify early exit",
			input:    "fn: function [] [count: 0\ncount: count + 1\nreturn count\ncount: count + 1]\nfn",
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		{
			name:     "return in nested conditional",
			input:    "fn: function [x] [if (> x 5) [if (> x 10) [return 2] return 1] return 0]\nfn 15",
			expected: value.NewIntVal(2),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestReturn_Nested validates return behavior in nested functions.
//
// Contract: return only exits the function it's called in, not outer functions.
func TestReturn_Nested(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "return in inner function only",
			input: `
				outer: function [] [
					inner: function [] [return 42]
					x: inner
					return x + 10
				]
				outer
			`,
			expected: value.NewIntVal(52),
			wantErr:  false,
		},
		{
			name: "outer function continues after inner return",
			input: `
				outer: function [] [
					inner: function [] [return 100]
					result1: inner
					result2: 200
					return result1 + result2
				]
				outer
			`,
			expected: value.NewIntVal(300),
			wantErr:  false,
		},
		{
			name: "multiple nested levels",
			input: `
				level1: function [] [
					level2: function [] [
						level3: function [] [return 1]
						x: level3
						return x + 2
					]
					y: level2
					return y + 3
				]
				level1
			`,
			expected: value.NewIntVal(6),
			wantErr:  false,
		},
		{
			name: "return in deeply nested function",
			input: `
				fn: function [] [
					a: function [] [
						b: function [] [
							c: function [] [return 42]
							x: c
							return x * 2
						]
						y: b
						return y + 10
					]
					z: a
					return z + 5
				]
				fn
			`,
			expected: value.NewIntVal(99),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestReturn_TransparentBlocks validates return through transparent blocks.
//
// Contract: return propagates through do, reduce, compose, when, unless blocks.
func TestReturn_TransparentBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return in do block exits function",
			input:    "fn: function [x] [do [return x]\nx + 100]\nfn 5",
			expected: value.NewIntVal(5),
			wantErr:  false,
		},
		{
			name:     "return in when exits function",
			input:    "fn: function [x] [when (> x 10) [return x]\nx + 100]\nfn 20",
			expected: value.NewIntVal(20),
			wantErr:  false,
		},
		{
			name:     "return in unless exits function",
			input:    "fn: function [x] [unless (> x 10) [return x]\nx + 100]\nfn 5",
			expected: value.NewIntVal(5),
			wantErr:  false,
		},
		{
			name:     "return in reduce exits function",
			input:    "fn: function [x] [reduce [return x]\nx + 100]\nfn 7",
			expected: value.NewIntVal(7),
			wantErr:  false,
		},
		{
			name:     "return in compose exits function",
			input:    "fn: function [x] [compose [(return x)]\nx + 100]\nfn 9",
			expected: value.NewIntVal(9),
			wantErr:  false,
		},
		{
			name:     "return in nested transparent blocks",
			input:    "fn: function [x] [do [when (> x 5) [do [return x]]]\nx + 100]\nfn 10",
			expected: value.NewIntVal(10),
			wantErr:  false,
		},
		{
			name:     "return in if true block exits function",
			input:    "fn: function [x] [if true [return x] [return x + 100]]\nfn 3",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "return in if false block exits function",
			input:    "fn: function [x] [if false [return x] [return x + 100]]\nfn 3",
			expected: value.NewIntVal(103),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestReturn_LoopInteraction validates return behavior in loops.
//
// Contract: return in loop exits function, not just the loop.
func TestReturn_LoopInteraction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return in loop exits function",
			input:    "fn: function [] [x: 0\nloop 3 [x: x + 1\nreturn 42]\nx + 100]\nfn",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "return in while loop exits function",
			input:    "fn: function [] [x: 0\nwhile (< x 10) [x: x + 1\nreturn 99]\nx + 100]\nfn",
			expected: value.NewIntVal(99),
			wantErr:  false,
		},
		{
			name:     "return in foreach exits function",
			input:    "fn: function [] [x: 0\nforeach [1 2 3] 'val [x: x + val\nreturn 77]\nx + 100]\nfn",
			expected: value.NewIntVal(77),
			wantErr:  false,
		},
		{
			name:     "break in loop, return after",
			input:    "fn: function [] [loop 3 [break]\nreturn 55]\nfn",
			expected: value.NewIntVal(55),
			wantErr:  false,
		},
		{
			name:     "return in loop with counter",
			input:    "fn: function [] [count: 0\nloop 5 [count: count + 1\nwhen (= count 3) [return count]]\ncount + 100]\nfn",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "return in nested loop exits function",
			input:    "fn: function [] [loop 3 [loop 3 [return 123]]\nprint \"never\"]\nfn",
			expected: value.NewIntVal(123),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestReturn_TopLevel validates return behavior at top level.
//
// Contract: return at top level exits script/REPL with the value.
func TestReturn_TopLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return at top level",
			input:    "return 42",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "return with no value at top level",
			input:    "x: 5\nreturn\nx: 10",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "return in top-level loop exits script",
			input:    "x: 0\nloop 3 [x: x + 1\nreturn x]\nx + 100",
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		{
			name:     "return in top-level conditional",
			input:    "x: 20\nwhen (> x 10) [return 99]\nx + 1",
			expected: value.NewIntVal(99),
			wantErr:  false,
		},
		{
			name:     "return in top-level conditional false case",
			input:    "x: 5\nwhen (> x 10) [return 99]\nreturn 77",
			expected: value.NewIntVal(77),
			wantErr:  false,
		},
		{
			name:     "return with string at top level",
			input:    "return \"hello world\"",
			expected: value.NewStrVal("hello world"),
			wantErr:  false,
		},
		{
			name:     "return with block at top level",
			input:    "return [1 2 3 4]",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3), value.NewIntVal(4)}),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestReturn_ValuePropagation validates complex value propagation.
//
// Contract: return can return any value type including function call results.
func TestReturn_ValuePropagation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return function call result",
			input:    "inner: function [] [42]\nouter: function [] [return inner]\nouter",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "return result of arithmetic",
			input:    "fn: function [a b c] [return a + b * c]\nfn 2 3 4",
			expected: value.NewIntVal(14),
			wantErr:  false,
		},
		{
			name:     "return result of series operation",
			input:    "fn: function [] [return first [10 20 30]]\nfn",
			expected: value.NewIntVal(10),
			wantErr:  false,
		},
		{
			name:     "return result of string operation",
			input:    "fn: function [] [return join \"hello\" \" world\"]\nfn",
			expected: value.NewStrVal("hello world"),
			wantErr:  false,
		},
		{
			name:     "return result of logic operation",
			input:    "fn: function [x y] [return and (> x 5) (< y 10)]\nfn 7 8",
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "return complex expression",
			input:    "fn: function [x] [return (+ (* x 2) 10)]\nfn 5",
			expected: value.NewIntVal(20),
			wantErr:  false,
		},
		{
			name:     "return nested function call",
			input:    "a: function [] [1]\nb: function [] [return a]\nc: function [] [return b]\nc",
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
