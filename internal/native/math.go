// Package native implements built-in native functions for Viro.
//
// Math natives implement arithmetic operations with overflow detection.
// Contract per contracts/math.md: +, -, *, / operate on integers.
package native

import (
	"math"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// intOp represents an integer arithmetic operation that may overflow.
// Returns the result and a boolean indicating if overflow occurred.
type intOp func(a, b int64) (result int64, overflow bool)

// decimalOp represents a decimal arithmetic operation.
type decimalOp func(ctx decimal.Context, result, a, b *decimal.Big) *decimal.Big

// intCompareFn represents an integer comparison operation.
// Returns true if the comparison holds.
type intCompareFn func(a, b int64) bool

// decimalCompareFn represents a decimal comparison operation.
// Returns true if the comparison holds.
type decimalCompareFn func(a, b *decimal.Big) bool

// mathOp provides a generic template for binary arithmetic operations.
// It handles type checking, decimal promotion, and overflow detection.
func mathOp(name string, args []core.Value, intFn intOp, decFn decimalOp) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError(name, 2, len(args))
	}

	// Check if either argument is decimal - if so, promote to decimal arithmetic
	if args[0].GetType() == value.TypeDecimal || args[1].GetType() == value.TypeDecimal {
		return decimalMathOp(name, args[0], args[1], decFn)
	}

	// Integer arithmetic
	a, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), mathTypeError(name, args[0])
	}

	b, ok := value.AsIntValue(args[1])
	if !ok {
		return value.NewNoneVal(), mathTypeError(name, args[1])
	}

	// Perform operation with overflow check
	result, overflow := intFn(a, b)
	if overflow {
		return value.NewNoneVal(), overflowError(name)
	}

	return value.NewIntVal(result), nil
}

// decimalMathOp handles decimal arithmetic with promotion.
func decimalMathOp(name string, a, b core.Value, decFn decimalOp) (core.Value, error) {
	aVal := promoteToDecimal(a, nil, nil)
	bVal := promoteToDecimal(b, nil, nil)
	if aVal == nil || bVal == nil {
		return value.NewNoneVal(), verror.NewMathError(name+"-type-error", [3]string{value.TypeToString(a.GetType()), value.TypeToString(b.GetType()), ""})
	}

	// Check for division by zero if the operation might divide
	if name == "/" && bVal.Sign() == 0 {
		return value.NewNoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	decFn(ctx, result, aVal, bVal)

	return value.DecimalVal(result, 2), nil
}

// compareOp provides a generic template for binary comparison operations.
// It handles type checking and decimal promotion.
func compareOp(name string, args []core.Value, intFn intCompareFn, decFn decimalCompareFn) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError(name, 2, len(args))
	}

	// Check if either argument is decimal - if so, promote to decimal comparison
	if args[0].GetType() == value.TypeDecimal || args[1].GetType() == value.TypeDecimal {
		return decimalCompareOp(name, args[0], args[1], decFn)
	}

	// Integer comparison
	a, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), mathTypeError(name, args[0])
	}

	b, ok := value.AsIntValue(args[1])
	if !ok {
		return value.NewNoneVal(), mathTypeError(name, args[1])
	}

	return value.NewLogicVal(intFn(a, b)), nil
}

// decimalCompareOp handles decimal comparison with promotion.
func decimalCompareOp(name string, a, b core.Value, decFn decimalCompareFn) (core.Value, error) {
	aVal := promoteToDecimal(a, nil, nil)
	bVal := promoteToDecimal(b, nil, nil)
	if aVal == nil || bVal == nil {
		return value.NewNoneVal(), verror.NewMathError(name+"-type-error", [3]string{value.TypeToString(a.GetType()), value.TypeToString(b.GetType()), ""})
	}

	return value.NewLogicVal(decFn(aVal, bVal)), nil
}

// Add implements the + native function.
//
// Contract: + value1 value2 → sum
// - Arguments can be integers or decimals
// - Returns arithmetic sum with type promotion (integer + decimal → decimal)
// - Detects overflow
func Add(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
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
func Subtract(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
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
func Multiply(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
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
func Divide(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// Special handling for division by zero in integer case
	if len(args) == 2 && args[0].GetType() != value.TypeDecimal && args[1].GetType() != value.TypeDecimal {
		if b, ok := value.AsIntValue(args[1]); ok && b == 0 {
			return value.NewNoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""})
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

// Mod implements the mod native function.
//
// Contract: mod value1 value2 → remainder
// - Arguments can be integers or decimals
// - Returns remainder after division (modulo operation)
// - Division by zero is an error
// - Supports infix notation
func Mod(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 2 && args[0].GetType() != value.TypeDecimal && args[1].GetType() != value.TypeDecimal {
		if b, ok := value.AsIntValue(args[1]); ok && b == 0 {
			return value.NewNoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""})
		}
	}

	return mathOp("mod", args,
		func(a, b int64) (int64, bool) {
			return a % b, false
		},
		func(ctx decimal.Context, result, a, b *decimal.Big) *decimal.Big {
			return ctx.Rem(result, a, b)
		})
}

// LessThan implements the < native function.
//
// Contract: < value1 value2 → logic
// - Arguments can be integers or decimals
// - Returns true if value1 < value2, false otherwise
func LessThan(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return compareOp("<", args,
		func(a, b int64) bool { return a < b },
		func(a, b *decimal.Big) bool { return a.Cmp(b) < 0 })
}

// GreaterThan implements the > native function.
//
// Contract: > value1 value2 → logic
// - Arguments can be integers or decimals
// - Returns true if value1 > value2, false otherwise
func GreaterThan(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return compareOp(">", args,
		func(a, b int64) bool { return a > b },
		func(a, b *decimal.Big) bool { return a.Cmp(b) > 0 })
}

// LessOrEqual implements the <= native function.
//
// Contract: <= value1 value2 → logic
// - Arguments can be integers or decimals
// - Returns true if value1 <= value2, false otherwise
func LessOrEqual(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return compareOp("<=", args,
		func(a, b int64) bool { return a <= b },
		func(a, b *decimal.Big) bool { return a.Cmp(b) <= 0 })
}

// GreaterOrEqual implements the >= native function.
//
// Contract: >= value1 value2 → logic
// - Arguments can be integers or decimals
// - Returns true if value1 >= value2, false otherwise
func GreaterOrEqual(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return compareOp(">=", args,
		func(a, b int64) bool { return a >= b },
		func(a, b *decimal.Big) bool { return a.Cmp(b) >= 0 })
}

// Equal implements the = native function.
//
// Contract: = value1 value2 → logic
// - Arguments can be any type
// - Returns true if value1 equals value2, false otherwise
// - Uses polymorphic Equals method for type-specific comparison
func Equal(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("=", 2, len(args))
	}
	return value.NewLogicVal(args[0].Equals(args[1])), nil
}

// NotEqual implements the <> native function.
//
// Contract: <> value1 value2 → logic
// - Arguments can be any type
// - Returns true if value1 does not equal value2, false otherwise
// - Uses polymorphic Equals method for type-specific comparison
func NotEqual(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("<>", 2, len(args))
	}
	return value.NewLogicVal(!args[0].Equals(args[1])), nil
}

// And implements the and native function.
//
// Contract: and value1 value2 → logic
// - Both arguments evaluated to logic (truthy conversion)
// - Returns true if both are truthy, false otherwise
// - Truthy: none/false → false, all others → true
func And(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("and", 2, len(args))
	}

	// Convert both to truthy (using ToTruthy from control.go)
	a := ToTruthy(args[0])
	b := ToTruthy(args[1])

	return value.NewLogicVal(a && b), nil
}

// Or implements the or native function.
//
// Contract: or value1 value2 → logic
// - Both arguments evaluated to logic (truthy conversion)
// - Returns true if either is truthy, false if both falsy
// - Truthy: none/false → false, all others → true
func Or(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("or", 2, len(args))
	}

	// Convert both to truthy (using ToTruthy from control.go)
	a := ToTruthy(args[0])
	b := ToTruthy(args[1])

	return value.NewLogicVal(a || b), nil
}

// Not implements the not native function.
//
// Contract: not value → logic
// - Argument evaluated to logic (truthy conversion)
// - Returns negation: true if falsy, false if truthy
// - Truthy: none/false → false, all others → true
func Not(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("not", 1, len(args))
	}

	// Convert to truthy and negate (using ToTruthy from control.go)
	return value.NewLogicVal(!ToTruthy(args[0])), nil
}
