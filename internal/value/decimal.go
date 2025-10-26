package value

import (
	"fmt"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
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
	return d.Mold()
}

// Mold returns the mold-formatted decimal representation.
func (d *DecimalValue) Mold() string {
	if d == nil || d.Magnitude == nil {
		return "0.0"
	}
	// Use scale to format with correct decimal places
	if f, ok := d.Magnitude.Float64(); ok {
		return fmt.Sprintf("%.*f", d.Scale, f)
	}
	// Fallback to scientific notation if conversion fails
	return d.Magnitude.String()
}

// Form returns the form-formatted decimal representation (same as mold for decimals).
func (d *DecimalValue) Form() string {
	return d.Mold()
}

func (d *DecimalValue) GetType() core.ValueType {
	return TypeDecimal
}

func (d *DecimalValue) GetPayload() any {
	return d
}

func (d *DecimalValue) Equals(other core.Value) bool {
	if other.GetType() != TypeDecimal {
		return false
	}
	otherDec, ok := other.GetPayload().(*DecimalValue)
	if !ok {
		return false
	}
	if d.Magnitude == nil && otherDec.Magnitude == nil {
		return true
	}
	if d.Magnitude == nil || otherDec.Magnitude == nil {
		return false
	}
	return d.Magnitude.Cmp(otherDec.Magnitude) == 0
}

// DecimalVal creates a Value wrapping a DecimalValue.
func DecimalVal(magnitude *decimal.Big, scale int16) Value {
	return Value{
		Type:    TypeDecimal,
		Payload: NewDecimal(magnitude, scale),
	}
}

// AsDecimal extracts the DecimalValue from a Value, or returns nil if wrong type.
func AsDecimal(v core.Value) (*DecimalValue, bool) {
	if v.GetType() != TypeDecimal {
		return nil, false
	}
	dec, ok := v.GetPayload().(*DecimalValue)
	return dec, ok
}
