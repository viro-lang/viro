// Package value defines the core value types for the Viro interpreter.
//
// All data in Viro is represented as implementations of the core.Value interface.
// Each value type implements the interface directly for maximum performance.
//
// Value types:
//   - None: Represents absence of value (NoneValue)
//   - Logic: Boolean true/false (LogicValue)
//   - Integer: 64-bit signed integers (IntValue)
//   - String: Character sequences (*StringValue)
//   - Word types: word, set-word, get-word, lit-word (WordValue, SetWordValue, etc.)
//   - Block: Series of values - deferred evaluation (*BlockValue)
//   - Paren: Series of values - immediate evaluation (*BlockValue with TypeParen)
//   - Function: Native or user-defined functions (*FunctionValue)
//   - Decimal: High-precision decimals (*DecimalValue)
//   - Binary: Raw byte sequences (*BinaryValue)
//   - Object: Object instances (*ObjectInstance)
//   - Port: I/O port abstraction (*Port)
//   - Path: Path expressions (*PathExpression)
//   - Datatype: Type literals (DatatypeValue)
//
// Constructor functions (NewIntVal, NewStrVal, etc.) provide type-safe value creation.
// Type assertion helpers (AsIntValue, AsStringValue, etc.) enable safe type extraction.
package value

import (
	"github.com/marcin-radoszewski/viro/internal/core"
)

// IsTruthy returns true if value is considered "true" in conditional contexts.
// Per contracts/control-flow.md: false and none are falsy, all others truthy.
// Note: 0, "", [] are truthy (unlike some languages).
func IsTruthy(v core.Value) bool {
	if v.GetType() == TypeNone {
		return false
	}
	if v.GetType() == TypeLogic {
		b, _ := AsLogicValue(v)
		return b
	}
	// All other values (including 0, "", []) are truthy
	return true
}
