// Package native - string-specific series operations
package native

import (
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// StringFirst returns the first character of a string.
// Feature: 004-dynamic-function-invocation
func StringFirst(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"first", "1", "0"})
	}

	str, ok := args[0].AsString()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", args[0].Type.String(), ""})
	}

	strVal := str.String()
	if len(strVal) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	// Return first character as a string
	return value.StrVal(string(strVal[0])), nil
}

// StringLast returns the last character of a string.
// Feature: 004-dynamic-function-invocation
func StringLast(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"last", "1", "0"})
	}

	str, ok := args[0].AsString()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", args[0].Type.String(), ""})
	}

	strVal := str.String()
	if len(strVal) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	// Return last character as a string
	return value.StrVal(string(strVal[len(strVal)-1])), nil
}

// StringAppend appends a string to the end of another string.
// Modifies the string in-place and returns it.
// Feature: 004-dynamic-function-invocation
func StringAppend(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"append", "2", string(rune(len(args) + '0'))})
	}

	str, ok := args[0].AsString()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", args[0].Type.String(), ""})
	}

	// Second argument must be a string
	appendStr, ok := args[1].AsString()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", args[1].Type.String(), ""})
	}

	// Append the string
	str.Append(appendStr)

	// Return the modified string
	return args[0], nil
}

// StringInsert inserts a string at the beginning of another string.
// Modifies the string in-place and returns it.
// Feature: 004-dynamic-function-invocation
func StringInsert(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"insert", "2", string(rune(len(args) + '0'))})
	}

	str, ok := args[0].AsString()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", args[0].Type.String(), ""})
	}

	// Second argument must be a string
	insertStr, ok := args[1].AsString()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", args[1].Type.String(), ""})
	}

	// Insert at the beginning
	str.Insert(insertStr)

	// Return the modified string
	return args[0], nil
}

// StringLength returns the number of characters in a string.
// Feature: 004-dynamic-function-invocation
func StringLength(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"length?", "1", "0"})
	}

	str, ok := args[0].AsString()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", args[0].Type.String(), ""})
	}

	return value.IntVal(int64(len(str.String()))), nil
}
