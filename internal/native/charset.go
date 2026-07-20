package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// NativeCharset creates a bitset (charset) from a specification.
// Accepts: string, block of characters/ranges, or another bitset.
func NativeCharset(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("charset", 1, len(args))
	}

	arg := args[0]

	// Handle string argument - create bitset from all characters
	if strVal, ok := value.AsStringValue(arg); ok {
		return value.NewBitsetFromString(strVal.String()), nil
	}

	// Handle bitset argument - clone it
	if bsVal, ok := value.AsBitsetValue(arg); ok {
		return bsVal.Clone(), nil
	}

	// Handle block argument - parse charset specification
	if blockVal, ok := value.AsBlockValue(arg); ok {
		return parseCharsetBlock(blockVal.Elements, eval)
	}

	return value.NewNoneVal(), typeError("charset", "string! block! bitset!", arg)
}

// parseCharsetBlock parses a block specification for charset.
// Supports:
//   - Individual characters: #"a"
//   - Strings: "abc"
//   - Ranges: [#"a" - #"z"]
//   - Negation: not [...]
//   - Union: multiple specs are unioned
func parseCharsetBlock(elements []core.Value, eval core.Evaluator) (core.Value, error) {
	result := value.NewBitsetValue()

	i := 0
	for i < len(elements) {
		elem := elements[i]

		// Handle "not" for complement
		if wordVal, ok := value.AsWordValue(elem); ok && wordVal == "not" {
			i++
			if i >= len(elements) {
				return value.NewNoneVal(), verror.NewScriptError("invalid-arg", [3]string{"charset", "'not' requires an argument", ""})
			}

			// Get the bitset to complement
			nextElem := elements[i]
			var toComplement *value.BitsetValue

			if strVal, ok := value.AsStringValue(nextElem); ok {
				toComplement = value.NewBitsetFromString(strVal.String())
			} else if bsVal, ok := value.AsBitsetValue(nextElem); ok {
				toComplement = bsVal
			} else if blockVal, ok := value.AsBlockValue(nextElem); ok {
				// Recursively parse the block
				bs, err := parseCharsetBlock(blockVal.Elements, eval)
				if err != nil {
					return value.NewNoneVal(), err
				}
				toComplement, _ = value.AsBitsetValue(bs)
			} else {
				return value.NewNoneVal(), verror.NewScriptError("invalid-arg", [3]string{"charset", "'not' argument must be string, block, or bitset", ""})
			}

			// Complement and union with result
			comp := toComplement.Complement()
			result = result.Union(comp)
			i++
			continue
		}

		// Handle character literal #"x"
		if strVal, ok := value.AsStringValue(elem); ok {
			str := strVal.String()
			if len(str) > 0 && str[0] == '#' && len(str) >= 3 {
				// This is a character literal like #"a"
				char := rune(str[2])
				result.Set(char)
				i++
				continue
			}
			// Regular string - add all characters
			for _, r := range str {
				result.Set(r)
			}
			i++
			continue
		}

		// Handle integer (character code)
		if intVal, ok := value.AsIntValue(elem); ok {
			result.Set(rune(intVal))
			i++
			continue
		}

		// Handle block for range [#"a" - #"z"]
		if blockVal, ok := value.AsBlockValue(elem); ok {
			if len(blockVal.Elements) == 3 {
				// Check if it's a range: [start - end]
				start, startOK := getCharFromValue(blockVal.Elements[0])
				dash, dashOK := value.AsWordValue(blockVal.Elements[1])
				end, endOK := getCharFromValue(blockVal.Elements[2])

				if startOK && dashOK && dash == "-" && endOK {
					// Add range
					for r := start; r <= end; r++ {
						result.Set(r)
					}
					i++
					continue
				}
			}
			// Not a range, parse as nested block
			bs, err := parseCharsetBlock(blockVal.Elements, eval)
			if err != nil {
				return value.NewNoneVal(), err
			}
			if bsVal, ok := value.AsBitsetValue(bs); ok {
				result = result.Union(bsVal)
			}
			i++
			continue
		}

		// Handle existing bitset
		if bsVal, ok := value.AsBitsetValue(elem); ok {
			result = result.Union(bsVal)
			i++
			continue
		}

		// Unknown element type
		return value.NewNoneVal(), verror.NewScriptError("invalid-arg", [3]string{"charset", "invalid charset element: " + value.TypeToString(elem.GetType()), ""})
	}

	return result, nil
}

// getCharFromValue extracts a character from a value (string literal or integer).
func getCharFromValue(v core.Value) (rune, bool) {
	// Handle character literal #"x"
	if strVal, ok := value.AsStringValue(v); ok {
		str := strVal.String()
		if len(str) > 0 && str[0] == '#' && len(str) >= 3 {
			return rune(str[2]), true
		}
		if len(str) == 1 {
			return rune(str[0]), true
		}
	}

	// Handle integer (character code)
	if intVal, ok := value.AsIntValue(v); ok {
		return rune(intVal), true
	}

	return 0, false
}
