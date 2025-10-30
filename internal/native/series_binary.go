// Package native - binary-specific series operations
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// BinaryFirst returns the first byte of a binary value.
// Feature: 004-dynamic-function-invocation
func BinaryFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"first", "1", fmt.Sprintf("%d", len(args))})
	}

	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if len(bin.Bytes()) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return value.NewIntVal(int64(bin.First())), nil
}

// BinaryLast returns the last byte of a binary value.
// Feature: 004-dynamic-function-invocation
func BinaryLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"last", "1", fmt.Sprintf("%d", len(args))})
	}

	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	if len(bin.Bytes()) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return value.NewIntVal(int64(bin.Last())), nil
}

// BinaryAppend appends a byte or binary value to the end of a binary value.
// Modifies the binary in-place and returns it.
// Feature: 004-dynamic-function-invocation
func BinaryAppend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"append", "2", fmt.Sprintf("%d", len(args))})
	}

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

// BinaryInsert inserts a byte or binary value at the beginning of a binary value.
// Modifies the binary in-place and returns it.
// Feature: 004-dynamic-function-invocation
func BinaryInsert(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"insert", "2", fmt.Sprintf("%d", len(args))})
	}

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

// BinaryLength returns the number of bytes in a binary value.
// Feature: 004-dynamic-function-invocation
func BinaryLength(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"length?", "1", fmt.Sprintf("%d", len(args))})
	}

	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	return value.NewIntVal(int64(bin.Length())), nil
}

// BinaryCopy implements copy action for binary values.
// Feature: 004-dynamic-function-invocation
func BinaryCopy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"copy", "1", "0"})
	}

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

// BinaryFind implements find action for binary values.
// Feature: 004-dynamic-function-invocation
func BinaryFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"find", "2", string(rune(len(args) + '0'))})
	}

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

// BinaryRemove implements remove action for binary values.
// Feature: 004-dynamic-function-invocation
func BinaryRemove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"remove", "1", "0"})
	}

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

// BinarySkip implements skip action for binary values.
// Feature: 004-dynamic-function-invocation
func BinarySkip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"skip", "2", string(rune(len(args) + '0'))})
	}

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

// BinaryTake implements take action for binary values.
// Feature: 004-dynamic-function-invocation
func BinaryTake(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"take", "2", string(rune(len(args) + '0'))})
	}

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

	start := bin.GetIndex()
	end := min(start+count, bin.Length())
	newBytes := bin.Bytes()[start:end]
	bin.SetIndex(end)

	return value.NewBinaryVal(newBytes), nil
}

// BinaryReverse implements reverse action for binary values.
// Feature: 004-dynamic-function-invocation
func BinaryReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"reverse", "1", "0"})
	}

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

// BinarySort implements sort action for binary values.
// Feature: 004-dynamic-function-invocation
func BinarySort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"sort", "1", "0"})
	}

	bin, ok := value.AsBinaryValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[0].GetType()), ""})
	}

	value.SortBinary(bin)
	return args[0], nil
}
