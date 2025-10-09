// Package native implements built-in native functions for Viro.
//
// Math natives implement arithmetic operations with overflow detection.
// Contract per contracts/math.md: +, -, *, / operate on integers.
package native

import (
	"math"
	"strconv"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Add implements the + native function.
//
// Contract: + value1 value2 → sum
// - Arguments can be integers or decimals
// - Returns arithmetic sum with type promotion (integer + decimal → decimal)
// - Detects overflow
func Add(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), mathArityError("+", 2, len(args))
	}

	// Check if either argument is decimal - if so, promote to decimal arithmetic
	if args[0].Type == value.TypeDecimal || args[1].Type == value.TypeDecimal {
		return addDecimal(args[0], args[1])
	}

	// Integer arithmetic
	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("+", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("+", args[1])
	}

	// Check for overflow
	// Positive overflow: a > 0 && b > 0 && a > MaxInt64 - b
	// Negative overflow: a < 0 && b < 0 && a < MinInt64 - b
	if a > 0 && b > 0 && a > math.MaxInt64-b {
		return value.NoneVal(), overflowError("+")
	}
	if a < 0 && b < 0 && a < math.MinInt64-b {
		return value.NoneVal(), underflowError("+")
	}

	return value.IntVal(a + b), nil
}

// Subtract implements the - native function.
//
// Contract: - value1 value2 → difference
// - Arguments can be integers or decimals
// - Returns arithmetic difference (value1 - value2) with type promotion
// - Detects overflow
func Subtract(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), mathArityError("-", 2, len(args))
	}

	// Check if either argument is decimal - if so, promote to decimal arithmetic
	if args[0].Type == value.TypeDecimal || args[1].Type == value.TypeDecimal {
		return subtractDecimal(args[0], args[1])
	}

	// Integer arithmetic
	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("-", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("-", args[1])
	}

	// Check for overflow
	// a - b can overflow if:
	// - a > 0, b < 0, and a > MaxInt64 + b (result too large)
	// - a < 0, b > 0, and a < MinInt64 + b (result too small)
	if a > 0 && b < 0 && a > math.MaxInt64+b {
		return value.NoneVal(), overflowError("-")
	}
	if a < 0 && b > 0 && a < math.MinInt64+b {
		return value.NoneVal(), underflowError("-")
	}

	return value.IntVal(a - b), nil
}

// Multiply implements the * native function.
//
// Contract: * value1 value2 → product
// - Arguments can be integers or decimals
// - Returns arithmetic product with type promotion
// - Detects overflow
func Multiply(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), mathArityError("*", 2, len(args))
	}

	// Check if either argument is decimal - if so, promote to decimal arithmetic
	if args[0].Type == value.TypeDecimal || args[1].Type == value.TypeDecimal {
		return multiplyDecimal(args[0], args[1])
	}

	// Integer arithmetic
	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("*", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("*", args[1])
	}
	if !ok {
		return value.NoneVal(), mathTypeError("*", args[1])
	}

	// Check for overflow using division
	// If a * b would overflow, then (a * b) / b != a
	// Special cases: a == 0 or b == 0 (no overflow), and MinInt64 * -1 (overflows)
	if a == 0 || b == 0 {
		return value.IntVal(0), nil
	}

	if a == math.MinInt64 && b == -1 {
		return value.NoneVal(), overflowError("*")
	}
	if b == math.MinInt64 && a == -1 {
		return value.NoneVal(), overflowError("*")
	}

	result := a * b
	if result/b != a {
		return value.NoneVal(), overflowError("*")
	}

	return value.IntVal(result), nil
}

// Divide implements the / native function.
//
// Contract: / value1 value2 → quotient
// - Both arguments must be integers
// - Returns arithmetic quotient (truncated toward zero)
// - Division by zero is an error
func Divide(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), mathArityError("/", 2, len(args))
	}

	// Check if either argument is decimal - if so, promote to decimal arithmetic
	if args[0].Type == value.TypeDecimal || args[1].Type == value.TypeDecimal {
		return divideDecimal(args[0], args[1])
	}

	// Integer arithmetic
	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("/", args[0])
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), mathTypeError("/", args[1])
	}

	// Check for division by zero
	if b == 0 {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""})
	}

	// Check for overflow: MinInt64 / -1 overflows
	if a == math.MinInt64 && b == -1 {
		return value.NoneVal(), overflowError("/")
	}

	return value.IntVal(a / b), nil
}

func mathArityError(name string, expected, actual int) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDArgCount,
		[3]string{name, strconv.Itoa(expected), strconv.Itoa(actual)},
	)
}

func mathTypeError(name string, got value.Value) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDTypeMismatch,
		[3]string{name, "integer", got.Type.String()},
	)
}

func overflowError(op string) *verror.Error {
	return verror.NewMathError(verror.ErrIDOverflow, [3]string{op, "", ""})
}

func underflowError(op string) *verror.Error {
	return verror.NewMathError(verror.ErrIDUnderflow, [3]string{op, "", ""})
}

// LessThan implements the < native function.
//
// Contract: < value1 value2 → logic
// - Both arguments must be integers
// - Returns true if value1 < value2, false otherwise
func LessThan(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"< expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"< expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"< expects integer arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"> expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"> expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"> expects integer arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"<= expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"<= expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"<= expects integer arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{">= expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{">= expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{">= expects integer arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"= expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"= expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"= expects integer arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"<> expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"<> expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"<> expects integer arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"and expects 2 arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"or expects 2 arguments", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"not expects 1 argument", "", ""})
	}

	// Convert to truthy and negate (using ToTruthy from control.go)
	return value.LogicVal(!ToTruthy(args[0])), nil
}

// Decimal arithmetic promotion helpers (Feature 002)
// These functions implement mixed integer/decimal arithmetic with automatic promotion

func addDecimal(a, b value.Value) (value.Value, *verror.Error) {
	aVal := promoteToDecimal(a)
	bVal := promoteToDecimal(b)
	if aVal == nil || bVal == nil {
		return value.NoneVal(), verror.NewMathError("add-type-error", [3]string{a.Type.String(), b.Type.String(), ""})
	}
	
	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Add(result, aVal, bVal)
	
	return value.DecimalVal(result, 2), nil
}

func subtractDecimal(a, b value.Value) (value.Value, *verror.Error) {
	aVal := promoteToDecimal(a)
	bVal := promoteToDecimal(b)
	if aVal == nil || bVal == nil {
		return value.NoneVal(), verror.NewMathError("subtract-type-error", [3]string{a.Type.String(), b.Type.String(), ""})
	}
	
	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Sub(result, aVal, bVal)
	
	return value.DecimalVal(result, 2), nil
}

func multiplyDecimal(a, b value.Value) (value.Value, *verror.Error) {
	aVal := promoteToDecimal(a)
	bVal := promoteToDecimal(b)
	if aVal == nil || bVal == nil {
		return value.NoneVal(), verror.NewMathError("multiply-type-error", [3]string{a.Type.String(), b.Type.String(), ""})
	}
	
	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Mul(result, aVal, bVal)
	
	return value.DecimalVal(result, 2), nil
}

func divideDecimal(a, b value.Value) (value.Value, *verror.Error) {
	aVal := promoteToDecimal(a)
	bVal := promoteToDecimal(b)
	if aVal == nil || bVal == nil {
		return value.NoneVal(), verror.NewMathError("divide-type-error", [3]string{a.Type.String(), b.Type.String(), ""})
	}
	
	// Check for division by zero
	if bVal.Sign() == 0 {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""})
	}
	
	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Quo(result, aVal, bVal)
	
	return value.DecimalVal(result, 2), nil
}

