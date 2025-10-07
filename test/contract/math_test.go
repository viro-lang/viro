package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// TestArithmeticNatives tests basic arithmetic operations.
// Contract per contracts/math.md: +, -, *, / operate on integers.
//
// TDD: This test is written FIRST and will FAIL until natives are implemented.
func TestArithmeticNatives(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		args     []value.Value
		expected value.Value
		wantErr  bool
	}{
		// Addition
		{
			name:     "add positive integers",
			op:       "+",
			args:     []value.Value{value.IntVal(3), value.IntVal(4)},
			expected: value.IntVal(7),
			wantErr:  false,
		},
		{
			name:     "add negative integers",
			op:       "+",
			args:     []value.Value{value.IntVal(-5), value.IntVal(10)},
			expected: value.IntVal(5),
			wantErr:  false,
		},
		{
			name:     "add with zero",
			op:       "+",
			args:     []value.Value{value.IntVal(42), value.IntVal(0)},
			expected: value.IntVal(42),
			wantErr:  false,
		},

		// Subtraction
		{
			name:     "subtract integers",
			op:       "-",
			args:     []value.Value{value.IntVal(10), value.IntVal(3)},
			expected: value.IntVal(7),
			wantErr:  false,
		},
		{
			name:     "subtract to negative",
			op:       "-",
			args:     []value.Value{value.IntVal(3), value.IntVal(10)},
			expected: value.IntVal(-7),
			wantErr:  false,
		},

		// Multiplication
		{
			name:     "multiply integers",
			op:       "*",
			args:     []value.Value{value.IntVal(6), value.IntVal(7)},
			expected: value.IntVal(42),
			wantErr:  false,
		},
		{
			name:     "multiply by zero",
			op:       "*",
			args:     []value.Value{value.IntVal(42), value.IntVal(0)},
			expected: value.IntVal(0),
			wantErr:  false,
		},
		{
			name:     "multiply negative",
			op:       "*",
			args:     []value.Value{value.IntVal(-3), value.IntVal(4)},
			expected: value.IntVal(-12),
			wantErr:  false,
		},

		// Division
		{
			name:     "divide integers",
			op:       "/",
			args:     []value.Value{value.IntVal(10), value.IntVal(3)},
			expected: value.IntVal(3), // Truncated toward zero
			wantErr:  false,
		},
		{
			name:     "divide negative",
			op:       "/",
			args:     []value.Value{value.IntVal(-10), value.IntVal(3)},
			expected: value.IntVal(-3), // Truncated toward zero
			wantErr:  false,
		},
		{
			name:     "divide by zero error",
			op:       "/",
			args:     []value.Value{value.IntVal(10), value.IntVal(0)},
			expected: value.NoneVal(),
			wantErr:  true, // Math error: division by zero
		},

		// Type errors
		{
			name:     "add string to integer error",
			op:       "+",
			args:     []value.Value{value.StrVal("hello"), value.IntVal(5)},
			expected: value.NoneVal(),
			wantErr:  true, // Type mismatch error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the appropriate native function
			var result value.Value
			var err *verror.Error

			switch tt.op {
			case "+":
				result, err = native.Add(tt.args)
			case "-":
				result, err = native.Subtract(tt.args)
			case "*":
				result, err = native.Multiply(tt.args)
			case "/":
				result, err = native.Divide(tt.args)
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

// TestOperatorPrecedence tests that parser respects traditional precedence.
// Contract per contracts/math.md: * and / before + and -.
//
// Critical design decision: Viro uses traditional precedence, NOT REBOL's left-to-right.
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		expr     string // Expression as string (will be parsed)
		expected value.Value
	}{
		{
			name:     "multiplication before addition",
			expr:     "3 + 4 * 2",
			expected: value.IntVal(11), // Not 14 (left-to-right)
		},
		{
			name:     "division before subtraction",
			expr:     "10 - 6 / 2",
			expected: value.IntVal(7), // Not 2 (left-to-right)
		},
		{
			name:     "parentheses override precedence",
			expr:     "(3 + 4) * 2",
			expected: value.IntVal(14),
		},
		{
			name:     "multiple operations",
			expr:     "2 + 3 * 4 + 5",
			expected: value.IntVal(19), // 2 + 12 + 5
		},
		{
			name:     "nested parentheses",
			expr:     "((2 + 3) * 4)",
			expected: value.IntVal(20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse expression
			values, err := parse.Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%s) error: %v", tt.expr, err)
			}

			// Evaluate parsed values
			e := eval.NewEvaluator()
			result, err := e.Do_Blk(values)
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
		args    []value.Value
		wantErr bool
	}{
		{
			name:    "addition overflow",
			op:      "+",
			args:    []value.Value{value.IntVal(9223372036854775807), value.IntVal(1)},
			wantErr: true,
		},
		{
			name:    "subtraction underflow",
			op:      "-",
			args:    []value.Value{value.IntVal(-9223372036854775808), value.IntVal(1)},
			wantErr: true,
		},
		{
			name:    "multiplication overflow",
			op:      "*",
			args:    []value.Value{value.IntVal(9223372036854775807), value.IntVal(2)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the appropriate native function
			var result value.Value
			var err *verror.Error

			switch tt.op {
			case "+":
				result, err = native.Add(tt.args)
			case "-":
				result, err = native.Subtract(tt.args)
			case "*":
				result, err = native.Multiply(tt.args)
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
