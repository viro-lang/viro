package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func BlockFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	sought := args[1]

	// --last refinement: find last occurrence
	lastVal, hasLast := refValues["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.NewLogicVal(true))

	elements := block.Elements

	if isLast {
		for i := len(elements) - 1; i >= 0; i-- {
			if elements[i].Equals(sought) {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	} else {
		for i, elem := range elements {
			if elem.Equals(sought) {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	}

	return value.NewNoneVal(), nil
}

func BlockReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	elements := block.Elements
	for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
		elements[i], elements[j] = elements[j], elements[i]
	}

	return args[0], nil
}

func BlockSort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(block.Elements) == 0 {
		return args[0], nil
	}

	firstType := block.Elements[0].GetType()
	for _, v := range block.Elements {
		if v.GetType() != firstType || (v.GetType() != value.TypeInteger && v.GetType() != value.TypeString) {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDNotComparable, [3]string{"sort", "mixed types", ""})
		}
	}

	value.SortBlock(block)
	return args[0], nil
}

func BlockAt(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	return seriesAt(block, zeroBasedIndex)
}

func BlockPoke(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if err := validateIndex(zeroBasedIndex, len(block.Elements)); err != nil {
		return value.NewNoneVal(), err
	}
	block.Elements[zeroBasedIndex] = args[2]
	return args[2], nil
}

func BlockTrim(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	hasHead := hasRefinement(refValues, "head")
	hasTail := hasRefinement(refValues, "tail")
	hasAuto := hasRefinement(refValues, "auto")   // ignored for blocks
	hasLines := hasRefinement(refValues, "lines") // ignored for blocks
	hasAll := hasRefinement(refValues, "all")
	hasWith, withVal := getRefinementValue(refValues, "with")

	flags := []bool{hasHead, hasTail, hasAuto, hasLines, hasAll, hasWith}
	flagCount := 0
	for _, flag := range flags {
		if flag {
			flagCount++
		}
	}

	if flagCount > 1 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trim refinements are mutually exclusive", "", ""},
		)
	}

	if hasWith {
		return blockTrimWith(block, withVal), nil
	}

	if flagCount == 0 {
		return blockTrimDefault(block), nil
	}

	if hasHead {
		return blockTrimHead(block), nil
	}
	if hasTail {
		return blockTrimTail(block), nil
	}
	if hasAll {
		return blockTrimAll(block), nil
	}

	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDAssertionFailed, [3]string{"unexpected trim refinement state", "", ""})
}

func isNoneLike(v core.Value) bool {
	if v.GetType() == value.TypeNone {
		return true
	}
	if v.GetType() == value.TypeWord {
		if word, ok := value.AsWordValue(v); ok && word == "none" {
			return true
		}
	}
	return false
}

func blockTrimDefault(block *value.BlockValue) core.Value {
	start := 0
	end := len(block.Elements) - 1

	for start <= end && isNoneLike(block.Elements[start]) {
		start++
	}

	for end >= start && isNoneLike(block.Elements[end]) {
		end--
	}

	block.Elements = block.Elements[start : end+1]
	return block
}

func blockTrimHead(block *value.BlockValue) core.Value {
	start := 0
	for start < len(block.Elements) && isNoneLike(block.Elements[start]) {
		start++
	}

	block.Elements = block.Elements[start:]
	return block
}

func blockTrimTail(block *value.BlockValue) core.Value {
	end := len(block.Elements) - 1
	for end >= 0 && isNoneLike(block.Elements[end]) {
		end--
	}

	block.Elements = block.Elements[:end+1]
	return block
}

func blockTrimAll(block *value.BlockValue) core.Value {
	elements := make([]core.Value, 0, len(block.Elements))

	for _, elem := range block.Elements {
		if !isNoneLike(elem) {
			elements = append(elements, elem)
		}
	}

	block.Elements = elements
	return block
}

func blockTrimWith(block *value.BlockValue, withVal core.Value) core.Value {
	elements := make([]core.Value, 0, len(block.Elements))

	for _, elem := range block.Elements {
		if !elem.Equals(withVal) {
			elements = append(elements, elem)
		}
	}

	block.Elements = elements
	return block
}

func BlockSelect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	hasDefault := false
	defaultVal, ok := refValues["default"]
	if ok && defaultVal.GetType() != value.TypeNone {
		hasDefault = true
	}

	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	sought := args[1]
	elements := block.Elements

	for i, elem := range elements {
		matches := false
		if isWordLike(elem.GetType()) && isWordLike(sought.GetType()) {
			elemSymbol, _ := value.AsWordValue(elem)
			searchSymbol, _ := value.AsWordValue(sought)
			matches = elemSymbol == searchSymbol
		} else {
			matches = elem.Equals(sought)
		}

		if matches {
			if i+1 < len(elements) {
				return elements[i+1], nil
			}
			return value.NewNoneVal(), nil
		}
	}

	if hasDefault {
		return defaultVal, nil
	}
	return value.NewNoneVal(), nil
}
