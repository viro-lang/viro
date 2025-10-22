// Package value defines the core value types for the Viro interpreter.
//
// All data in Viro is represented using the Value type, which is a type-tagged
// union. The Type field discriminates between different value types, and the
// Payload field holds the type-specific data.
//
// Value types:
//   - None: Represents absence of value
//   - Logic: Boolean true/false
//   - Integer: 64-bit signed integers
//   - String: Character sequences (runes)
//   - Word types: word, set-word, get-word, lit-word
//   - Block: Series of values (deferred evaluation)
//   - Paren: Series of values (immediate evaluation)
//   - Function: Native or user-defined functions
//
// Constructor functions (IntVal, StrVal, etc.) provide type-safe value creation.
// Type assertion helpers (AsInteger, AsString, etc.) enable safe payload access.
package value

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// Value is the universal data representation in Viro.
// All data (literals, words, blocks, functions) are represented as Values
// with a type tag and type-specific payload.
//
// Design per Constitution Principle III: Type Dispatch Fidelity
// - Type tag enables efficient type-based dispatch during evaluation
// - Payload is type-erased interface{} (discriminated union pattern)
// - Constructor functions ensure type/payload consistency
type Value struct {
	Type    core.ValueType // Discriminator identifying value type
	Payload any            // Type-specific data (int64, string, *BlockValue, etc.)
}

func (v Value) GetType() core.ValueType {
	return v.Type
}

func (v Value) GetPayload() any {
	return v.Payload
}

// String returns a string representation of the value for debugging and display.
func (v Value) String() string {
	switch v.Type {
	case TypeNone:
		return "none"
	case TypeLogic:
		if v.Payload.(bool) {
			return "true"
		}
		return "false"
	case TypeInteger:
		return fmt.Sprintf("%d", v.Payload.(int64))
	case TypeString:
		if str, ok := v.Payload.(*StringValue); ok {
			return fmt.Sprintf(`"%s"`, str.String())
		}
		return fmt.Sprintf(`"%v"`, v.Payload)
	case TypeWord:
		return v.Payload.(string)
	case TypeSetWord:
		return v.Payload.(string) + ":"
	case TypeGetWord:
		return ":" + v.Payload.(string)
	case TypeLitWord:
		return "'" + v.Payload.(string)
	case TypeBlock:
		if blk, ok := v.Payload.(*BlockValue); ok {
			return blk.String()
		}
		return "[...]"
	case TypeParen:
		if blk, ok := v.Payload.(*BlockValue); ok {
			return "(" + blk.StringElements() + ")"
		}
		return "(...)"
	case TypeFunction:
		if fn, ok := v.Payload.(*FunctionValue); ok {
			return fn.String()
		}
		return "function"
	case TypeDecimal:
		if dec, ok := v.Payload.(*DecimalValue); ok {
			return dec.String()
		}
		return "0.0"
	case TypeObject:
		if obj, ok := v.Payload.(*ObjectInstance); ok {
			return obj.String()
		}
		return "object[]"
	case TypePort:
		if port, ok := v.Payload.(*Port); ok {
			return port.String()
		}
		return "port[closed]"
	case TypePath:
		if path, ok := v.Payload.(*PathExpression); ok {
			return path.String()
		}
		return "path[]"
	case TypeDatatype:
		if name, ok := v.Payload.(string); ok {
			return name
		}
		return "datatype!"
	case TypeBinary:
		if bin, ok := v.Payload.(*BinaryValue); ok {
			return bin.String()
		}
		return "#{...}"
	default:
		return fmt.Sprintf("<%s>", TypeToString(v.Type))
	}
}

// Equals performs deep equality comparison between two values.
// Used for the = native function and testing.
func (v Value) Equals(other core.Value) bool {
	if v.Type != other.GetType() {
		return false
	}

	switch v.Type {
	case TypeNone:
		return true // all none values are equal
	case TypeLogic:
		return v.Payload.(bool) == other.GetPayload().(bool)
	case TypeInteger:
		return v.Payload.(int64) == other.GetPayload().(int64)
	case TypeString:
		vStr, vOk := v.Payload.(*StringValue)
		oStr, oOk := other.GetPayload().(*StringValue)
		if !vOk || !oOk {
			return false
		}
		return vStr.Equals(oStr)
	case TypeWord, TypeSetWord, TypeGetWord, TypeLitWord:
		return v.Payload.(string) == other.GetPayload().(string)
	case TypeDatatype:
		return v.Payload.(string) == other.GetPayload().(string)
	case TypeBlock, TypeParen:
		vBlk, vOk := v.Payload.(*BlockValue)
		oBlk, oOk := other.GetPayload().(*BlockValue)
		if !vOk || !oOk {
			return false
		}
		return vBlk.Equals(oBlk)
	case TypeFunction:
		// Functions compared by identity (pointer equality)
		return v.Payload == other.GetPayload()
	case TypeBinary:
		vBin, vOk := v.Payload.(*BinaryValue)
		oBin, oOk := other.GetPayload().(*BinaryValue)
		if !vOk || !oOk {
			return false
		}
		return vBin.Equals(oBin)
	default:
		return false
	}
}

// Constructor functions ensure type/payload consistency.
// These are the ONLY way to create values (no direct struct construction).

// NoneVal creates a none value (represents absence/null).
func NoneVal() core.Value {
	return Value{Type: TypeNone, Payload: nil}
}

// LogicVal creates a logic value (true/false).
func LogicVal(b bool) Value {
	return Value{Type: TypeLogic, Payload: b}
}

// IntVal creates an integer value (64-bit signed).
func IntVal(i int64) Value {
	return Value{Type: TypeInteger, Payload: i}
}

// StrVal creates a string value from a Go string.
// Converts to StringValue (rune array) for series operations.
func StrVal(s string) Value {
	return Value{Type: TypeString, Payload: NewStringValue(s)}
}

// WordVal creates a word value (symbol that evaluates to bound value).
func WordVal(symbol string) Value {
	return Value{Type: TypeWord, Payload: symbol}
}

// SetWordVal creates a set-word value (assignment symbol).
func SetWordVal(symbol string) Value {
	return Value{Type: TypeSetWord, Payload: symbol}
}

// GetWordVal creates a get-word value (fetch without evaluation).
func GetWordVal(symbol string) Value {
	return Value{Type: TypeGetWord, Payload: symbol}
}

// LitWordVal creates a lit-word value (quoted symbol).
func LitWordVal(symbol string) Value {
	return Value{Type: TypeLitWord, Payload: symbol}
}

// BlockVal creates a block value (deferred evaluation series).
func BlockVal(elements []core.Value) Value {
	return Value{Type: TypeBlock, Payload: NewBlockValue(elements)}
}

// ParenVal creates a paren value (immediate evaluation series).
func ParenVal(elements []core.Value) Value {
	return Value{Type: TypeParen, Payload: NewBlockValue(elements)}
}

// FuncVal creates a function value.
func FuncVal(fn *FunctionValue) Value {
	return Value{Type: TypeFunction, Payload: fn}
}

// DatatypeVal creates a datatype value (e.g., object!, integer!).
func DatatypeVal(name string) Value {
	return Value{Type: TypeDatatype, Payload: name}
}

// BinaryVal creates a binary value from a byte slice.
func BinaryVal(data []byte) Value {
	return Value{Type: TypeBinary, Payload: NewBinaryValue(data)}
}

// Type assertion helpers for safe payload extraction.
// Return (value, true) on success or (zero-value, false) on type mismatch.

// AsInteger extracts integer payload if value is TypeInteger.
func AsInteger(v core.Value) (int64, bool) {
	if v.GetType() != TypeInteger {
		return 0, false
	}
	i, ok := v.GetPayload().(int64)
	return i, ok
}

// AsLogic extracts boolean payload if value is TypeLogic.
func AsLogic(v core.Value) (bool, bool) {
	if v.GetType() != TypeLogic {
		return false, false
	}
	b, ok := v.GetPayload().(bool)
	return b, ok
}

// AsString extracts StringValue if value is TypeString.
func AsString(v core.Value) (*StringValue, bool) {
	if v.GetType() != TypeString {
		return nil, false
	}
	s, ok := v.GetPayload().(*StringValue)
	return s, ok
}

// AsWord extracts symbol string if value is any word type.
func AsWord(v core.Value) (string, bool) {
	if !IsWord(v.GetType()) {
		return "", false
	}
	sym, ok := v.GetPayload().(string)
	return sym, ok
}

func AsBlock(v core.Value) (*BlockValue, bool) {
	if v.GetType() != TypeBlock && v.GetType() != TypeParen {
		return nil, false
	}
	blk, ok := v.GetPayload().(*BlockValue)
	return blk, ok
}

func AsFunction(v core.Value) (*FunctionValue, bool) {
	if v.GetType() != TypeFunction {
		return nil, false
	}
	fn, ok := v.GetPayload().(*FunctionValue)
	return fn, ok
}

func AsDatatype(v core.Value) (string, bool) {
	if v.GetType() != TypeDatatype {
		return "", false
	}
	name, ok := v.GetPayload().(string)
	return name, ok
}

// AsBinary extracts BinaryValue if value is TypeBinary.
func AsBinary(v core.Value) (*BinaryValue, bool) {
	if v.GetType() != TypeBinary {
		return nil, false
	}
	b, ok := v.GetPayload().(*BinaryValue)
	return b, ok
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
