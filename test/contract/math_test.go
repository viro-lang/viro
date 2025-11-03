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
			args:     []core.Value{value.NewIntVal(3), value.NewIntVal(4)},
			expected: value.NewIntVal(7),
			wantErr:  false,
		},
		{
			name:     "add negative integers",
			op:       "+",
			args:     []core.Value{value.NewIntVal(-5), value.NewIntVal(10)},
			expected: value.NewIntVal(5),
			wantErr:  false,
		},
		{
			name:     "add with zero",
			op:       "+",
			args:     []core.Value{value.NewIntVal(42), value.NewIntVal(0)},
			expected: value.NewIntVal(42),
			wantErr:  false,
		},

		// Subtraction
		{
			name:     "subtract integers",
			op:       "-",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(3)},
			expected: value.NewIntVal(7),
			wantErr:  false,
		},
		{
			name:     "subtract to negative",
			op:       "-",
			args:     []core.Value{value.NewIntVal(3), value.NewIntVal(10)},
			expected: value.NewIntVal(-7),
			wantErr:  false,
		},

		// Multiplication
		{
			name:     "multiply integers",
			op:       "*",
			args:     []core.Value{value.NewIntVal(6), value.NewIntVal(7)},
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "multiply by zero",
			op:       "*",
			args:     []core.Value{value.NewIntVal(42), value.NewIntVal(0)},
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:     "multiply negative",
			op:       "*",
			args:     []core.Value{value.NewIntVal(-3), value.NewIntVal(4)},
			expected: value.NewIntVal(-12),
			wantErr:  false,
		},

		// Division
		{
			name:     "divide integers",
			op:       "/",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(3)},
			expected: value.NewIntVal(3), // Truncated toward zero
			wantErr:  false,
		},
		{
			name:     "divide negative",
			op:       "/",
			args:     []core.Value{value.NewIntVal(-10), value.NewIntVal(3)},
			expected: value.NewIntVal(-3), // Truncated toward zero
			wantErr:  false,
		},
		{
			name:     "divide by zero error",
			op:       "/",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(0)},
			expected: value.NewNoneVal(),
			wantErr:  true, // Math error: division by zero
		},

		// Type errors
		{
			name:     "add string to integer error",
			op:       "+",
			args:     []core.Value{value.NewStrVal("hello"), value.NewIntVal(5)},
			expected: value.NewNoneVal(),
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
// Contract: left-to-right evaluation, no operator precedence.
//
// Design decision: Viro now uses left-to-right evaluation, following Viro semantics.
func TestLeftToRightEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		expr     string // Expression as string (will be parsed)
		expected core.Value
	}{
		{
			name:     "left-to-right addition and multiplication",
			expr:     "3 + 4 * 2",
			expected: value.NewIntVal(14), // (3 + 4) * 2 = 7 * 2 = 14
		},
		{
			name:     "left-to-right subtraction and division",
			expr:     "10 - 6 / 2",
			expected: value.NewIntVal(2), // (10 - 6) / 2 = 4 / 2 = 2
		},
		{
			name:     "parentheses force specific order",
			expr:     "(3 + 4) * 2",
			expected: value.NewIntVal(14), // Same as left-to-right
		},
		{
			name:     "multiple operations left-to-right",
			expr:     "2 + 3 * 4 + 5",
			expected: value.NewIntVal(25), // ((2 + 3) * 4) + 5 = (5 * 4) + 5 = 20 + 5 = 25
		},
		{
			name:     "nested parentheses",
			expr:     "((2 + 3) * 4)",
			expected: value.NewIntVal(20),
		},
		{
			name:     "division then multiplication left-to-right",
			expr:     "20 / 2 * 3",
			expected: value.NewIntVal(30), // (20 / 2) * 3 = 10 * 3 = 30
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
			args:    []core.Value{value.NewIntVal(9223372036854775807), value.NewIntVal(1)},
			wantErr: true,
		},
		{
			name:    "subtraction underflow",
			op:      "-",
			args:    []core.Value{value.NewIntVal(-9223372036854775808), value.NewIntVal(1)},
			wantErr: true,
		},
		{
			name:    "multiplication overflow",
			op:      "*",
			args:    []core.Value{value.NewIntVal(9223372036854775807), value.NewIntVal(2)},
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
			if tt.wantErr && !result.Equals(value.NewNoneVal()) {
				t.Errorf("%s(%v) on error should return none, got %v", tt.op, tt.args, result)
			}
		})
	}
}

func TestNativeMod(t *testing.T) {
	tests := []struct {
		name     string
		args     []core.Value
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "basic modulo",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(3)},
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		{
			name:     "zero remainder",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(5)},
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:     "negative dividend",
			args:     []core.Value{value.NewIntVal(-10), value.NewIntVal(3)},
			expected: value.NewIntVal(-1),
			wantErr:  false,
		},
		{
			name:     "modulo by 1",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(1)},
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:     "large numbers",
			args:     []core.Value{value.NewIntVal(1000000), value.NewIntVal(7)},
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		{
			name:     "division by zero error",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(0)},
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name:     "type error non-integer first arg",
			args:     []core.Value{value.NewStrVal("hello"), value.NewIntVal(3)},
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name:     "type error non-integer second arg",
			args:     []core.Value{value.NewIntVal(10), value.NewStrVal("hello")},
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name:     "arity error too few args",
			args:     []core.Value{value.NewIntVal(10)},
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name:     "arity error too many args",
			args:     []core.Value{value.NewIntVal(10), value.NewIntVal(3), value.NewIntVal(5)},
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()

			result, err := native.Mod(tt.args, map[string]core.Value{}, e)

			if (err != nil) != tt.wantErr {
				t.Errorf("Mod(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Mod(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestEqualityAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		args     []core.Value
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "integers equal",
			op:       "=",
			args:     []core.Value{value.NewIntVal(42), value.NewIntVal(42)},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "integers not equal",
			op:       "=",
			args:     []core.Value{value.NewIntVal(42), value.NewIntVal(43)},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "strings equal",
			op:       "=",
			args:     []core.Value{value.NewStrVal("hello"), value.NewStrVal("hello")},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "strings not equal",
			op:       "=",
			args:     []core.Value{value.NewStrVal("hello"), value.NewStrVal("world")},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "strings case sensitive",
			op:       "=",
			args:     []core.Value{value.NewStrVal("Hello"), value.NewStrVal("hello")},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "empty strings equal",
			op:       "=",
			args:     []core.Value{value.NewStrVal(""), value.NewStrVal("")},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "logic values true equals true",
			op:       "=",
			args:     []core.Value{value.NewLogicVal(true), value.NewLogicVal(true)},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "logic values false equals false",
			op:       "=",
			args:     []core.Value{value.NewLogicVal(false), value.NewLogicVal(false)},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "logic values true not equals false",
			op:       "=",
			args:     []core.Value{value.NewLogicVal(true), value.NewLogicVal(false)},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "none values equal",
			op:       "=",
			args:     []core.Value{value.NewNoneVal(), value.NewNoneVal()},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "blocks equal",
			op:       "=",
			args:     []core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}), value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)})},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "blocks not equal different values",
			op:       "=",
			args:     []core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}), value.NewBlockVal([]core.Value{value.NewIntVal(2), value.NewIntVal(1)})},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "blocks not equal different length",
			op:       "=",
			args:     []core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}), value.NewBlockVal([]core.Value{value.NewIntVal(1)})},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "empty blocks equal",
			op:       "=",
			args:     []core.Value{value.NewBlockVal([]core.Value{}), value.NewBlockVal([]core.Value{})},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "nested blocks equal",
			op:       "=",
			args:     []core.Value{value.NewBlockVal([]core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}), value.NewIntVal(3)}), value.NewBlockVal([]core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}), value.NewIntVal(3)})},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "type mismatch returns false not error",
			op:       "=",
			args:     []core.Value{value.NewIntVal(5), value.NewStrVal("5")},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "type mismatch logic and int",
			op:       "=",
			args:     []core.Value{value.NewLogicVal(true), value.NewIntVal(1)},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "type mismatch block and string",
			op:       "=",
			args:     []core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1)}), value.NewStrVal("[1]")},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "not-equal integers different",
			op:       "<>",
			args:     []core.Value{value.NewIntVal(42), value.NewIntVal(43)},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "not-equal integers same",
			op:       "<>",
			args:     []core.Value{value.NewIntVal(42), value.NewIntVal(42)},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "not-equal strings different",
			op:       "<>",
			args:     []core.Value{value.NewStrVal("hello"), value.NewStrVal("world")},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "not-equal strings same",
			op:       "<>",
			args:     []core.Value{value.NewStrVal("hello"), value.NewStrVal("hello")},
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "not-equal type mismatch",
			op:       "<>",
			args:     []core.Value{value.NewIntVal(5), value.NewStrVal("5")},
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "equal arity error too few",
			op:       "=",
			args:     []core.Value{value.NewIntVal(5)},
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name:     "equal arity error too many",
			op:       "=",
			args:     []core.Value{value.NewIntVal(5), value.NewIntVal(5), value.NewIntVal(5)},
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()

			var result core.Value
			var err error

			switch tt.op {
			case "=":
				result, err = native.Equal(tt.args, map[string]core.Value{}, e)
			case "<>":
				result, err = native.NotEqual(tt.args, map[string]core.Value{}, e)
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
