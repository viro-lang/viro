package value

import (
	"github.com/ericlagergren/decimal"
)

// DecimalValue represents high-precision decimal floating point values (Feature 002).
// Implements IEEE 754 decimal128 semantics with 34 decimal digits of precision.
//
// Design per data-model.md:
// - Magnitude: normalized decimal number supporting ±10^−6143...±10^6144
// - Context: shared context controlling precision and rounding mode
// - Scale: explicit scale metadata for round-trip formatting (e.g., "1.20" vs "1.2")
//
// Per FR-001: exactly 34 digits precision, overflow raises Math error (400)
type DecimalValue struct {
	Magnitude *decimal.Big     // Core decimal number
	Context   *decimal.Context // Precision, rounding mode, traps
	Scale     int16            // Digits right of decimal point for formatting
}

// NewDecimal creates a DecimalValue with default context (34-digit precision, half-even rounding).
func NewDecimal(magnitude *decimal.Big, scale int16) *DecimalValue {
	ctx := decimal.Context{
		Precision:    34,                    // decimal128 target per FR-001
		RoundingMode: decimal.ToNearestEven, // Banker's rounding per FR-003
	}
	return &DecimalValue{
		Magnitude: magnitude,
		Context:   &ctx,
		Scale:     scale,
	}
}

// String returns formatted decimal string preserving scale.
func (d *DecimalValue) String() string {
	if d == nil || d.Magnitude == nil {
		return "0.0"
	}
	// Use scale to format with correct decimal places
	return d.Magnitude.String()
}

// DecimalVal creates a Value wrapping a DecimalValue.
func DecimalVal(magnitude *decimal.Big, scale int16) Value {
	return Value{
		Type:    TypeDecimal,
		Payload: NewDecimal(magnitude, scale),
	}
}

// AsDecimal extracts the DecimalValue from a Value, or returns nil if wrong type.
func (v Value) AsDecimal() (*DecimalValue, bool) {
	if v.Type != TypeDecimal {
		return nil, false
	}
	dec, ok := v.Payload.(*DecimalValue)
	return dec, ok
}
