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
// Constructor functions (IntVal, StrVal, etc.) provide type-safe value creation.
// Type assertion helpers (AsInteger, AsString, etc.) enable safe type extraction.
package value

import (
	"github.com/marcin-radoszewski/viro/internal/core"
)

// Constructor functions ensure type/payload consistency.
// These are the ONLY way to create values (no direct struct construction).

// NoneVal creates a none value (represents absence/null).
func NoneVal() core.Value {
	return NewNoneVal()
}

// LogicVal creates a logic value (true/false).
func LogicVal(b bool) core.Value {
	return NewLogicVal(b)
}

// IntVal creates an integer value (64-bit signed).
func IntVal(i int64) core.Value {
	return NewIntVal(i)
}

// StrVal creates a string value from a Go string.
// Converts to StringValue (rune array) for series operations.
func StrVal(s string) core.Value {
	return NewStrVal(s)
}

// WordVal creates a word value (symbol that evaluates to bound value).
func WordVal(symbol string) core.Value {
	return NewWordVal(symbol)
}

// SetWordVal creates a set-word value (assignment symbol).
func SetWordVal(symbol string) core.Value {
	return NewSetWordVal(symbol)
}

// GetWordVal creates a get-word value (fetch without evaluation).
func GetWordVal(symbol string) core.Value {
	return NewGetWordVal(symbol)
}

// LitWordVal creates a lit-word value (quoted symbol).
func LitWordVal(symbol string) core.Value {
	return NewLitWordVal(symbol)
}

// BlockVal creates a block value (deferred evaluation series).
func BlockVal(elements []core.Value) core.Value {
	return NewBlockVal(elements)
}

// ParenVal creates a paren value (immediate evaluation series).
func ParenVal(elements []core.Value) core.Value {
	return NewParenVal(elements)
}

// FuncVal creates a function value.
func FuncVal(fn *FunctionValue) core.Value {
	return NewFuncVal(fn)
}

// DatatypeVal creates a datatype value (e.g., object!, integer!).
func DatatypeVal(name string) core.Value {
	return NewDatatypeVal(name)
}

// BinaryVal creates a binary value from a byte slice.
func BinaryVal(data []byte) core.Value {
	return NewBinaryVal(data)
}

// Type assertion helpers for safe payload extraction.
// Return (value, true) on success or (zero-value, false) on type mismatch.

// AsInteger extracts integer payload if value is TypeInteger.
func AsInteger(v core.Value) (int64, bool) {
	return AsIntValue(v)
}

// AsLogic extracts boolean payload if value is TypeLogic.
func AsLogic(v core.Value) (bool, bool) {
	return AsLogicValue(v)
}

// AsString extracts StringValue if value is TypeString.
func AsString(v core.Value) (*StringValue, bool) {
	return AsStringValue(v)
}

// AsWord extracts symbol string if value is any word type.
func AsWord(v core.Value) (string, bool) {
	return AsWordValue(v)
}

func AsBlock(v core.Value) (*BlockValue, bool) {
	return AsBlockValue(v)
}

func AsFunction(v core.Value) (*FunctionValue, bool) {
	return AsFunctionValue(v)
}

func AsDatatype(v core.Value) (string, bool) {
	return AsDatatypeValue(v)
}

// AsBinary extracts BinaryValue if value is TypeBinary.
func AsBinary(v core.Value) (*BinaryValue, bool) {
	return AsBinaryValue(v)
}

// IsTruthy returns true if value is considered "true" in conditional contexts.
// Per contracts/control-flow.md: false and none are falsy, all others truthy.
// Note: 0, "", [] are truthy (unlike some languages).
func IsTruthy(v core.Value) bool {
	if v.GetType() == TypeNone {
		return false
	}
	if v.GetType() == TypeLogic {
		b, _ := AsLogic(v)
		return b
	}
	// All other values (including 0, "", []) are truthy
	return true
}
