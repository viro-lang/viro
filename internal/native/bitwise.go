package native

import (
	"fmt"
	"math/bits"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func BitAnd(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.and", 2, len(args))
	}

	leftType := args[0].GetType()
	rightType := args[1].GetType()

	if leftType != rightType {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"bit.and", "operands must be same type", ""},
		)
	}

	switch leftType {
	case value.TypeInteger:
		left, _ := value.AsIntValue(args[0])
		right, _ := value.AsIntValue(args[1])
		return value.NewIntVal(left & right), nil

	case value.TypeBinary:
		left, _ := value.AsBinaryValue(args[0])
		right, _ := value.AsBinaryValue(args[1])
		return binaryAnd(left, right), nil

	default:
		return value.NewNoneVal(), typeError("bit.and", "integer! binary!", args[0])
	}
}

func BitOr(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.or", 2, len(args))
	}

	leftType := args[0].GetType()
	rightType := args[1].GetType()

	if leftType != rightType {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"bit.or", "operands must be same type", ""},
		)
	}

	switch leftType {
	case value.TypeInteger:
		left, _ := value.AsIntValue(args[0])
		right, _ := value.AsIntValue(args[1])
		return value.NewIntVal(left | right), nil

	case value.TypeBinary:
		left, _ := value.AsBinaryValue(args[0])
		right, _ := value.AsBinaryValue(args[1])
		return binaryOr(left, right), nil

	default:
		return value.NewNoneVal(), typeError("bit.or", "integer! binary!", args[0])
	}
}

func BitXor(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.xor", 2, len(args))
	}

	leftType := args[0].GetType()
	rightType := args[1].GetType()

	if leftType != rightType {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"bit.xor", "operands must be same type", ""},
		)
	}

	switch leftType {
	case value.TypeInteger:
		left, _ := value.AsIntValue(args[0])
		right, _ := value.AsIntValue(args[1])
		return value.NewIntVal(left ^ right), nil

	case value.TypeBinary:
		left, _ := value.AsBinaryValue(args[0])
		right, _ := value.AsBinaryValue(args[1])
		return binaryXor(left, right), nil

	default:
		return value.NewNoneVal(), typeError("bit.xor", "integer! binary!", args[0])
	}
}

func BitNot(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("bit.not", 1, len(args))
	}

	switch args[0].GetType() {
	case value.TypeInteger:
		val, _ := value.AsIntValue(args[0])
		return value.NewIntVal(^val), nil

	case value.TypeBinary:
		bin, _ := value.AsBinaryValue(args[0])
		return binaryNot(bin), nil

	default:
		return value.NewNoneVal(), typeError("bit.not", "integer! binary!", args[0])
	}
}

func BitShl(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.shl", 2, len(args))
	}

	countType := args[1].GetType()
	if countType != value.TypeInteger {
		return value.NewNoneVal(), typeError("bit.shl", "integer!", args[1])
	}
	count, _ := value.AsIntValue(args[1])

	if count < 0 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDOutOfBounds,
			[3]string{"bit.shl", "shift count must be non-negative", ""},
		)
	}

	switch args[0].GetType() {
	case value.TypeInteger:
		val, _ := value.AsIntValue(args[0])
		return value.NewIntVal(val << uint(count)), nil

	case value.TypeBinary:
		bin, _ := value.AsBinaryValue(args[0])
		return binaryShl(bin, count), nil

	default:
		return value.NewNoneVal(), typeError("bit.shl", "integer! binary!", args[0])
	}
}

func BitShr(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.shr", 2, len(args))
	}

	countType := args[1].GetType()
	if countType != value.TypeInteger {
		return value.NewNoneVal(), typeError("bit.shr", "integer!", args[1])
	}
	count, _ := value.AsIntValue(args[1])

	if count < 0 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDOutOfBounds,
			[3]string{"bit.shr", "shift count must be non-negative", ""},
		)
	}

	switch args[0].GetType() {
	case value.TypeInteger:
		val, _ := value.AsIntValue(args[0])
		return value.NewIntVal(val >> uint(count)), nil

	case value.TypeBinary:
		bin, _ := value.AsBinaryValue(args[0])
		return binaryShr(bin, count), nil

	default:
		return value.NewNoneVal(), typeError("bit.shr", "integer! binary!", args[0])
	}
}

func BitOn(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.on", 2, len(args))
	}

	if args[0].GetType() != value.TypeInteger {
		return value.NewNoneVal(), typeError("bit.on", "integer!", args[0])
	}

	if args[1].GetType() != value.TypeInteger {
		return value.NewNoneVal(), typeError("bit.on", "integer!", args[1])
	}

	val, _ := value.AsIntValue(args[0])
	pos, _ := value.AsIntValue(args[1])

	if pos < 0 || pos >= 64 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("bit.on: bit position %d out of range (valid: 0-63)", pos), "", ""},
		)
	}

	return value.NewIntVal(val | (1 << uint(pos))), nil
}

func BitOff(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.off", 2, len(args))
	}

	if args[0].GetType() != value.TypeInteger {
		return value.NewNoneVal(), typeError("bit.off", "integer!", args[0])
	}

	if args[1].GetType() != value.TypeInteger {
		return value.NewNoneVal(), typeError("bit.off", "integer!", args[1])
	}

	val, _ := value.AsIntValue(args[0])
	pos, _ := value.AsIntValue(args[1])

	if pos < 0 || pos >= 64 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("bit.off: bit position %d out of range (valid: 0-63)", pos), "", ""},
		)
	}

	return value.NewIntVal(val &^ (1 << uint(pos))), nil
}

func BitCount(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("bit.count", 1, len(args))
	}

	switch args[0].GetType() {
	case value.TypeInteger:
		val, _ := value.AsIntValue(args[0])
		return value.NewIntVal(int64(bits.OnesCount64(uint64(val)))), nil

	case value.TypeBinary:
		bin, _ := value.AsBinaryValue(args[0])
		return value.NewIntVal(countBinaryBits(bin)), nil

	default:
		return value.NewNoneVal(), typeError("bit.count", "integer! binary!", args[0])
	}
}

func binaryLogicOp(left, right *value.BinaryValue, op func(byte, byte) byte, padZero bool) core.Value {
	leftBytes := left.Bytes()
	rightBytes := right.Bytes()

	maxLen := len(leftBytes)
	if len(rightBytes) > maxLen {
		maxLen = len(rightBytes)
	}

	result := make([]byte, maxLen)
	leftLen := len(leftBytes)
	rightLen := len(rightBytes)

	for i := 0; i < maxLen; i++ {
		leftIdx := leftLen - maxLen + i
		rightIdx := rightLen - maxLen + i

		leftByte := byte(0)
		if leftIdx >= 0 {
			leftByte = leftBytes[leftIdx]
		}

		rightByte := byte(0)
		if rightIdx >= 0 {
			rightByte = rightBytes[rightIdx]
		}

		if padZero {
			result[i] = op(leftByte, rightByte)
		} else {
			if leftIdx >= 0 && rightIdx >= 0 {
				result[i] = op(leftByte, rightByte)
			} else if leftIdx >= 0 {
				result[i] = leftByte
			} else {
				result[i] = rightByte
			}
		}
	}

	return value.NewBinaryValue(result)
}

func binaryAnd(left, right *value.BinaryValue) core.Value {
	return binaryLogicOp(left, right, func(a, b byte) byte { return a & b }, true)
}

func binaryOr(left, right *value.BinaryValue) core.Value {
	return binaryLogicOp(left, right, func(a, b byte) byte { return a | b }, false)
}

func binaryXor(left, right *value.BinaryValue) core.Value {
	return binaryLogicOp(left, right, func(a, b byte) byte { return a ^ b }, false)
}

func binaryNot(b *value.BinaryValue) core.Value {
	data := b.Bytes()
	result := make([]byte, len(data))

	for i, byteVal := range data {
		result[i] = ^byteVal
	}

	return value.NewBinaryValue(result)
}

func binaryShl(b *value.BinaryValue, count int64) core.Value {
	data := b.Bytes()
	if count <= 0 || len(data) == 0 {
		return value.NewBinaryValue(data)
	}

	bitCount := int(count)
	byteShift := bitCount / 8
	bitShift := bitCount % 8

	result := make([]byte, len(data))

	if byteShift >= len(data) {
		return value.NewBinaryValue(result)
	}

	for i := 0; i < len(data)-byteShift; i++ {
		result[i+byteShift] = data[i]
	}

	if bitShift > 0 {
		carry := byte(0)
		for i := len(result) - 1; i >= 0; i-- {
			newCarry := result[i] >> (8 - bitShift)
			result[i] = (result[i] << bitShift) | carry
			carry = newCarry
		}
	}

	return value.NewBinaryValue(result)
}

func binaryShr(b *value.BinaryValue, count int64) core.Value {
	data := b.Bytes()
	if count <= 0 || len(data) == 0 {
		return value.NewBinaryValue(data)
	}

	bitCount := int(count)
	byteShift := bitCount / 8
	bitShift := bitCount % 8

	result := make([]byte, len(data))

	if byteShift >= len(data) {
		return value.NewBinaryValue(result)
	}

	for i := byteShift; i < len(data); i++ {
		result[i-byteShift] = data[i]
	}

	if bitShift > 0 {
		carry := byte(0)
		for i := 0; i < len(result); i++ {
			newCarry := result[i] << (8 - bitShift)
			result[i] = (result[i] >> bitShift) | carry
			carry = newCarry
		}
	}

	return value.NewBinaryValue(result)
}

func countBinaryBits(b *value.BinaryValue) int64 {
	data := b.Bytes()
	count := 0
	for _, byteVal := range data {
		count += bits.OnesCount8(byteVal)
	}
	return int64(count)
}
