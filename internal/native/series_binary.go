package native

import (
	"bytes"
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func BinaryFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if len(bin.Bytes()) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"first element", "", ""})
	}
	if bin.GetIndex() >= len(bin.Bytes()) {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", bin.GetIndex()), fmt.Sprintf("%d", len(bin.Bytes())), ""})
	}

	return value.NewIntVal(int64(bin.Bytes()[bin.GetIndex()])), nil
}

func BinaryLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if len(bin.Bytes()) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"last element", "", ""})
	}

	return value.NewIntVal(int64(bin.Last())), nil
}

func BinaryAppend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	// Accept integers (0-255) or binary values
	switch args[1].GetType() {
	case value.TypeInteger:
		intVal, _ := value.AsIntValue(args[1])
		if intVal < 0 || intVal > 255 {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", intVal), "255", ""})
		}
		bin.Append(byte(intVal))
	case value.TypeBinary:
		appendBin, _ := value.AsBinaryValue(args[1])
		bin.Append(appendBin)
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer or binary", value.TypeToString(args[1].GetType()), ""})
	}

	return args[0], nil
}

func BinaryInsert(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	// Accept integers (0-255) or binary values
	switch args[1].GetType() {
	case value.TypeInteger:
		intVal, _ := value.AsIntValue(args[1])
		if intVal < 0 || intVal > 255 {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", intVal), "255", ""})
		}
		bin.SetIndex(0)
		bin.Insert(byte(intVal))
	case value.TypeBinary:
		insertBin, _ := value.AsBinaryValue(args[1])
		bin.SetIndex(0)
		bin.Insert(insertBin)
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer or binary", value.TypeToString(args[1].GetType()), ""})
	}

	return args[0], nil
}

func BinaryLength(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	return value.NewIntVal(int64(bin.Length())), nil
}

func BinaryCopy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	count, hasPart, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if hasPart {
		if err := validatePartCount(bin, count); err != nil {
			return value.NewNoneVal(), err
		}
		// Copy first count bytes
		bytes := make([]byte, count)
		copy(bytes, bin.Bytes()[:count])
		return value.NewBinaryVal(bytes), nil
	}

	// Full copy
	return value.NewBinaryVal(append([]byte{}, bin.Bytes()...)), nil
}

func BinaryFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	sought := args[1]
	soughtByte, ok := value.AsIntValue(sought)
	if !ok || soughtByte < 0 || soughtByte > 255 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"byte value (0-255)", value.TypeToString(sought.GetType()), ""})
	}

	// --last refinement: find last occurrence
	lastVal, hasLast := refValues["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.NewLogicVal(true))

	bytes := bin.Bytes()
	targetByte := byte(soughtByte)

	if isLast {
		for i := len(bytes) - 1; i >= 0; i-- {
			if bytes[i] == targetByte {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	} else {
		for i, b := range bytes {
			if b == targetByte {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	}

	return value.NewNoneVal(), nil
}

func BinaryRemove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	count, _, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if err := validatePartCount(bin, count); err != nil {
		return value.NewNoneVal(), err
	}

	bin.SetIndex(0)
	bin.Remove(count)
	return args[0], nil
}

func BinarySkip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	newIndex := bin.GetIndex() + count
	if newIndex < 0 || newIndex > bin.Length() {
		newIndex = bin.Length()
	}
	bin.SetIndex(newIndex)

	return args[0], nil
}

func BinaryReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	bytes := bin.Bytes()
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}

	return args[0], nil
}

func BinarySort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	value.SortBinary(bin)
	return args[0], nil
}

func BinaryAt(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	return seriesAt(bin, zeroBasedIndex)
}

func BinaryTake(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	bytes := bin.Bytes()
	start := bin.GetIndex()
	end := min(start+count, len(bytes))
	takenBytes := bytes[start:end]
	bin.SetIndex(end)

	return value.NewBinaryVal(takenBytes), nil
}

func BinaryPoke(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if err := validateIndex(zeroBasedIndex, len(bin.Bytes())); err != nil {
		return value.NewNoneVal(), err
	}
	b, err := validateByteValue(args[2])
	if err != nil {
		return value.NewNoneVal(), err
	}
	bin.Bytes()[zeroBasedIndex] = b
	return args[2], nil
}

func BinarySelect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	hasDefault := false
	defaultVal, ok := refValues["default"]
	if ok && defaultVal.GetType() != value.TypeNone {
		hasDefault = true
	}

	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if targetBytes, ok2 := value.AsBinaryValue(args[1]); ok2 {
		haystack := bin.Bytes()
		needle := targetBytes.Bytes()
		pos := bytes.Index(haystack, needle)
		if pos == -1 {
			if hasDefault {
				return defaultVal, nil
			}
			return value.NewNoneVal(), nil
		}
		remainder := haystack[pos+len(needle):]
		return value.NewBinaryValue(remainder), nil
	}
	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[1].GetType()), ""})
}

func BinaryClear(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	bin.SetIndex(0)
	bin.Remove(bin.Length())
	return args[0], nil
}

func BinaryChange(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	currentIndex := bin.GetIndex()
	if err := validateIndex(currentIndex, len(bin.Bytes())); err != nil {
		return value.NewNoneVal(), err
	}

	b, err := validateByteValue(args[1])
	if err != nil {
		return value.NewNoneVal(), err
	}
	bytes := bin.Bytes()
	bytes[currentIndex] = b
	newIndex := currentIndex + 1
	if newIndex > len(bytes) {
		newIndex = len(bytes)
	}
	bin.SetIndex(newIndex)
	return args[1], nil
}
