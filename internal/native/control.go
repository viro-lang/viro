// Package native implements built-in native functions for Viro.
//
// Control flow natives implement conditional execution and iteration.
// Contract per contracts/control-flow.md: when, if, loop, while
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// When implements the 'when' conditional native.
//
// Contract: when condition [block]
// - Evaluates condition to truthy/falsy
// - If truthy: evaluates block and returns result
// - If falsy: returns none without evaluating block
//
// This is a special native that needs access to evaluator to evaluate blocks.
func When(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"when", "2", fmt.Sprintf("%d", len(args))},
		)
	}

	// First argument is condition (already evaluated)
	condition := args[0]

	// Second argument must be a block (NOT evaluated yet)
	if args[1].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"when", "block", args[1].Type.String()},
		)
	}

	// Convert condition to truthy/falsy
	// Per contract: none and false are falsy, all others are truthy
	isTruthy := ToTruthy(condition)

	if isTruthy {
		// Evaluate the block
		block, _ := args[1].AsBlock()
		return eval.Do_Blk(block.Elements)
	}

	// Condition is falsy, return none
	return value.NoneVal(), nil
}

// If implements the 'if' conditional native.
//
// Contract: if condition [true-block] [false-block]
// - Evaluates condition to truthy/falsy
// - If truthy: evaluates true-block and returns result
// - If falsy: evaluates false-block and returns result
// - Both blocks required (error if missing)
func If(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 3 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"if", "3", fmt.Sprintf("%d", len(args))},
		)
	}

	// First argument is condition (already evaluated)
	condition := args[0]

	// Second argument must be a block (true branch)
	if args[1].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"if", "block for true branch", args[1].Type.String()},
		)
	}

	// Third argument must be a block (false branch)
	if args[2].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"if", "block for false branch", args[2].Type.String()},
		)
	}

	// Convert condition to truthy/falsy
	isTruthy := ToTruthy(condition)

	if isTruthy {
		// Evaluate true-block
		block, _ := args[1].AsBlock()
		return eval.Do_Blk(block.Elements)
	}

	// Evaluate false-block
	block, _ := args[2].AsBlock()
	return eval.Do_Blk(block.Elements)
}

// Loop implements the 'loop' iteration native.
//
// Contract: loop count [block]
// - Count must be a non-negative integer
// - Executes block count times
// - Returns result of last iteration, or none if count is 0
func Loop(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"loop", "2", fmt.Sprintf("%d", len(args))},
		)
	}

	// First argument must be an integer
	count, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"loop", "integer for count", args[0].Type.String()},
		)
	}

	// Count must be non-negative
	if count < 0 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"loop count must be non-negative", "", ""},
		)
	}

	// Second argument must be a block
	if args[1].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"loop", "block for body", args[1].Type.String()},
		)
	}

	block, _ := args[1].AsBlock()

	// If count is 0, return none without executing
	if count == 0 {
		return value.NoneVal(), nil
	}

	// Execute block count times
	var result value.Value
	var err *verror.Error
	for i := int64(0); i < count; i++ {
		result, err = eval.Do_Blk(block.Elements)
		if err != nil {
			return value.NoneVal(), err
		}
	}

	return result, nil
}

// While implements the 'while' conditional loop native.
//
// Contract: while [condition] [body]
// - Condition must be a block (re-evaluated each iteration)
// - Body must be a block
// - Loops while condition evaluates to truthy
// - Returns result of last iteration, or none if never executed
func While(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"while", "2", fmt.Sprintf("%d", len(args))},
		)
	}

	// First argument must be a block (condition)
	if args[0].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"while", "block for condition", args[0].Type.String()},
		)
	}

	// Second argument must be a block (body)
	if args[1].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"while", "block for body", args[1].Type.String()},
		)
	}

	conditionBlock, _ := args[0].AsBlock()
	bodyBlock, _ := args[1].AsBlock()

	result := value.NoneVal()

	// Loop while condition is truthy
	for {
		// Evaluate condition block
		conditionResult, err := eval.Do_Blk(conditionBlock.Elements)
		if err != nil {
			return value.NoneVal(), err
		}

		// Check if condition is truthy
		if !ToTruthy(conditionResult) {
			break
		}

		// Evaluate body block
		result, err = eval.Do_Blk(bodyBlock.Elements)
		if err != nil {
			return value.NoneVal(), err
		}
	}

	return result, nil
}

// ToTruthy converts a value to truthy/falsy per Viro semantics.
//
// Contract per contracts/control-flow.md:
// - none → false
// - false (logic value) → false
// - All other values → true (including 0, "", [])
func ToTruthy(val value.Value) bool {
	switch val.Type {
	case value.TypeNone:
		return false
	case value.TypeLogic:
		b, _ := val.AsLogic()
		return b
	default:
		// All other values are truthy (including 0, "", [])
		return true
	}
}
