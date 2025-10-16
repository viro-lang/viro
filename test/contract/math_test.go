package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestArithmeticNatives tests basic arithmetic operations.
// Contract per contracts/math.md: +, -, *, / operate on integers.
//
// TDD: This test is written FIRST and will FAIL until natives are implemented.
func TestArithmeticNatives(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		args     []core.Value
		expected core.Value
		wantErr  bool
	}{
		// Addition
		{
			name:     "add positive integers",
			op:       "+",
			args:     []core.Value{value.IntVal(3), value.IntVal(4)},
			expected: value.IntVal(7),
			wantErr:  false,
		},
		{
			name:     "add negative integers",
			op:       "+",
			args:     []core.Value{value.IntVal(-5), value.IntVal(10)},
			expected: value.IntVal(5),
			wantErr:  false,
		},
		{
			name:     "add with zero",
			op:       "+",
			args:     []core.Value{value.IntVal(42), value.IntVal(0)},
			expected: value.IntVal(42),
			wantErr:  false,
		},

		// Subtraction
		{
			name:     "subtract integers",
			op:       "-",
			args:     []core.Value{value.IntVal(10), value.IntVal(3)},
			expected: value.IntVal(7),
			wantErr:  false,
		},
		{
			name:     "subtract to negative",
			op:       "-",
			args:     []core.Value{value.IntVal(3), value.IntVal(10)},
			expected: value.IntVal(-7),
			wantErr:  false,
		},

		// Multiplication
		{
			name:     "multiply integers",
			op:       "*",
			args:     []core.Value{value.IntVal(6), value.IntVal(7)},
			expected: value.IntVal(42),
			wantErr:  false,
		},
		{
			name:     "multiply by zero",
			op:       "*",
			args:     []core.Value{value.IntVal(42), value.IntVal(0)},
			expected: value.IntVal(0),
			wantErr:  false,
		},
		{
			name:     "multiply negative",
			op:       "*",
			args:     []core.Value{value.IntVal(-3), value.IntVal(4)},
			expected: value.IntVal(-12),
			wantErr:  false,
		},

		// Division
		{
			name:     "divide integers",
			op:       "/",
			args:     []core.Value{value.IntVal(10), value.IntVal(3)},
			expected: value.IntVal(3), // Truncated toward zero
			wantErr:  false,
		},
		{
			name:     "divide negative",
			op:       "/",
			args:     []core.Value{value.IntVal(-10), value.IntVal(3)},
			expected: value.IntVal(-3), // Truncated toward zero
			wantErr:  false,
		},
		{
			name:     "divide by zero error",
			op:       "/",
			args:     []core.Value{value.IntVal(10), value.IntVal(0)},
			expected: value.NoneVal(),
			wantErr:  true, // Math error: division by zero
		},

		// Type errors
		{
			name:     "add string to integer error",
			op:       "+",
			args:     []core.Value{value.StrVal("hello"), value.IntVal(5)},
			expected: value.NoneVal(),
			wantErr:  true, // Type mismatch error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()

			// Call the appropriate native function
			var result core.Value
			var err error

			switch tt.op {
			case "+":
				result, err = native.Add(tt.args, map[string]core.Value{}, e)
			case "-":
				result, err = native.Subtract(tt.args, map[string]core.Value{}, e)
			case "*":
				result, err = native.Multiply(tt.args, map[string]core.Value{}, e)
			case "/":
				result, err = native.Divide(tt.args, map[string]core.Value{}, e)
			default:
				t.Fatalf("Unknown operator: %s", tt.op)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%v) error = %v, wantErr %v", tt.op, tt.args, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("%s(%v) = %v, want %v", tt.op, tt.args, result, tt.expected)
			}
		})
	}
}

// TestLeftToRightEvaluation tests that parser uses left-to-right evaluation.
// Contract: REBOL-style left-to-right evaluation, no operator precedence.
//
// Design decision: Viro now uses left-to-right evaluation, matching REBOL.
func TestLeftToRightEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		expr     string // Expression as string (will be parsed)
		expected core.Value
	}{
		{
			name:     "left-to-right addition and multiplication",
			expr:     "3 + 4 * 2",
			expected: value.IntVal(14), // (3 + 4) * 2 = 7 * 2 = 14
		},
		{
			name:     "left-to-right subtraction and division",
			expr:     "10 - 6 / 2",
			expected: value.IntVal(2), // (10 - 6) / 2 = 4 / 2 = 2
		},
		{
			name:     "parentheses force specific order",
			expr:     "(3 + 4) * 2",
			expected: value.IntVal(14), // Same as left-to-right
		},
		{
			name:     "multiple operations left-to-right",
			expr:     "2 + 3 * 4 + 5",
			expected: value.IntVal(25), // ((2 + 3) * 4) + 5 = (5 * 4) + 5 = 20 + 5 = 25
		},
		{
			name:     "nested parentheses",
			expr:     "((2 + 3) * 4)",
			expected: value.IntVal(20),
		},
		{
			name:     "division then multiplication left-to-right",
			expr:     "20 / 2 * 3",
			expected: value.IntVal(30), // (20 / 2) * 3 = 10 * 3 = 30
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Evaluate expression
			result, err := Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("Eval(%s) error: %v", tt.expr, err)
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Eval(%s) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

// TestArithmeticOverflow tests that overflow is detected.
// Contract per contracts/math.md: Detect and error on overflow/underflow.
func TestArithmeticOverflow(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		args    []core.Value
		wantErr bool
	}{
		{
			name:    "addition overflow",
			op:      "+",
			args:    []core.Value{value.IntVal(9223372036854775807), value.IntVal(1)},
			wantErr: true,
		},
		{
			name:    "subtraction underflow",
			op:      "-",
			args:    []core.Value{value.IntVal(-9223372036854775808), value.IntVal(1)},
			wantErr: true,
		},
		{
			name:    "multiplication overflow",
			op:      "*",
			args:    []core.Value{value.IntVal(9223372036854775807), value.IntVal(2)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()

			// Call the appropriate native function
			var result core.Value
			var err error

			switch tt.op {
			case "+":
				result, err = native.Add(tt.args, map[string]core.Value{}, e)
			case "-":
				result, err = native.Subtract(tt.args, map[string]core.Value{}, e)
			case "*":
				result, err = native.Multiply(tt.args, map[string]core.Value{}, e)
			default:
				t.Fatalf("Unknown operator: %s", tt.op)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%v) error = %v, wantErr %v", tt.op, tt.args, err, tt.wantErr)
			}

			// Verify result is none on error
			if tt.wantErr && !result.Equals(value.NoneVal()) {
				t.Errorf("%s(%v) on error should return none, got %v", tt.op, tt.args, result)
			}
		})
	}
}
