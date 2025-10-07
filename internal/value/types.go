// Package value implements the core value types for the Viro interpreter.
// All data in Viro is represented as tagged union values with type discrimination.
package value

// ValueType identifies the runtime type of a Value.
// Uses uint8 for compact representation (11 types fit in 8 bits).
type ValueType uint8

// Value type constants define all supported data types in Viro.
// These align with REBOL's type system with Viro-specific additions (Paren).
const (
	TypeNone     ValueType = iota // Represents absence of value (nil/null)
	TypeLogic                     // Boolean true/false
	TypeInteger                   // 64-bit signed integer
	TypeString                    // UTF-8 character sequence
	TypeWord                      // Symbol identifier (evaluates to bound value)
	TypeSetWord                   // Assignment symbol (x: value)
	TypeGetWord                   // Fetch symbol (evaluates without evaluation)
	TypeLitWord                   // Quoted symbol (returns word itself)
	TypeBlock                     // Series of values (deferred evaluation)
	TypeParen                     // Series of values (immediate evaluation)
	TypeFunction                  // Executable function (native or user-defined)
)

// String returns the type name for debugging and error messages.
func (t ValueType) String() string {
	switch t {
	case TypeNone:
		return "none"
	case TypeLogic:
		return "logic"
	case TypeInteger:
		return "integer"
	case TypeString:
		return "string"
	case TypeWord:
		return "word"
	case TypeSetWord:
		return "set-word"
	case TypeGetWord:
		return "get-word"
	case TypeLitWord:
		return "lit-word"
	case TypeBlock:
		return "block"
	case TypeParen:
		return "paren"
	case TypeFunction:
		return "function"
	default:
		return "unknown"
	}
}

// IsWord returns true if the type is any word variant.
func (t ValueType) IsWord() bool {
	return t == TypeWord || t == TypeSetWord || t == TypeGetWord || t == TypeLitWord
}

// IsSeries returns true if the type supports series operations.
func (t ValueType) IsSeries() bool {
	return t == TypeBlock || t == TypeParen || t == TypeString
}
