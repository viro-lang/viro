package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func BlockIntersect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block1, ok1 := value.AsBlockValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	block2, ok2 := value.AsBlockValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[1].GetType()), ""})
	}

	resultMap := make(map[string]core.Value)
	set2 := make(map[string]bool)

	for _, elem := range block2.Elements {
		key := elem.Mold()
		set2[key] = true
	}

	for _, elem := range block1.Elements {
		key := elem.Mold()
		if set2[key] {
			if _, exists := resultMap[key]; !exists {
				resultMap[key] = elem
			}
		}
	}

	result := make([]core.Value, 0, len(resultMap))
	for _, elem := range block1.Elements {
		key := elem.Mold()
		if val, exists := resultMap[key]; exists {
			result = append(result, val)
			delete(resultMap, key)
		}
	}

	return value.NewBlockValue(result), nil
}

func BlockDifference(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block1, ok1 := value.AsBlockValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	block2, ok2 := value.AsBlockValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[1].GetType()), ""})
	}

	set2 := make(map[string]bool)
	for _, elem := range block2.Elements {
		key := elem.Mold()
		set2[key] = true
	}

	resultMap := make(map[string]core.Value)
	for _, elem := range block1.Elements {
		key := elem.Mold()
		if !set2[key] {
			if _, exists := resultMap[key]; !exists {
				resultMap[key] = elem
			}
		}
	}

	result := make([]core.Value, 0, len(resultMap))
	for _, elem := range block1.Elements {
		key := elem.Mold()
		if val, exists := resultMap[key]; exists {
			result = append(result, val)
			delete(resultMap, key)
		}
	}

	return value.NewBlockValue(result), nil
}

func BlockUnion(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	block1, ok1 := value.AsBlockValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	block2, ok2 := value.AsBlockValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[1].GetType()), ""})
	}

	seen := make(map[string]core.Value)
	result := make([]core.Value, 0)

	for _, elem := range block1.Elements {
		key := elem.Mold()
		if _, exists := seen[key]; !exists {
			seen[key] = elem
			result = append(result, elem)
		}
	}

	for _, elem := range block2.Elements {
		key := elem.Mold()
		if _, exists := seen[key]; !exists {
			seen[key] = elem
			result = append(result, elem)
		}
	}

	return value.NewBlockValue(result), nil
}

func StringIntersect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str1, ok1 := value.AsStringValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	str2, ok2 := value.AsStringValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	set2 := make(map[rune]bool)
	for _, r := range str2.Runes() {
		set2[r] = true
	}

	seen := make(map[rune]bool)
	result := make([]rune, 0)

	for _, r := range str1.Runes() {
		if set2[r] && !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}

	return value.NewStrVal(string(result)), nil
}

func StringDifference(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str1, ok1 := value.AsStringValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	str2, ok2 := value.AsStringValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	set2 := make(map[rune]bool)
	for _, r := range str2.Runes() {
		set2[r] = true
	}

	seen := make(map[rune]bool)
	result := make([]rune, 0)

	for _, r := range str1.Runes() {
		if !set2[r] && !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}

	return value.NewStrVal(string(result)), nil
}

func StringUnion(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	str1, ok1 := value.AsStringValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[0].GetType()), ""})
	}

	str2, ok2 := value.AsStringValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	seen := make(map[rune]bool)
	result := make([]rune, 0)

	for _, r := range str1.Runes() {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}

	for _, r := range str2.Runes() {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}

	return value.NewStrVal(string(result)), nil
}

func BinaryIntersect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin1, ok1 := value.AsBinaryValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	bin2, ok2 := value.AsBinaryValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[1].GetType()), ""})
	}

	set2 := make(map[byte]bool)
	for _, b := range bin2.Bytes() {
		set2[b] = true
	}

	seen := make(map[byte]bool)
	result := make([]byte, 0)

	for _, b := range bin1.Bytes() {
		if set2[b] && !seen[b] {
			seen[b] = true
			result = append(result, b)
		}
	}

	return value.NewBinaryValue(result), nil
}

func BinaryDifference(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin1, ok1 := value.AsBinaryValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	bin2, ok2 := value.AsBinaryValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[1].GetType()), ""})
	}

	set2 := make(map[byte]bool)
	for _, b := range bin2.Bytes() {
		set2[b] = true
	}

	seen := make(map[byte]bool)
	result := make([]byte, 0)

	for _, b := range bin1.Bytes() {
		if !set2[b] && !seen[b] {
			seen[b] = true
			result = append(result, b)
		}
	}

	return value.NewBinaryValue(result), nil
}

func BinaryUnion(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin1, ok1 := value.AsBinaryValue(args[0])
	if !ok1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	bin2, ok2 := value.AsBinaryValue(args[1])
	if !ok2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[1].GetType()), ""})
	}

	seen := make(map[byte]bool)
	result := make([]byte, 0)

	for _, b := range bin1.Bytes() {
		if !seen[b] {
			seen[b] = true
			result = append(result, b)
		}
	}

	for _, b := range bin2.Bytes() {
		if !seen[b] {
			seen[b] = true
			result = append(result, b)
		}
	}

	return value.NewBinaryValue(result), nil
}
