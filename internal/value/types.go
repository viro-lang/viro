// Package value implements the core value types for the Viro interpreter.
// All data in Viro is represented as tagged union values with type discrimination.
package value

import "github.com/marcin-radoszewski/viro/internal/core"

// ValueType identifies the runtime type of a Value.
// Uses uint8 for compact representation (11 types fit in 8 bits).

// Value type constants define all supported data types in Viro.
// These define the Viro type system with Viro-specific additions (Paren).
const (
	TypeNone     core.ValueType = iota // Represents absence of value (nil/null)
	TypeLogic                          // Boolean true/false
	TypeInteger                        // 64-bit signed integer
	TypeString                         // UTF-8 character sequence
	TypeWord                           // Symbol identifier (evaluates to bound value)
	TypeSetWord                        // Assignment symbol (x: value)
	TypeGetWord                        // Fetch symbol (evaluates without evaluation)
	TypeLitWord                        // Quoted symbol (returns word itself)
	TypeBlock                          // Series of values (deferred evaluation)
	TypeParen                          // Series of values (immediate evaluation)
	TypeFunction                       // Executable function (native or user-defined)

	// Feature 002: Deferred Language Capabilities
	TypeDecimal  // IEEE 754 decimal128 high-precision decimal
	TypeObject   // Object instance with frame-based fields
	TypePort     // I/O port abstraction (file, TCP, HTTP)
	TypePath     // Path expression (transient evaluation type)
	TypeGetPath  // Get-path expression (transient evaluation type)
	TypeSetPath  // Set-path expression (transient evaluation type)
	TypeDatatype // Datatype literal (e.g., object!, integer!)
	TypeBinary   // Raw byte sequence
	
	// Feature 030: Parse Dialect
	TypeBitset // Character set for parse dialect (charset)
)

// TypeToString returns the type name for debugging and error messages.
func TypeToString(t core.ValueType) string {
	switch t {
	case TypeNone:
		return "none!"
	case TypeLogic:
		return "logic!"
	case TypeInteger:
		return "integer!"
	case TypeString:
		return "string!"
	case TypeWord:
		return "word!"
	case TypeSetWord:
		return "set-word!"
	case TypeGetWord:
		return "get-word!"
	case TypeLitWord:
		return "lit-word!"
	case TypeBlock:
		return "block!"
	case TypeParen:
		return "paren!"
	case TypeFunction:
		return "function!"
	case TypeDecimal:
		return "decimal!"
	case TypeObject:
		return "object!"
	case TypePort:
		return "port!"
	case TypePath:
		return "path!"
	case TypeGetPath:
		return "get-path!"
	case TypeSetPath:
		return "set-path!"
	case TypeDatatype:
		return "datatype!"
	case TypeBinary:
		return "binary!"
	case TypeBitset:
		return "bitset!"
	default:
		return "unknown!"
	}
}

func IsWord(t core.ValueType) bool {
	return t == TypeWord || t == TypeSetWord || t == TypeGetWord || t == TypeLitWord
}

// IsSeries returns true if the type supports series operations.
func IsSeries(t core.ValueType) bool {
	return t == TypeBlock || t == TypeParen || t == TypeString || t == TypeBinary
}
