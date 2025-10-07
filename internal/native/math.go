// Package native implements built-in native functions for Viro.
//
// Math natives implement arithmetic operations with overflow detection.
// Contract per contracts/math.md: +, -, *, / operate on integers.
package native

import (
	"math"

	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Add implements the + native function.
//
// Contract: + value1 value2 → sum
// - Both arguments must be integers
// - Returns arithmetic sum
// - Detects overflow
func Add(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"+ expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"+ expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"+ expects integer arguments", "", ""})
	}

	// Check for overflow
	// Positive overflow: a > 0 && b > 0 && a > MaxInt64 - b
	// Negative overflow: a < 0 && b < 0 && a < MinInt64 - b
	if a > 0 && b > 0 && a > math.MaxInt64-b {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer overflow in addition", "", ""})
	}
	if a < 0 && b < 0 && a < math.MinInt64-b {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer underflow in addition", "", ""})
	}

	return value.IntVal(a + b), nil
}

// Subtract implements the - native function.
//
// Contract: - value1 value2 → difference
// - Both arguments must be integers
// - Returns arithmetic difference (value1 - value2)
// - Detects overflow
func Subtract(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"- expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"- expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"- expects integer arguments", "", ""})
	}

	// Check for overflow
	// a - b can overflow if:
	// - a > 0, b < 0, and a > MaxInt64 + b (result too large)
	// - a < 0, b > 0, and a < MinInt64 + b (result too small)
	if a > 0 && b < 0 && a > math.MaxInt64+b {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer overflow in subtraction", "", ""})
	}
	if a < 0 && b > 0 && a < math.MinInt64+b {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer underflow in subtraction", "", ""})
	}

	return value.IntVal(a - b), nil
}

// Multiply implements the * native function.
//
// Contract: * value1 value2 → product
// - Both arguments must be integers
// - Returns arithmetic product
// - Detects overflow
func Multiply(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"* expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"* expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"* expects integer arguments", "", ""})
	}

	// Check for overflow using division
	// If a * b would overflow, then (a * b) / b != a
	// Special cases: a == 0 or b == 0 (no overflow), and MinInt64 * -1 (overflows)
	if a == 0 || b == 0 {
		return value.IntVal(0), nil
	}

	if a == math.MinInt64 && b == -1 {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer overflow in multiplication", "", ""})
	}
	if b == math.MinInt64 && a == -1 {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer overflow in multiplication", "", ""})
	}

	result := a * b
	if result/b != a {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer overflow in multiplication", "", ""})
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
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"/ expects 2 arguments", "", ""})
	}

	a, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"/ expects integer arguments", "", ""})
	}

	b, ok := args[1].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"/ expects integer arguments", "", ""})
	}

	// Check for division by zero
	if b == 0 {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"division by zero", "", ""})
	}

	// Check for overflow: MinInt64 / -1 overflows
	if a == math.MinInt64 && b == -1 {
		return value.NoneVal(), verror.NewMathError(verror.ErrIDDivByZero, [3]string{"integer overflow in division", "", ""})
	}

	return value.IntVal(a / b), nil
}
