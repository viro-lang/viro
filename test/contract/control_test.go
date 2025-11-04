// Package contract validates control flow natives per contracts/control-flow.md
package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
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
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "when true evaluates block",
			input:    "when true [42]",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "when false returns none",
			input:    "when false [99]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "when none returns none",
			input:    "when none [42]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "when integer (truthy) evaluates block",
			input:    "when 1 [99]",
			expected: value.NewIntVal(99),
			wantErr:  false,
		},
		{
			name:     "when with expression in block",
			input:    "when true [(+ 1 1)]",
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
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "if true evaluates first block",
			input:    "if true [42] [99]",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "if false evaluates second block",
			input:    "if false [42] [99]",
			expected: value.NewIntVal(99),
			wantErr:  false,
		},
		{
			name:     "if with comparison",
			input:    "if (< 1 2) [10] [20]",
			expected: value.NewIntVal(10),
			wantErr:  false,
		},
		{
			name:     "if none is falsy",
			input:    "if none [10] [20]",
			expected: value.NewIntVal(20),
			wantErr:  false,
		},
		{
			name:     "if with expressions in blocks",
			input:    "if true [(+ 1 1)] [(+ 2 2)]",
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
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "loop 3 times",
			input:    "loop 3 [42]",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "loop 0 times returns none",
			input:    "loop 0 [42]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "loop 1 time",
			input:    "loop 1 [99]",
			expected: value.NewIntVal(99),
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

// TestControlFlow_LoopWithIndex validates the 'loop --with-index' refinement.
func TestControlFlow_LoopWithIndex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "loop --with-index collects indices",
			input:    "result: []\nloop 3 --with-index 'i [\n  result: (append result i)\n]\nresult",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(0), value.NewIntVal(1), value.NewIntVal(2)}),
			wantErr:  false,
		},
		{
			name:     "loop --with-index single iteration",
			input:    `loop 1 --with-index 'idx [idx]`,
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:     "loop --with-index zero iterations",
			input:    `loop 0 --with-index 'i [i]`,
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "loop --with-index accumulates sum of indices",
			input:    "sum: 0\nloop 5 --with-index 'i [sum: (+ sum i)]\nsum",
			expected: value.NewIntVal(10), // 0+1+2+3+4 = 10
			wantErr:  false,
		},
		{
			name: "loop --with-index overwrites existing binding",
			input: `i: 99
loop 3 --with-index 'i [i]
`,
			expected: value.NewIntVal(2), // Last value of i in loop iterations: 0, 1, 2
			wantErr:  false,
		},
		{
			name: "loop --with-index with complex expression",
			input: `result: []
loop 3 --with-index 'pos [
  result: (append result (* pos 10))
]
result`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(0), value.NewIntVal(10), value.NewIntVal(20)}),
			wantErr:  false,
		},
		{
			name:    "loop --with-index non-word value fails",
			input:   `loop 3 --with-index 42 [42]`,
			wantErr: true,
		},
		{
			name:    "loop --with-index with string value fails",
			input:   `loop 3 --with-index "i" [42]`,
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

	expected := value.NewIntVal(5)
	if !result.Equals(expected) {
		t.Errorf("Expected count=5, got %v", result)
	}
}

// TestControlFlow_While validates the 'while' iteration native.
//
// Contract: while condition [body-block]
// - Condition can be any value or block
// - If condition is a block, it is re-evaluated before each iteration
// - If condition is not a block, it is evaluated once and constant
// - Body must be a block
// - Stops when condition evaluates to falsy
// - Returns result of last body evaluation (or none if never true)
func TestControlFlow_While(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "while with block false returns none",
			input:    "while [false] [42]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name: "while with block counter",
			input: `n: 0
while [(< n 3)] [n: (+ n 1)]`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "while with literal false (constant)",
			input:    "while false [42]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "while with none (constant, falsy)",
			input:    "while none [42]",
			expected: value.NewNoneVal(),
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
		expected core.Value
	}{
		{
			name:     "none is falsy",
			input:    "when none [42]",
			expected: value.NewNoneVal(),
		},
		{
			name:     "false is falsy",
			input:    "when false [42]",
			expected: value.NewNoneVal(),
		},
		{
			name:     "true is truthy",
			input:    "when true [42]",
			expected: value.NewIntVal(42),
		},
		{
			name:     "zero is truthy",
			input:    "when 0 [42]",
			expected: value.NewIntVal(42),
		},
		{
			name:     "empty string is truthy",
			input:    `when "" [42]`,
			expected: value.NewIntVal(42),
		},
		{
			name:     "empty block is truthy",
			input:    "when [] [42]",
			expected: value.NewIntVal(42),
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
		expected core.Value
	}{
		{
			name:     "less than true",
			input:    "< 3 5",
			expected: value.NewLogicVal(true),
		},
		{
			name:     "less than false",
			input:    "< 5 3",
			expected: value.NewLogicVal(false),
		},
		{
			name:     "greater than true",
			input:    "> 5 3",
			expected: value.NewLogicVal(true),
		},
		{
			name:     "less or equal true (equal)",
			input:    "<= 3 3",
			expected: value.NewLogicVal(true),
		},
		{
			name:     "greater or equal true",
			input:    ">= 5 3",
			expected: value.NewLogicVal(true),
		},
		{
			name:     "equal true",
			input:    "= 3 3",
			expected: value.NewLogicVal(true),
		},
		{
			name:     "not equal true",
			input:    "<> 3 5",
			expected: value.NewLogicVal(true),
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

// TestControlFlow_ForeachWithIndex validates the 'foreach --with-index' refinement.
func TestControlFlow_ForeachWithIndex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "foreach --with-index single variable collects indices",
			input: `result: []
foreach [10 20 30] --with-index 'pos [n] [result: (append result pos)]
result`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(0), value.NewIntVal(1), value.NewIntVal(2)}),
			wantErr:  false,
		},
		{
			name:     "foreach --with-index with single element",
			input:    `foreach [42] --with-index 'idx [n] [idx]`,
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:     "foreach --with-index empty series returns none",
			input:    `foreach [] --with-index 'i [n] [i]`,
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name: "foreach --with-index accumulates sum of indices",
			input: `sum: 0
foreach [100 200 300 400] --with-index 'pos [n] [sum: (+ sum pos)]
sum`,
			expected: value.NewIntVal(6), // 0+1+2+3 = 6
			wantErr:  false,
		},
		{
			name: "foreach --with-index overwrites existing binding",
			input: `pos: 99
foreach [1 2 3] --with-index 'pos [n] [pos]
`,
			expected: value.NewIntVal(2), // Last value of pos in loop iterations: 0, 1, 2
			wantErr:  false,
		},
		{
			name: "foreach --with-index with multiple variables",
			input: `result: []
foreach [1 2 3 4] --with-index 'idx [a b] [result: (append result idx)]
result`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(0), value.NewIntVal(1)}),
			wantErr:  false,
		},
		{
			name: "foreach --with-index with string series",
			input: `result: []
foreach "abc" --with-index 'pos [c] [result: (append result pos)]
result`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(0), value.NewIntVal(1), value.NewIntVal(2)}),
			wantErr:  false,
		},
		{
			name: "foreach --with-index combines index with value",
			input: `result: []
foreach [10 20 30] --with-index 'pos [n] [result: (append result (+ pos n))]
result`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(10), value.NewIntVal(21), value.NewIntVal(32)}),
			wantErr:  false,
		},
		{
			name: "foreach --with-index with odd-length series and multiple vars",
			input: `result: []
foreach [1 2 3 4 5] --with-index 'idx [a b] [
  result: (append result idx)
]
result`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(0), value.NewIntVal(1), value.NewIntVal(2)}),
			wantErr:  false,
		},
		{
			name:    "foreach --with-index non-word value fails",
			input:   `foreach [1 2 3] --with-index 42 [n] [n]`,
			wantErr: true,
		},
		{
			name:    "foreach --with-index with string value fails",
			input:   `foreach [1 2 3] --with-index "pos" [n] [n]`,
			wantErr: true,
		},
		{
			name:    "foreach --with-index with integer value fails",
			input:   `foreach [1 2 3] --with-index 123 [n] [n]`,
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

func TestControlFlow_Foreach(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "foreach returns last value",
			input:    "foreach [1 2 3] [n] [n]",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "foreach empty series returns none",
			input:    "foreach [] [n] [n]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "foreach with variable binding",
			input:    "foreach [1 2 3] [x] [(* x 2)]",
			expected: value.NewIntVal(6),
			wantErr:  false,
		},
		{
			name:     "foreach accumulates in outer variable",
			input:    "sum: 0\nforeach [10 20 30] [n] [sum: (+ sum n)]\nsum",
			expected: value.NewIntVal(60),
			wantErr:  false,
		},
		{
			name:     "foreach with single element",
			input:    "foreach [42] [x] [x]",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "foreach with string elements",
			input:    "result: \"\"\nforeach [\"a\" \"b\" \"c\"] [s] [result: (join result s)]\nresult",
			expected: value.NewStrVal("abc"),
			wantErr:  false,
		},
		{
			name:     "foreach with single word var (quoted)",
			input:    "foreach [1 2 3] n [n]",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "foreach with string series",
			input:    "result: \"\"\nforeach \"abc\" [c] [result: (join result c)]\nresult",
			expected: value.NewStrVal("abc"),
			wantErr:  false,
		},
		{
			name:     "foreach with string series and single word var",
			input:    "count: 0\nforeach \"hello\" c [count: (+ count 1)]\ncount",
			expected: value.NewIntVal(5),
			wantErr:  false,
		},
		{
			name:     "foreach with multiple vars",
			input:    "foreach [1 2 3 4] [a b] [b]",
			expected: value.NewIntVal(4),
			wantErr:  false,
		},
		{
			name:     "foreach with multiple vars accumulate",
			input:    "sum: 0\nforeach [1 2 3 4 5 6] [a b] [sum: (+ sum (+ a b))]\nsum",
			expected: value.NewIntVal(21),
			wantErr:  false,
		},
		{
			name:     "foreach with multiple vars odd length",
			input:    "foreach [1 2 3 4 5] [a b] [a]",
			expected: value.NewIntVal(5),
			wantErr:  false,
		},
		{
			name:     "foreach with multiple vars odd length binds none",
			input:    "result: []\nforeach \"abc\" [a b] [result: (append result a) result: (append result b)]\nresult",
			expected: value.NewBlockVal([]core.Value{value.NewStrVal("a"), value.NewStrVal("b"), value.NewStrVal("c"), value.NewNoneVal()}),
			wantErr:  false,
		},
		{
			name:    "foreach wrong arity zero args",
			input:   "foreach",
			wantErr: true,
		},
		{
			name:    "foreach wrong arity one arg",
			input:   "foreach [1 2 3]",
			wantErr: true,
		},
		{
			name:    "foreach wrong arity two args",
			input:   "foreach [1 2 3] [n]",
			wantErr: true,
		},
		{
			name:    "foreach non-series value",
			input:   "foreach 42 [n] [n]",
			wantErr: true,
		},
		{
			name:    "foreach non-block body",
			input:   "foreach [1 2 3] [n] n",
			wantErr: true,
		},
		{
			name:    "foreach non-word in variable block",
			input:   "foreach [1 2 3] [42] [n]",
			wantErr: true,
		},
		{
			name:    "foreach empty vars block",
			input:   "foreach [1 2 3] [] [n]",
			wantErr: true,
		},
		{
			name:    "foreach vars is integer",
			input:   "foreach [1 2 3] 42 [n]",
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
