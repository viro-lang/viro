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

import "fmt"

// Value is the universal data representation in Viro.
// All data (literals, words, blocks, functions) are represented as Values
// with a type tag and type-specific payload.
//
// Design per Constitution Principle III: Type Dispatch Fidelity
// - Type tag enables efficient type-based dispatch during evaluation
// - Payload is type-erased interface{} (discriminated union pattern)
// - Constructor functions ensure type/payload consistency
type Value struct {
	Type    ValueType   // Discriminator identifying value type
	Payload interface{} // Type-specific data (int64, string, *BlockValue, etc.)
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
	case TypeAction:
		if action, ok := v.Payload.(*ActionValue); ok {
			return action.String()
		}
		return "action"
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
	default:
		return fmt.Sprintf("<%s>", v.Type)
	}
}

// Equals performs deep equality comparison between two values.
// Used for the = native function and testing.
func (v Value) Equals(other Value) bool {
	if v.Type != other.Type {
		return false
	}

	switch v.Type {
	case TypeNone:
		return true // all none values are equal
	case TypeLogic:
		return v.Payload.(bool) == other.Payload.(bool)
	case TypeInteger:
		return v.Payload.(int64) == other.Payload.(int64)
	case TypeString:
		vStr, vOk := v.Payload.(*StringValue)
		oStr, oOk := other.Payload.(*StringValue)
		if !vOk || !oOk {
			return false
		}
		return vStr.Equals(oStr)
	case TypeWord, TypeSetWord, TypeGetWord, TypeLitWord:
		return v.Payload.(string) == other.Payload.(string)
	case TypeDatatype:
		return v.Payload.(string) == other.Payload.(string)
	case TypeBlock, TypeParen:
		vBlk, vOk := v.Payload.(*BlockValue)
		oBlk, oOk := other.Payload.(*BlockValue)
		if !vOk || !oOk {
			return false
		}
		return vBlk.Equals(oBlk)
	case TypeFunction:
		// Functions compared by identity (pointer equality)
		return v.Payload == other.Payload
	case TypeAction:
		// Actions compared by identity (pointer equality)
		return v.Payload == other.Payload
	default:
		return false
	}
}

// Constructor functions ensure type/payload consistency.
// These are the ONLY way to create values (no direct struct construction).

// NoneVal creates a none value (represents absence/null).
func NoneVal() Value {
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
func BlockVal(elements []Value) Value {
	return Value{Type: TypeBlock, Payload: NewBlockValue(elements)}
}

// ParenVal creates a paren value (immediate evaluation series).
func ParenVal(elements []Value) Value {
	return Value{Type: TypeParen, Payload: NewBlockValue(elements)}
}

// FuncVal creates a function value.
func FuncVal(fn *FunctionValue) Value {
	return Value{Type: TypeFunction, Payload: fn}
}

// ActionVal creates an action value.
func ActionVal(action *ActionValue) Value {
	return Value{Type: TypeAction, Payload: action}
}

// DatatypeVal creates a datatype value (e.g., object!, integer!).
func DatatypeVal(name string) Value {
	return Value{Type: TypeDatatype, Payload: name}
}

// Type assertion helpers for safe payload extraction.
// Return (value, true) on success or (zero-value, false) on type mismatch.

// AsInteger extracts integer payload if value is TypeInteger.
func (v Value) AsInteger() (int64, bool) {
	if v.Type != TypeInteger {
		return 0, false
	}
	i, ok := v.Payload.(int64)
	return i, ok
}

// AsLogic extracts boolean payload if value is TypeLogic.
func (v Value) AsLogic() (bool, bool) {
	if v.Type != TypeLogic {
		return false, false
	}
	b, ok := v.Payload.(bool)
	return b, ok
}

// AsString extracts StringValue if value is TypeString.
func (v Value) AsString() (*StringValue, bool) {
	if v.Type != TypeString {
		return nil, false
	}
	s, ok := v.Payload.(*StringValue)
	return s, ok
}

// AsWord extracts symbol string if value is any word type.
func (v Value) AsWord() (string, bool) {
	if !v.Type.IsWord() {
		return "", false
	}
	sym, ok := v.Payload.(string)
	return sym, ok
}

// AsBlock extracts BlockValue if value is TypeBlock or TypeParen.
func (v Value) AsBlock() (*BlockValue, bool) {
	if v.Type != TypeBlock && v.Type != TypeParen {
		return nil, false
	}
	blk, ok := v.Payload.(*BlockValue)
	return blk, ok
}

// AsFunction extracts FunctionValue if value is TypeFunction.
func (v Value) AsFunction() (*FunctionValue, bool) {
	if v.Type != TypeFunction {
		return nil, false
	}
	fn, ok := v.Payload.(*FunctionValue)
	return fn, ok
}

// AsAction extracts ActionValue if value is TypeAction.
func (v Value) AsAction() (*ActionValue, bool) {
	if v.Type != TypeAction {
		return nil, false
	}
	action, ok := v.Payload.(*ActionValue)
	return action, ok
}

// AsDatatype extracts datatype name if value is TypeDatatype.
func (v Value) AsDatatype() (string, bool) {
	if v.Type != TypeDatatype {
		return "", false
	}
	name, ok := v.Payload.(string)
	return name, ok
}

// IsTruthy returns true if value is considered "true" in conditional contexts.
// Per contracts/control-flow.md: false and none are falsy, all others truthy.
// Note: 0, "", [] are truthy (unlike some languages).
func (v Value) IsTruthy() bool {
	if v.Type == TypeNone {
		return false
	}
	if v.Type == TypeLogic {
		b, _ := v.AsLogic()
		return b
	}
	// All other values (including 0, "", []) are truthy
	return true
}
