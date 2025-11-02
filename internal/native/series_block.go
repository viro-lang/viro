// Package native - block-specific series operations
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// BlockFirst returns the element at the current position of a block.
// Feature: 004-dynamic-function-invocation
func BlockFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"first", "1", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(blk.Elements) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	if blk.GetIndex() >= len(blk.Elements) {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is at tail", "", ""})
	}

	return blk.Elements[blk.GetIndex()], nil
}

// BlockLast returns the last element of a block.
// Feature: 004-dynamic-function-invocation
func BlockLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"last", "1", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(blk.Elements) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return blk.Elements[len(blk.Elements)-1], nil
}

// BlockAppend appends a value to the end of a block.
// Modifies the block in-place and returns it.
// Feature: 004-dynamic-function-invocation
func BlockAppend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"append", "2", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	// Append the value to the block
	blk.Elements = append(blk.Elements, args[1])

	// Return the modified block
	return args[0], nil
}

// BlockInsert inserts a value at the beginning of a block.
// Modifies the block in-place and returns it.
// Feature: 004-dynamic-function-invocation
func BlockInsert(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"insert", "2", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	// Insert the value at the beginning
	blk.Elements = append([]core.Value{args[1]}, blk.Elements...)

	// Return the modified block
	return args[0], nil
}

// BlockLength returns the number of elements in a block.
// Feature: 004-dynamic-function-invocation
func BlockLength(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"length?", "1", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	return value.NewIntVal(int64(len(blk.Elements))), nil
}

// BlockCopy implements copy action for block values.
// Feature: 004-dynamic-function-invocation
func BlockCopy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"copy", "1", "0"})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	// --part refinement: copy only first N elements
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
		if count < 0 || count > len(blk.Elements) {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"copy --part", "block", "out of range"})
		}
		elems := make([]core.Value, count)
		copy(elems, blk.Elements[:count])
		return value.NewBlockVal(elems), nil
	}

	// Full copy
	return value.NewBlockVal(append([]core.Value{}, blk.Elements...)), nil
}

// BlockFind implements find action for block values.
// Feature: 004-dynamic-function-invocation
func BlockFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"find", "2", string(rune(len(args) + '0'))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	sought := args[1]

	// --last refinement: find last occurrence
	lastVal, hasLast := refValues["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.NewLogicVal(true))

	if isLast {
		for i := len(blk.Elements) - 1; i >= 0; i-- {
			if blk.Elements[i].Equals(sought) {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	} else {
		for i, v := range blk.Elements {
			if v.Equals(sought) {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	}

	return value.NewNoneVal(), nil
}

// BlockRemove implements remove action for block values.
// Feature: 004-dynamic-function-invocation
func BlockRemove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"remove", "1", "0"})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	// --part refinement: remove N elements
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

	if count < 0 || count > len(blk.Elements) {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"remove", "block", "out of range"})
	}

	blk.SetIndex(0)
	blk.Remove(count)
	return args[0], nil
}

// BlockSkip implements skip action for block values.
// Feature: 004-dynamic-function-invocation
func BlockSkip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"skip", "2", string(rune(len(args) + '0'))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	newIndex := blk.GetIndex() + count
	if newIndex < 0 || newIndex > len(blk.Elements) {
		newIndex = len(blk.Elements)
	}
	blk.SetIndex(newIndex)

	return args[0], nil
}

// BlockNext implements next action for block values.
// Returns a new block reference with index advanced by 1.
// Feature: 004-dynamic-function-invocation
func BlockNext(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"next", "1", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	// Create a new reference with advanced index
	newBlock := blk.Clone()
	newIndex := blk.GetIndex() + 1
	if newIndex > len(blk.Elements) {
		newIndex = len(blk.Elements)
	}
	newBlock.SetIndex(newIndex)

	return newBlock, nil
}

func BlockBack(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if err := ensureArgCount(args, 1, "back"); err != nil {
		return value.NewNoneVal(), err
	}

	return seriesBack(args[0], "back")
}

// BlockHead implements head action for block values.
// Returns a new block reference positioned at index 0 (head).
// Feature: 004-dynamic-function-invocation
func BlockHead(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"head", "1", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	// Create a new reference with index at head (0)
	newBlock := blk.Clone()
	newBlock.SetIndex(0)

	return newBlock, nil
}

// BlockTake implements take action for block values.
// Feature: 004-dynamic-function-invocation
func BlockTake(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"take", "2", string(rune(len(args) + '0'))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	start := blk.GetIndex()
	end := min(start+count, len(blk.Elements))
	newElements := blk.Elements[start:end]
	blk.SetIndex(end)

	return value.NewBlockVal(newElements), nil
}

// BlockSort implements sort action for block values.
// Feature: 004-dynamic-function-invocation
func BlockSort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"sort", "1", "0"})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(blk.Elements) == 0 {
		return args[0], nil
	}

	// Check if all elements are of the same comparable type
	firstType := blk.Elements[0].GetType()
	for _, v := range blk.Elements {
		if v.GetType() != firstType || (v.GetType() != value.TypeInteger && v.GetType() != value.TypeString) {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDNotComparable, [3]string{"sort", "mixed types", ""})
		}
	}

	value.SortBlock(blk)
	return args[0], nil
}

// BlockReverse implements reverse action for block values.
// Feature: 004-dynamic-function-invocation
func BlockReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"reverse", "1", "0"})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	// Reverse the elements in place
	for i, j := 0, len(blk.Elements)-1; i < j; i, j = i+1, j-1 {
		blk.Elements[i], blk.Elements[j] = blk.Elements[j], blk.Elements[i]
	}

	return args[0], nil
}

func BlockIndex(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"index?", "1", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	return value.NewIntVal(int64(blk.GetIndex() + 1)), nil
}

func BlockAt(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"at", "2", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	if zeroBasedIndex < 0 || zeroBasedIndex >= len(blk.Elements) {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"at", "block", "index out of range"})
	}

	return blk.Elements[zeroBasedIndex], nil
}

// BlockTail returns a new block containing all elements except the first one.
// Feature: 004-dynamic-function-invocation
func BlockTail(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"tail", "1", fmt.Sprintf("%d", len(args))})
	}

	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(blk.Elements) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return value.NewBlockVal(append([]core.Value{}, blk.Elements[1:]...)), nil
}
