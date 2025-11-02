package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func BinaryFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if bin.GetIndex() >= len(bin.Bytes()) {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is at tail", "", ""})
	}

	return value.NewIntVal(int64(bin.Bytes()[bin.GetIndex()])), nil
}

func BinaryLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if len(bin.Bytes()) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
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
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"byte value must be 0-255", "", ""})
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
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"byte value must be 0-255", "", ""})
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

	// --part refinement: copy only first N bytes
	partVal, hasPart := refValues["part"]
	hasPart = hasPart && partVal.GetType() != value.TypeNone

	if hasPart {
		if partVal.GetType() != value.TypeInteger {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
		}
		count64, ok := value.AsIntValue(partVal)
		if !ok {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
		}
		count := int(count64)
		if count < 0 || count > bin.Length() {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"copy --part", "binary", "out of range"})
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

	// --part refinement: remove N bytes
	partVal, hasPart := refValues["part"]
	hasPart = hasPart && partVal.GetType() != value.TypeNone

	count := 1
	if hasPart {
		if partVal.GetType() != value.TypeInteger {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
		}
		count64, _ := value.AsIntValue(partVal)
		count = int(count64)
	}

	if count < 0 || count > bin.Length() {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"remove", "binary", "out of range"})
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

func BinaryNext(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesNext(args[0])
}

func BinaryBack(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesBack(args[0])
}

func BinaryHead(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesHead(args[0])
}

func BinaryIndex(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesIndex(args[0])
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

func BinaryTail(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesTail(args[0])
}
