package value

// Word types in Viro follow Viro semantics with symbol-based identification.
// All word variants store a symbol string as payload.
//
// Word types (distinguished by Value.Type):
// - Word: Evaluates to bound value in context
// - SetWord: Assignment operator (symbol: value)
// - GetWord: Fetches value without evaluation (:symbol)
// - LitWord: Returns word itself ('symbol)
//
// Design per data-model.md:
// - Case-sensitive symbols (per spec clarification)
// - Binding resolved at evaluation time via Frame lookup
// - Symbol stored as string payload (not separate struct needed)

// No separate struct needed for word types - they use string payload directly.
// Constructor functions already defined in value.go:
// - WordVal(symbol string) Value
// - SetWordVal(symbol string) Value
// - GetWordVal(symbol string) Value
// - LitWordVal(symbol string) Value

// ValidWordSymbol checks if a string is a valid word symbol.
// Word symbols can contain letters, digits, hyphens, and some special characters.
// Cannot be empty or start with a digit.
func ValidWordSymbol(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Cannot start with digit
	if s[0] >= '0' && s[0] <= '9' {
		return false
	}
	// Valid characters: letters, digits, hyphens, underscores, question marks
	for _, r := range s {
		if !isWordChar(r) {
			return false
		}
	}
	return true
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '?' || r == '!'
}
