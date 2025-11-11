package native

import (
	"math/bits"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// BitAnd performs bitwise AND operation.
// Supports both integer! and binary! types.
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

// BitOr performs bitwise OR operation.
// Supports both integer! and binary! types.
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

// BitXor performs bitwise XOR operation.
// Supports both integer! and binary! types.
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

// BitNot performs bitwise NOT operation.
// Supports both integer! and binary! types.
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

// BitShl performs left shift operation.
// Supports both integer! and binary! types.
func BitShl(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.shl", 2, len(args))
	}

	// Second arg must be integer (shift count)
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

	// First arg can be integer or binary
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

// BitShr performs right shift operation.
// Supports both integer! and binary! types.
func BitShr(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("bit.shr", 2, len(args))
	}

	// Second arg must be integer (shift count)
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

	// First arg can be integer or binary
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

// BitOn sets a specific bit to 1 in an integer.
// Integer-only operation.
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

	return value.NewIntVal(val | (1 << uint(pos))), nil
}

// BitOff clears a specific bit to 0 in an integer.
// Integer-only operation.
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

	return value.NewIntVal(val &^ (1 << uint(pos))), nil
}

// BitCount counts the number of set bits (1-bits).
// Supports both integer! and binary! types.
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

// binaryAnd performs right-aligned AND on binary values.
// Zero-pads shorter operand.
func binaryAnd(left, right *value.BinaryValue) core.Value {
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

		result[i] = leftByte & rightByte
	}

	return value.NewBinaryValue(result)
}

// binaryOr performs right-aligned OR on binary values.
// Copies remainder from longer operand.
func binaryOr(left, right *value.BinaryValue) core.Value {
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

		result[i] = leftByte | rightByte
	}

	return value.NewBinaryValue(result)
}

// binaryXor performs right-aligned XOR on binary values.
// Copies remainder from longer operand.
func binaryXor(left, right *value.BinaryValue) core.Value {
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

		result[i] = leftByte ^ rightByte
	}

	return value.NewBinaryValue(result)
}

// binaryNot flips all bits in all bytes.
func binaryNot(b *value.BinaryValue) core.Value {
	data := b.Bytes()
	result := make([]byte, len(data))

	for i, byteVal := range data {
		result[i] = ^byteVal
	}

	return value.NewBinaryValue(result)
}

// binaryShl shifts binary value left by specified bits.
// Overflow is lost, no new bytes created.
func binaryShl(b *value.BinaryValue, count int64) core.Value {
	data := b.Bytes()
	if count <= 0 || len(data) == 0 {
		return value.NewBinaryValue(data)
	}

	bitCount := int(count)
	byteShift := bitCount / 8
	bitShift := bitCount % 8

	result := make([]byte, len(data))

	// Handle byte-level shift
	if byteShift >= len(data) {
		// Complete overflow
		return value.NewBinaryValue(result) // All zeros
	}

	// Shift bytes left (to higher indices)
	for i := 0; i < len(data)-byteShift; i++ {
		result[i+byteShift] = data[i]
	}

	// Shift bits within bytes
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

// binaryShr shifts binary value right by specified bits.
// Underflow is lost, no new bytes created.
func binaryShr(b *value.BinaryValue, count int64) core.Value {
	data := b.Bytes()
	if count <= 0 || len(data) == 0 {
		return value.NewBinaryValue(data)
	}

	bitCount := int(count)
	byteShift := bitCount / 8
	bitShift := bitCount % 8

	result := make([]byte, len(data))

	// Handle byte-level shift
	if byteShift >= len(data) {
		// Complete underflow
		return value.NewBinaryValue(result) // All zeros
	}

	// Shift bytes
	for i := byteShift; i < len(data); i++ {
		result[i-byteShift] = data[i]
	}

	// Shift bits within bytes
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

// countBinaryBits counts set bits across all bytes in binary value.
func countBinaryBits(b *value.BinaryValue) int64 {
	data := b.Bytes()
	count := 0
	for _, byteVal := range data {
		count += bits.OnesCount8(byteVal)
	}
	return int64(count)
}
