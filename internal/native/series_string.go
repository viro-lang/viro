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

// StringCopy implements copy action for string values.
// Feature: 004-dynamic-function-invocation
func StringCopy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"copy", "1", "0"})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	// --part refinement: copy only first N characters
	partVal, hasPart := refValues["part"]
	hasPart = hasPart && partVal.GetType() != value.TypeNone

	if hasPart {
		if partVal.GetType() != value.TypeInteger {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
		}
		count64, ok := value.AsInteger(partVal)
		if !ok {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
		}
		count := int(count64)
		if count < 0 || count > len(str.String()) {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"copy --part", "string", "out of range"})
		}
		// Use substring
		runes := []rune(str.String())
		return value.StrVal(string(runes[:count])), nil
	}

	// Full copy
	return value.StrVal(str.String()), nil
}

// StringFind implements find action for string values.
// Feature: 004-dynamic-function-invocation
func StringFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"find", "2", string(rune(len(args) + '0'))})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	sought, ok := value.AsString(args[1])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	// --last refinement: find last occurrence
	lastVal, hasLast := refValues["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.LogicVal(true))

	runes := []rune(str.String())
	soughtRunes := []rune(sought.String())

	if len(soughtRunes) == 0 {
		return value.NoneVal(), nil // Empty string not found
	}

	if isLast {
		for i := len(runes) - len(soughtRunes); i >= 0; i-- {
			match := true
			for j := range soughtRunes {
				if runes[i+j] != soughtRunes[j] {
					match = false
					break
				}
			}
			if match {
				return value.IntVal(int64(i + 1)), nil
			}
		}
	} else {
		for i := 0; i <= len(runes)-len(soughtRunes); i++ {
			match := true
			for j := range soughtRunes {
				if runes[i+j] != soughtRunes[j] {
					match = false
					break
				}
			}
			if match {
				return value.IntVal(int64(i + 1)), nil
			}
		}
	}

	return value.NoneVal(), nil
}

// StringRemove implements remove action for string values.
// Feature: 004-dynamic-function-invocation
func StringRemove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"remove", "1", "0"})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	// --part refinement: remove N characters
	partVal, hasPart := refValues["part"]
	hasPart = hasPart && partVal.GetType() != value.TypeNone

	count := 1
	if hasPart {
		if partVal.GetType() != value.TypeInteger {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
		}
		count64, _ := value.AsInteger(partVal)
		count = int(count64)
	}

	if count < 0 || count > len(str.String()) {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"remove", "string", "out of range"})
	}

	str.SetIndex(0)
	str.Remove(count)
	return args[0], nil
}

// StringSkip implements skip action for string values.
// Feature: 004-dynamic-function-invocation
func StringSkip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"skip", "2", string(rune(len(args) + '0'))})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsInteger(countVal)
	count := int(count64)

	newIndex := str.Index() + count
	if newIndex < 0 || newIndex > len(str.String()) {
		newIndex = len(str.String())
	}
	str.SetIndex(newIndex)

	return args[0], nil
}

// StringTake implements take action for string values.
// Feature: 004-dynamic-function-invocation
func StringTake(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"take", "2", string(rune(len(args) + '0'))})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsInteger(countVal)
	count := int(count64)

	start := str.Index()
	end := min(start+count, len(str.String()))
	newRunes := str.Runes()[start:end]
	str.SetIndex(end)

	return value.StrVal(string(newRunes)), nil
}

// StringReverse implements reverse action for string values.
// Feature: 004-dynamic-function-invocation
func StringReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"reverse", "1", "0"})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	r := str.Runes()
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	str.SetRunes(r)

	return args[0], nil
}

// StringSort implements sort action for string values.
// Feature: 004-dynamic-function-invocation
func StringSort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"sort", "1", "0"})
	}

	str, ok := value.AsString(args[0])
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	value.SortString(str)
	return args[0], nil
}
