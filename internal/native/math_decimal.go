// Package native provides advanced math operations for decimal values.
// This file implements transcendental and rounding functions for decimal! type.
package native

import (
	"math"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Decimal constructor native - creates decimal from integer, decimal, or string
func DecimalConstructor(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("decimal-constructor-arity", [3]string{"1", "", ""})
	}

	arg := args[0]

	switch arg.GetType() {
	case value.TypeInteger:
		// Convert integer to decimal with scale 0
		if i, ok := value.AsIntValue(arg); ok {
			d := decimal.New(i, 0)
			return value.DecimalVal(d, 0), nil
		}
		return value.NewNoneVal(), verror.NewMathError("invalid-integer", [3]string{"", "", ""})

	case value.TypeDecimal:
		// Already a decimal, return as-is
		return arg, nil

	case value.TypeString:
		// Parse string to decimal
		if str, ok := value.AsStringValue(arg); ok {
			goStr := str.String()
			d := new(decimal.Big)
			_, ok := d.SetString(goStr)
			if !ok {
				return value.NewNoneVal(), verror.NewMathError("invalid-decimal-string", [3]string{goStr, "", ""})
			}
			// Calculate scale from string representation
			scale := int16(0)
			if idx := findDecimalPoint(goStr); idx >= 0 {
				scale = int16(len(goStr) - idx - 1)
			}
			return value.DecimalVal(d, scale), nil
		}
		return value.NewNoneVal(), verror.NewMathError("invalid-string", [3]string{"", "", ""})

	default:
		return value.NewNoneVal(), verror.NewMathError("decimal-invalid-type", [3]string{value.TypeToString(arg.GetType()), "", ""})
	}
}

// findDecimalPoint finds the position of decimal point in a numeric string
func findDecimalPoint(s string) int {
	for i, ch := range s {
		if ch == '.' {
			return i
		}
	}
	return -1
}

// Pow computes base^exponent for decimal values
func Pow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewMathError("pow-arity", [3]string{"2", "", ""})
	}

	base := promoteToDecimal(args[0], nil, nil)
	exp := promoteToDecimal(args[1], nil, nil)

	if base == nil || exp == nil {
		return value.NewNoneVal(), verror.NewMathError("pow-invalid-type", [3]string{value.TypeToString(args[0].GetType()), value.TypeToString(args[1].GetType()), ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Pow(result, base, exp)

	// Check for overflow/underflow
	if result.IsInf(0) {
		return value.NewNoneVal(), verror.NewMathError("pow-overflow", [3]string{"", "", ""})
	}

	return value.DecimalVal(result, 2), nil
}

// Sqrt computes square root of a decimal value
func Sqrt(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("sqrt-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("sqrt-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Domain check: negative values not allowed
	if val.Sign() < 0 {
		return value.NewNoneVal(), verror.NewMathError(verror.ErrIDSqrtNegative, [3]string{val.String(), "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Sqrt(result, val)

	return value.DecimalVal(result, 2), nil
}

// Exp computes e^x for decimal values
func Exp(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("exp-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("exp-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Exp(result, val)

	// Check for overflow
	if result.IsInf(0) {
		return value.NewNoneVal(), verror.NewMathError(verror.ErrIDExpOverflow, [3]string{val.String(), "", ""})
	}

	return value.DecimalVal(result, 2), nil
}

// Log computes natural logarithm for decimal values
func Log(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("log-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("log-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Domain check: must be positive
	if val.Sign() <= 0 {
		return value.NewNoneVal(), verror.NewMathError(verror.ErrIDLogDomain, [3]string{val.String(), "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Log(result, val)

	return value.DecimalVal(result, 2), nil
}

// Log10 computes base-10 logarithm for decimal values
func Log10(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("log-10-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("log-10-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Domain check: must be positive
	if val.Sign() <= 0 {
		return value.NewNoneVal(), verror.NewMathError("log-10-domain", [3]string{val.String(), "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Log10(result, val)

	return value.DecimalVal(result, 2), nil
}

// Sin computes sine for decimal values (input in radians)
func Sin(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("sin-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("sin-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Convert to float64 for trig functions
	f, _ := val.Float64()
	result := math.Sin(f)

	d := decimal.New(int64(result*1e10), -10)
	return value.DecimalVal(d, 10), nil
}

// Cos computes cosine for decimal values (input in radians)
func Cos(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("cos-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("cos-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Convert to float64 for trig functions
	f, _ := val.Float64()
	result := math.Cos(f)

	d := decimal.New(int64(result*1e10), -10)
	return value.DecimalVal(d, 10), nil
}

// Tan computes tangent for decimal values (input in radians)
func Tan(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("tan-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("tan-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Convert to float64 for trig functions
	f, _ := val.Float64()
	result := math.Tan(f)

	d := decimal.New(int64(result*1e10), -10)
	return value.DecimalVal(d, 10), nil
}

// Asin computes arcsine for decimal values (result in radians)
func Asin(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("asin-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("asin-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Domain check: [-1, 1]
	f, _ := val.Float64()
	if f < -1.0 || f > 1.0 {
		return value.NewNoneVal(), verror.NewMathError(verror.ErrIDAsinDomain, [3]string{val.String(), "", ""})
	}

	result := math.Asin(f)
	d := decimal.New(int64(result*1e10), -10)
	return value.DecimalVal(d, 10), nil
}

// Acos computes arccosine for decimal values (result in radians)
func Acos(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("acos-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("acos-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	// Domain check: [-1, 1]
	f, _ := val.Float64()
	if f < -1.0 || f > 1.0 {
		return value.NewNoneVal(), verror.NewMathError(verror.ErrIDAcosDomain, [3]string{val.String(), "", ""})
	}

	result := math.Acos(f)
	d := decimal.New(int64(result*1e10), -10)
	return value.DecimalVal(d, 10), nil
}

// Atan computes arctangent for decimal values (result in radians)
func Atan(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("atan-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("atan-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	f, _ := val.Float64()
	result := math.Atan(f)

	d := decimal.New(int64(result*1e10), -10)
	return value.DecimalVal(d, 10), nil
}

// Round rounds a decimal to specified places with optional rounding mode
func Round(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) < 1 || len(args) > 3 {
		return value.NewNoneVal(), verror.NewMathError("round-arity", [3]string{"1-3", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("round-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	places := int32(0)
	if len(args) >= 2 && args[1].GetType() == value.TypeInteger {
		if i, ok := value.AsIntValue(args[1]); ok {
			places = int32(i)
		}
	}

	// For now, use default half-even rounding

	result := new(decimal.Big)
	result.Copy(val)
	result.Round(int(places))

	return value.DecimalVal(result, int16(places)), nil
}

// Ceil returns the smallest integer >= the decimal value
func Ceil(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("ceil-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("ceil-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Ceil(result, val)

	return value.DecimalVal(result, 0), nil
}

// Floor returns the largest integer <= the decimal value
func Floor(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("floor-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("floor-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	ctx := decimal.Context128
	result := new(decimal.Big)
	ctx.Floor(result, val)

	return value.DecimalVal(result, 0), nil
}

// Truncate returns the integer part of a decimal value
func Truncate(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewMathError("truncate-arity", [3]string{"1", "", ""})
	}

	val := promoteToDecimal(args[0], nil, nil)
	if val == nil {
		return value.NewNoneVal(), verror.NewMathError("truncate-invalid-type", [3]string{value.TypeToString(args[0].GetType()), "", ""})
	}

	result := new(decimal.Big)
	result.Copy(val)
	result.RoundToInt()

	return value.DecimalVal(result, 0), nil
}

// promoteToDecimal converts integer or decimal values to *decimal.Big
func promoteToDecimal(v core.Value, _ map[string]core.Value, _ core.Evaluator) *decimal.Big {
	switch v.GetType() {
	case value.TypeDecimal:
		if dec, ok := value.AsDecimal(v); ok && dec != nil {
			return dec.Magnitude
		}
		return nil
	case value.TypeInteger:
		if i, ok := value.AsIntValue(v); ok {
			return decimal.New(i, 0)
		}
		return nil
	default:
		return nil
	}
}
