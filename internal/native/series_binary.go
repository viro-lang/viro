package native

import (
	"bytes"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

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
