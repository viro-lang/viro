package native

import (
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func StringFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	sought := args[1]
	soughtStr, ok := value.AsStringValue(sought)
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(sought.GetType()), ""})
	}

	// --last refinement: find last occurrence
	lastVal, hasLast := refValues["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.NewLogicVal(true))

	haystack := str.String()
	needle := soughtStr.String()

	if isLast {
		pos := strings.LastIndex(haystack, needle)
		if pos == -1 {
			return value.NewNoneVal(), nil
		}
		return value.NewIntVal(int64(pos + 1)), nil
	} else {
		pos := strings.Index(haystack, needle)
		if pos == -1 {
			return value.NewNoneVal(), nil
		}
		return value.NewIntVal(int64(pos + 1)), nil
	}
}

func StringReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	runes := str.Runes()
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	str.SetRunes(runes)
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

func StringPoke(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	if err := validateIndex(zeroBasedIndex, len(str.Runes())); err != nil {
		return value.NewNoneVal(), err
	}
	r, err := validateStringValue(args[2])
	if err != nil {
		return value.NewNoneVal(), err
	}
	str.Runes()[zeroBasedIndex] = r
	return args[2], nil
}

func StringTrim(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	input := string(str.Runes())

	hasHead := hasRefinement(refValues, "head")
	hasTail := hasRefinement(refValues, "tail")
	hasAuto := hasRefinement(refValues, "auto")
	hasLines := hasRefinement(refValues, "lines")
	hasAll := hasRefinement(refValues, "all")
	hasWith, withVal := getRefinementValue(refValues, "with")

	flagCount := countTrue(hasHead, hasTail, hasAuto, hasLines, hasAll, hasWith)

	if flagCount > 1 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trim refinements are mutually exclusive", "", ""},
		)
	}

	if hasWith {
		withStr, ok := value.AsStringValue(withVal)
		if !ok {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"string", value.TypeToString(withVal.GetType()), "--with"},
			)
		}
		charsToRemove := string(withStr.Runes())
		str.SetRunes([]rune(trimWith(input, charsToRemove)))
		return args[0], nil
	}

	if flagCount == 0 {
		str.SetRunes([]rune(trimDefault(input)))
		return args[0], nil
	}

	if hasHead {
		str.SetRunes([]rune(trimHead(input)))
		return args[0], nil
	}
	if hasTail {
		str.SetRunes([]rune(trimTail(input)))
		return args[0], nil
	}
	if hasAuto {
		str.SetRunes([]rune(trimAuto(input)))
		return args[0], nil
	}
	if hasLines {
		str.SetRunes([]rune(trimLines(input)))
		return args[0], nil
	}
	if hasAll {
		str.SetRunes([]rune(trimAll(input)))
		return args[0], nil
	}

	panic("unreachable: all trim refinement combinations should be handled above")
}

func StringSelect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	hasDefault := false
	defaultVal, ok := refValues["default"]
	if ok && defaultVal.GetType() != value.TypeNone {
		hasDefault = true
	}

	str, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	if targetStr, ok2 := value.AsStringValue(args[1]); ok2 {
		haystack := str.String()
		needle := targetStr.String()
		pos := strings.Index(haystack, needle)
		if pos == -1 {
			if hasDefault {
				return defaultVal, nil
			}
			return value.NewNoneVal(), nil
		}
		remainder := haystack[pos+len(needle):]
		return value.NewStringValue(remainder), nil
	}
	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
}
