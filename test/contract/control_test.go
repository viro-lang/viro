// Package contract validates control flow natives per contracts/control-flow.md
package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestControlFlow_When validates the 'when' conditional native.
//
// Contract: when condition [block]
// - If condition is truthy: evaluate block, return result
// - If condition is falsy: return none without evaluating block
func TestControlFlow_When(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
		wantErr  bool
	}{
		{
			name:     "when true evaluates block",
			input:    "when true [42]",
			expected: value.IntVal(42),
			wantErr:  false,
		},
		{
			name:     "when false returns none",
			input:    "when false [99]",
			expected: value.NoneVal(),
			wantErr:  false,
		},
		{
			name:     "when none returns none",
			input:    "when none [42]",
			expected: value.NoneVal(),
			wantErr:  false,
		},
		{
			name:     "when integer (truthy) evaluates block",
			input:    "when 1 [99]",
			expected: value.IntVal(99),
			wantErr:  false,
		},
		{
			name:     "when with expression in block",
			input:    "when true [(+ 1 1)]",
			expected: value.IntVal(2),
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

// TestControlFlow_If validates the 'if' conditional native.
//
// Contract: if condition [true-block] [false-block]
// - Both blocks required
// - If condition is truthy: evaluate and return true-block
// - If condition is falsy: evaluate and return false-block
func TestControlFlow_If(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
		wantErr  bool
	}{
		{
			name:     "if true evaluates first block",
			input:    "if true [42] [99]",
			expected: value.IntVal(42),
			wantErr:  false,
		},
		{
			name:     "if false evaluates second block",
			input:    "if false [42] [99]",
			expected: value.IntVal(99),
			wantErr:  false,
		},
		{
			name:     "if with comparison",
			input:    "if (< 1 2) [10] [20]",
			expected: value.IntVal(10),
			wantErr:  false,
		},
		{
			name:     "if none is falsy",
			input:    "if none [10] [20]",
			expected: value.IntVal(20),
			wantErr:  false,
		},
		{
			name:     "if with expressions in blocks",
			input:    "if true [(+ 1 1)] [(+ 2 2)]",
			expected: value.IntVal(2),
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

// TestControlFlow_Loop validates the 'loop' iteration native.
//
// Contract: loop count [block]
// - count must be non-negative integer
// - Evaluates block count times
// - Returns result of last evaluation (or none if count=0)
func TestControlFlow_Loop(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
		wantErr  bool
	}{
		{
			name:     "loop 3 times",
			input:    "loop 3 [42]",
			expected: value.IntVal(42),
			wantErr:  false,
		},
		{
			name:     "loop 0 times returns none",
			input:    "loop 0 [42]",
			expected: value.NoneVal(),
			wantErr:  false,
		},
		{
			name:     "loop 1 time",
			input:    "loop 1 [99]",
			expected: value.IntVal(99),
			wantErr:  false,
		},
		{
			name:    "loop negative count fails",
			input:   "loop -1 [42]",
			wantErr: true,
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

// TestControlFlow_LoopWithCounter validates loop increments counter correctly
func TestControlFlow_LoopWithCounter(t *testing.T) {
	input := `count: 0
loop 5 [count: (+ count 1)]
count`

	result, err := Evaluate(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := value.IntVal(5)
	if !result.Equals(expected) {
		t.Errorf("Expected count=5, got %v", result)
	}
}

// TestControlFlow_While validates the 'while' iteration native.
//
// Contract: while [condition-block] [body-block]
// - Both blocks required
// - Re-evaluates condition-block before each iteration
// - Stops when condition-block evaluates to falsy
// - Returns result of last body evaluation (or none if never true)
func TestControlFlow_While(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
		wantErr  bool
	}{
		{
			name:     "while false returns none",
			input:    "while [false] [42]",
			expected: value.NoneVal(),
			wantErr:  false,
		},
		{
			name: "while with counter",
			input: `n: 0
while [(< n 3)] [n: (+ n 1)]`,
			expected: value.IntVal(3),
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

// TestControlFlow_TruthyConversion validates truthy conversion rules
func TestControlFlow_TruthyConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
	}{
		{
			name:     "none is falsy",
			input:    "when none [42]",
			expected: value.NoneVal(),
		},
		{
			name:     "false is falsy",
			input:    "when false [42]",
			expected: value.NoneVal(),
		},
		{
			name:     "true is truthy",
			input:    "when true [42]",
			expected: value.IntVal(42),
		},
		{
			name:     "zero is truthy",
			input:    "when 0 [42]",
			expected: value.IntVal(42),
		},
		{
			name:     "empty string is truthy",
			input:    `when "" [42]`,
			expected: value.IntVal(42),
		},
		{
			name:     "empty block is truthy",
			input:    "when [] [42]",
			expected: value.IntVal(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
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

// TestControlFlow_ComparisonOperators validates comparison operators used by control flow
func TestControlFlow_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected value.Value
	}{
		{
			name:     "less than true",
			input:    "< 3 5",
			expected: value.LogicVal(true),
		},
		{
			name:     "less than false",
			input:    "< 5 3",
			expected: value.LogicVal(false),
		},
		{
			name:     "greater than true",
			input:    "> 5 3",
			expected: value.LogicVal(true),
		},
		{
			name:     "less or equal true (equal)",
			input:    "<= 3 3",
			expected: value.LogicVal(true),
		},
		{
			name:     "greater or equal true",
			input:    ">= 5 3",
			expected: value.LogicVal(true),
		},
		{
			name:     "equal true",
			input:    "= 3 3",
			expected: value.LogicVal(true),
		},
		{
			name:     "not equal true",
			input:    "<> 3 5",
			expected: value.LogicVal(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
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
