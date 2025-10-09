// Package native implements built-in native functions for Viro.
//
// Math natives implement arithmetic operations with overflow detection.
// Contract per contracts/math.md: +, -, *, / operate on integers.
package native

import (
	"math"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// intOp represents an integer arithmetic operation that may overflow.
// Returns the result and a boolean indicating if overflow occurred.
type intOp func(a, b int64) (result int64, overflow bool)

// decimalOp represents a decimal arithmetic operation.
type decimalOp func(ctx decimal.Context, result, a, b *decimal.Big) *decimal.Big

// mathOp provides a generic template for binary arithmetic operations.
// It handles type checking, decimal promotion, and overflow detection.
func mathOp(name string, args []value.Value, intFn intOp, decFn decimalOp) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError(name, 2, len(args))
	}

	// Check if either argument is decimal - if so, promote to decimal arithmetic
	if args[0].Type == value.TypeDecimal || args[1].Type == value.TypeDecimal {
		return decimalMathOp(name, args[0], args[1], decFn)
	}

	// Integer arithmetic
	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError(name, args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError(name, args[1])
	}

	// Perform operation with overflow check
	result, overflow := intFn(a, b)
	if overflow {
		return value.NoneVal(), overflowError(name)
	}

	return value.IntVal(result), nil
}

// decimalMathOp handles decimal arithmetic with promotion.
func decimalMathOp(name string, a, b value.Value, decFn decimalOp) (value.Value, *verror.Error) {
	aVal := promoteToDecimal(a)
	bVal := promoteToDecimal(b)
	if aVal == nil || bVal == nil {
		return value.NoneVal(), verror.NewMathError(name+"-type-error", [3]string{a.Type.String(), b.Type.String(), ""})
	}

	// Check for division by zero if the operation might divide
	if name == "/" && bVal.Sign() == 0 {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	decFn(ctx, result, aVal, bVal)

	return value.DecimalVal(result, 2), nil
}

// Add implements the + native function.
//
// Contract: + value1 value2 → sum
// - Arguments can be integers or decimals
// - Returns arithmetic sum with type promotion (integer + decimal → decimal)
// - Detects overflow
func Add(args []value.Value) (value.Value, *verror.Error) {
	return mathOp("+", args,
		func(a, b int64) (int64, bool) {
			// Check for overflow
			// Positive overflow: a > 0 && b > 0 && a > MaxInt64 - b
			// Negative overflow: a < 0 && b < 0 && a < MinInt64 - b
			if a > 0 && b > 0 && a > math.MaxInt64-b {
				return 0, true
			}
			if a < 0 && b < 0 && a < math.MinInt64-b {
				return 0, true
			}
			return a + b, false
		},
		func(ctx decimal.Context, result, a, b *decimal.Big) *decimal.Big {
			return ctx.Add(result, a, b)
		})
}

// Subtract implements the - native function.
//
// Contract: - value1 value2 → difference
// - Arguments can be integers or decimals
// - Returns arithmetic difference (value1 - value2) with type promotion
// - Detects overflow
func Subtract(args []value.Value) (value.Value, *verror.Error) {
	return mathOp("-", args,
		func(a, b int64) (int64, bool) {
			// Check for overflow
			// a - b can overflow if:
			// - a > 0, b < 0, and a > MaxInt64 + b (result too large)
			// - a < 0, b > 0, and a < MinInt64 + b (result too small)
			if a > 0 && b < 0 && a > math.MaxInt64+b {
				return 0, true
			}
			if a < 0 && b > 0 && a < math.MinInt64+b {
				return 0, true
			}
			return a - b, false
		},
		func(ctx decimal.Context, result, a, b *decimal.Big) *decimal.Big {
			return ctx.Sub(result, a, b)
		})
}

// Multiply implements the * native function.
//
// Contract: * value1 value2 → product
// - Arguments can be integers or decimals
// - Returns arithmetic product with type promotion
// - Detects overflow
func Multiply(args []value.Value) (value.Value, *verror.Error) {
	return mathOp("*", args,
		func(a, b int64) (int64, bool) {
			// Special cases: a == 0 or b == 0 (no overflow)
			if a == 0 || b == 0 {
				return 0, false
			}

			// MinInt64 * -1 overflows
			if a == math.MinInt64 && b == -1 {
				return 0, true
			}
			if b == math.MinInt64 && a == -1 {
				return 0, true
			}

			result := a * b
			// Check for overflow using division
			if result/b != a {
				return 0, true
			}

			return result, false
		},
		func(ctx decimal.Context, result, a, b *decimal.Big) *decimal.Big {
			return ctx.Mul(result, a, b)
		})
}

// Divide implements the / native function.
//
// Contract: / value1 value2 → quotient
// - Both arguments must be integers
// - Returns arithmetic quotient (truncated toward zero)
// - Division by zero is an error
func Divide(args []value.Value) (value.Value, *verror.Error) {
	// Special handling for division by zero in integer case
	if len(args) == 2 && args[0].Type != value.TypeDecimal && args[1].Type != value.TypeDecimal {
		if b, ok := args[1].AsInteger(); ok && b == 0 {
			return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""})
		}
	}

	return mathOp("/", args,
		func(a, b int64) (int64, bool) {
			// Check for overflow: MinInt64 / -1 overflows
			if a == math.MinInt64 && b == -1 {
				return 0, true
			}
			return a / b, false
		},
		func(ctx decimal.Context, result, a, b *decimal.Big) *decimal.Big {
			return ctx.Quo(result, a, b)
		})
}

// LessThan implements the < native function.
//
// Contract: < value1 value2 → logic
// - Both arguments must be integers
// - Returns true if value1 < value2, false otherwise
func LessThan(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("<", 2, len(args))
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("<", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("<", args[1])
	}

	return value.LogicVal(a < b), nil
}

// GreaterThan implements the > native function.
//
// Contract: > value1 value2 → logic
// - Both arguments must be integers
// - Returns true if value1 > value2, false otherwise
func GreaterThan(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError(">", 2, len(args))
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError(">", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError(">", args[1])
	}

	return value.LogicVal(a > b), nil
}

// LessOrEqual implements the <= native function.
//
// Contract: <= value1 value2 → logic
// - Both arguments must be integers
// - Returns true if value1 <= value2, false otherwise
func LessOrEqual(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("<=", 2, len(args))
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("<=", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("<=", args[1])
	}

	return value.LogicVal(a <= b), nil
}

// GreaterOrEqual implements the >= native function.
//
// Contract: >= value1 value2 → logic
// - Both arguments must be integers
// - Returns true if value1 >= value2, false otherwise
func GreaterOrEqual(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError(">=", 2, len(args))
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError(">=", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError(">=", args[1])
	}

	return value.LogicVal(a >= b), nil
}

// Equal implements the = native function.
//
// Contract: = value1 value2 → logic
// - Both arguments must be integers
// - Returns true if value1 == value2, false otherwise
func Equal(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("=", 2, len(args))
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("=", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("=", args[1])
	}

	return value.LogicVal(a == b), nil
}

// NotEqual implements the <> native function.
//
// Contract: <> value1 value2 → logic
// - Both arguments must be integers
// - Returns true if value1 != value2, false otherwise
func NotEqual(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("<>", 2, len(args))
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("<>", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("<>", args[1])
	}

	return value.LogicVal(a != b), nil
}

// And implements the and native function.
//
// Contract: and value1 value2 → logic
// - Both arguments evaluated to logic (truthy conversion)
// - Returns true if both are truthy, false otherwise
// - Truthy: none/false → false, all others → true
func And(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("and", 2, len(args))
	}

	// Convert both to truthy (using ToTruthy from control.go)
	a := ToTruthy(args[0])
	b := ToTruthy(args[1])

	return value.LogicVal(a && b), nil
}

// Or implements the or native function.
//
// Contract: or value1 value2 → logic
// - Both arguments evaluated to logic (truthy conversion)
// - Returns true if either is truthy, false if both falsy
// - Truthy: none/false → false, all others → true
func Or(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("or", 2, len(args))
	}

	// Convert both to truthy (using ToTruthy from control.go)
	a := ToTruthy(args[0])
	b := ToTruthy(args[1])

	return value.LogicVal(a || b), nil
}

// Not implements the not native function.
//
// Contract: not value → logic
// - Argument evaluated to logic (truthy conversion)
// - Returns negation: true if falsy, false if truthy
// - Truthy: none/false → false, all others → true
func Not(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("not", 1, len(args))
	}

	// Convert to truthy and negate (using ToTruthy from control.go)
	return value.LogicVal(!ToTruthy(args[0])), nil
}
