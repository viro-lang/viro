// Package native - string-specific series operations
package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// StringFirst returns the first character of a string.
// Feature: 004-dynamic-function-invocation
func StringFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"first", "1", "0"})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	strVal := str.String()
	if len(strVal) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return value.StrVal(string(strVal[0])), nil
}

// StringLast returns the last character of a string.
// Feature: 004-dynamic-function-invocation
func StringLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"last", "1", "0"})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	strVal := str.String()
	if len(strVal) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return value.StrVal(string(strVal[len(strVal)-1])), nil
}

// StringAppend appends a string to the end of another string.
// Modifies the string in-place and returns it.
// Feature: 004-dynamic-function-invocation
func StringAppend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"append", "2", string(rune(len(args) + '0'))})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	appendStr, ok := value.AsString(args[1])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	str.Append(appendStr)

	return args[0], nil
}

// StringInsert inserts a string at the beginning of another string.
// Modifies the string in-place and returns it.
// Feature: 004-dynamic-function-invocation
func StringInsert(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"insert", "2", string(rune(len(args) + '0'))})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	insertStr, ok := value.AsString(args[1])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	str.Insert(insertStr)

	return args[0], nil
}

// StringLength returns the number of characters in a string.
// Feature: 004-dynamic-function-invocation
func StringLength(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"length?", "1", "0"})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	return value.IntVal(int64(len(str.String()))), nil
}
