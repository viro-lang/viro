// Package native - block-specific series operations
package native

import (
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// BlockFirst returns the first element of a block.
// Feature: 004-dynamic-function-invocation
func BlockFirst(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"first", "1", "0"})
	}

	blk, ok := args[0].AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", args[0].Type.String(), ""})
	}

	if len(blk.Elements) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return blk.Elements[0], nil
}

// BlockLast returns the last element of a block.
// Feature: 004-dynamic-function-invocation
func BlockLast(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"last", "1", "0"})
	}

	blk, ok := args[0].AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", args[0].Type.String(), ""})
	}

	if len(blk.Elements) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
	}

	return blk.Elements[len(blk.Elements)-1], nil
}

// BlockAppend appends a value to the end of a block.
// Modifies the block in-place and returns it.
// Feature: 004-dynamic-function-invocation
func BlockAppend(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"append", "2", string(rune(len(args) + '0'))})
	}

	blk, ok := args[0].AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", args[0].Type.String(), ""})
	}

	// Append the value to the block
	blk.Elements = append(blk.Elements, args[1])

	// Return the modified block
	return args[0], nil
}

// BlockInsert inserts a value at the beginning of a block.
// Modifies the block in-place and returns it.
// Feature: 004-dynamic-function-invocation
func BlockInsert(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"insert", "2", string(rune(len(args) + '0'))})
	}

	blk, ok := args[0].AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", args[0].Type.String(), ""})
	}

	// Insert the value at the beginning
	blk.Elements = append([]value.Value{args[1]}, blk.Elements...)

	// Return the modified block
	return args[0], nil
}

// BlockLength returns the number of elements in a block.
// Feature: 004-dynamic-function-invocation
func BlockLength(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDArgCount, [3]string{"length?", "1", "0"})
	}

	blk, ok := args[0].AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", args[0].Type.String(), ""})
	}

	return value.IntVal(int64(len(blk.Elements))), nil
}
