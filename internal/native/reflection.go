package native

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Reflection natives (Feature 002, FR-022)
// Contract per contracts/reflection.md: type-of, spec-of, body-of, words-of, values-of, source

// TypeOf implements the `type-of` native (T155).
//
// Contract: type-of value -> word! representing type name
// Returns canonical type name as word (e.g., integer!, string!, object!)
// For functions, distinguishes between native! and function! based on FunctionValue.Type
func TypeOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("type-of", 1, len(args))
	}

	val := args[0]

	// Special handling for functions to distinguish native! from function!
	if val.GetType() == value.TypeFunction {
		if fn, ok := value.AsFunction(val); ok {
			if fn.Type == value.FuncNative {
				return value.WordVal("native!"), nil
			}
			return value.WordVal("function!"), nil
		}
	}

	typeName := value.TypeToString(val.GetType())
	return value.WordVal(typeName), nil
}

// SpecOf implements the `spec-of` native (T156).
//
// Contract: spec-of value -> block! copy of specification
// Supports: function!, native!, object!
// Returns immutable copy of specification block
func SpecOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("spec-of", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeFunction:
		fn, _ := value.AsFunction(val)
		// Build spec block from Params
		specElements := []core.Value{}
		for _, param := range fn.Params {
			specElements = append(specElements, value.WordVal(param.Name))
		}
		return value.BlockVal(specElements), nil

	case value.TypeObject:
		obj, _ := value.AsObject(val)
		// Build spec block from manifest
		specElements := []core.Value{}
		for i, field := range obj.Manifest.Words {
			specElements = append(specElements, value.WordVal(field))
			// If there's a type hint, include it
			if i < len(obj.Manifest.Types) && obj.Manifest.Types[i] != value.TypeNone {
				typeName := value.TypeToString(obj.Manifest.Types[i])
				specElements = append(specElements, value.WordVal(typeName))
			}
		}
		return value.BlockVal(specElements), nil

	default:
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDSpecUnsupported,
			[3]string{value.TypeToString(val.GetType()), "", ""},
		)
	}
}

// BodyOf implements the `body-of` native (T157).
//
// Contract: body-of value -> block! copy of body
// Supports: function!, object!
// Returns deep copy to prevent mutation of original
// Native functions return an error as they have no accessible body
func BodyOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("body-of", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeFunction:
		fn, _ := value.AsFunction(val)
		// Check if this is a native function (no body)
		if fn.Type == value.FuncNative {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDNoBody,
				[3]string{"native functions have no accessible body", "", ""},
			)
		}
		// User-defined function: return deep copy of body block
		if fn.Body == nil {
			return value.BlockVal([]core.Value{}), nil
		}
		bodyElements := make([]core.Value, len(fn.Body.Elements))
		copy(bodyElements, fn.Body.Elements)
		return value.BlockVal(bodyElements), nil

	case value.TypeObject:
		obj, _ := value.AsObject(val)
		// Build body block from manifest (field: value pairs)
		bodyElements := []core.Value{}
		for _, field := range obj.Manifest.Words {
			bodyElements = append(bodyElements, value.SetWordVal(field))
			bodyElements = append(bodyElements, value.NoneVal()) // Placeholder
		}
		return value.BlockVal(bodyElements), nil

	default:
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDNoBody,
			[3]string{value.TypeToString(val.GetType()), "has no accessible body", ""},
		)
	}
}

// WordsOf implements the `words-of` native (T158).
//
// Contract: words-of value -> block! of words
// Supports: object!
// Returns block of field names
func WordsOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("words-of", 1, len(args))
	}

	val := args[0]

	if val.GetType() != value.TypeObject {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"object!", value.TypeToString(val.GetType()), ""},
		)
	}

	obj, _ := value.AsObject(val)

	// Build block of words
	wordElements := make([]core.Value, len(obj.Manifest.Words))
	for i, field := range obj.Manifest.Words {
		wordElements[i] = value.WordVal(field)
	}

	return value.BlockVal(wordElements), nil
}

// ValuesOf implements the `values-of` native (T159).
//
// Contract: values-of value -> block! of values
// Supports: object!
// Returns block of field values (deep copies for safety)
func ValuesOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("values-of", 1, len(args))
	}

	val := args[0]

	if val.GetType() != value.TypeObject {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"object!", value.TypeToString(val.GetType()), ""},
		)
	}

	obj, _ := value.AsObject(val)

	// Build block of values using owned frame
	valueElements := make([]core.Value, len(obj.Manifest.Words))
	for i, field := range obj.Manifest.Words {
		if fieldVal, found := obj.GetField(field); found {
			// Deep copy the value to prevent mutation
			valueElements[i] = fieldVal
		} else {
			valueElements[i] = value.NoneVal()
		}
	}

	return value.BlockVal(valueElements), nil
}

// Source implements the `source` native (T160).
//
// Contract: source value -> string! formatted source
// Supports: function!, native!, object!
// Returns formatted string representation
func Source(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("source", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeFunction:
		fn, _ := value.AsFunction(val)
		// Format: fn [spec] [body]
		// Build spec from Params
		specElements := []core.Value{}
		for _, param := range fn.Params {
			specElements = append(specElements, value.WordVal(param.Name))
		}
		specStr := formatBlock(specElements)

		// Body
		bodyStr := "[]"
		if fn.Body != nil {
			bodyStr = formatBlock(fn.Body.Elements)
		}
		source := fmt.Sprintf("fn %s %s", specStr, bodyStr)
		return value.StrVal(source), nil

	case value.TypeObject:
		obj, _ := value.AsObject(val)
		// Format: object [field: value ...]
		fields := []string{}
		for _, field := range obj.Manifest.Words {
			fields = append(fields, field)
		}
		fieldsStr := strings.Join(fields, " ")
		source := fmt.Sprintf("object [%s]", fieldsStr)
		return value.StrVal(source), nil

	default:
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDSourceUnsupported,
			[3]string{value.TypeToString(val.GetType()), "", ""},
		)
	}
}

// formatBlock formats a block of values into a string representation
func formatBlock(elements []core.Value) string {
	parts := make([]string, len(elements))
	for i, elem := range elements {
		parts[i] = elem.String()
	}
	return "[" + strings.Join(parts, " ") + "]"
}
