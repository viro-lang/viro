package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func StringFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	strVal := str.String()
	if len(strVal) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"first element", "", ""})
	}
	if str.GetIndex() >= len(strVal) {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", str.GetIndex()), fmt.Sprintf("%d", len(strVal)), ""})
	}

	return value.NewStrVal(string(strVal[str.GetIndex()])), nil
}

func StringLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	strVal := str.String()
	if len(strVal) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"last element", "", ""})
	}

	return value.NewStrVal(string(strVal[len(strVal)-1])), nil
}

func StringAppend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	appendStr, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	str.Append(appendStr)

	return args[0], nil
}

func StringInsert(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	insertStr, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	str.Insert(insertStr)

	return args[0], nil
}

func StringLength(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	return value.NewIntVal(int64(len(str.String()))), nil
}

func StringCopy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	count, hasPart, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if hasPart {
		if err := validatePartCount(str, count); err != nil {
			return value.NewNoneVal(), err
		}
		// Use substring
		runes := []rune(str.String())
		return value.NewStrVal(string(runes[:count])), nil
	}

	// Full copy
	return value.NewStrVal(str.String()), nil
}

func StringFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	sought, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	// --last refinement: find last occurrence
	lastVal, hasLast := refValues["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.NewLogicVal(true))

	runes := []rune(str.String())
	soughtRunes := []rune(sought.String())

	if len(soughtRunes) == 0 {
		return value.NewNoneVal(), nil // Empty string not found
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
				return value.NewIntVal(int64(i + 1)), nil
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
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	}

	return value.NewNoneVal(), nil
}

func StringRemove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	count, _, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if err := validatePartCount(str, count); err != nil {
		return value.NewNoneVal(), err
	}

	str.SetIndex(0)
	str.Remove(count)
	return args[0], nil
}

func StringSkip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	newIndex := str.GetIndex() + count
	if newIndex < 0 || newIndex > len(str.String()) {
		newIndex = len(str.String())
	}
	str.SetIndex(newIndex)

	return args[0], nil
}

func StringNext(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesNext(args[0])
}

func StringBack(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesBack(args[0])
}

func StringHead(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesHead(args[0])
}

func StringIndex(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesIndex(args[0])
}

func StringReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	r := str.Runes()
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	str.SetRunes(r)

	return args[0], nil
}

func StringSort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	value.SortString(str)
	return args[0], nil
}

func StringAt(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	return seriesAt(str, zeroBasedIndex)
}

func StringTake(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	runes := str.Runes()
	start := str.GetIndex()
	end := min(start+count, len(runes))
	takenRunes := runes[start:end]
	str.SetIndex(end)

	return value.NewStrVal(string(takenRunes)), nil
}

func StringTail(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesTail(args[0])
}

func StringEmpty(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesEmpty(args[0])
}

func StringHeadQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesHeadQ(args[0])
}

func StringTailQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesTailQ(args[0])
}
