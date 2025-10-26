// Package native implements built-in native functions for Viro.
//
// Control flow natives implement conditional execution and iteration.
// Contract per contracts/control-flow.md: when, if, loop, while
package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/trace"
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
func When(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("when", 2, len(args))
	}

	// First argument is condition (already evaluated)
	condition := args[0]

	// Second argument must be a block (NOT evaluated yet)
	if args[1].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("when", "block", args[1])
	}

	// Convert condition to truthy/falsy
	// Per contract: none and false are falsy, all others are truthy
	isTruthy := ToTruthy(condition)

	if isTruthy {
		// Evaluate the block
		block, _ := value.AsBlockValue(args[1])
		return eval.DoBlock(block.Elements)
	}

	// Condition is falsy, return none
	return value.NewNoneVal(), nil
}

// If implements the 'if' conditional native.
//
// Contract: if condition [true-block] [false-block]
// - Evaluates condition to truthy/falsy
// - If truthy: evaluates true-block and returns result
// - If falsy: evaluates false-block and returns result
// - Both blocks required (error if missing)
func If(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("if", 3, len(args))
	}

	// First argument is condition (already evaluated)
	condition := args[0]

	// Second argument must be a block (true branch)
	if args[1].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("if", "block for true branch", args[1])
	}

	// Third argument must be a block (false branch)
	if args[2].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("if", "block for false branch", args[2])
	}

	// Convert condition to truthy/falsy
	isTruthy := ToTruthy(condition)

	if isTruthy {
		// Evaluate true-block
		block, _ := value.AsBlockValue(args[1])
		return eval.DoBlock(block.Elements)
	}

	// Evaluate false-block
	block, _ := value.AsBlockValue(args[2])
	return eval.DoBlock(block.Elements)
}

// Loop implements the 'loop' iteration native.
//
// Contract: loop count [block]
// - Count must be a non-negative integer
// - Executes block count times
// - Returns result of last iteration, or none if count is 0
func Loop(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("loop", 2, len(args))
	}

	// First argument must be an integer
	count, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("loop", "integer for count", args[0])
	}

	// Count must be non-negative
	if count < 0 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"loop count must be non-negative", "", ""},
		)
	}

	// Second argument must be a block
	if args[1].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("loop", "block for body", args[1])
	}

	block, _ := value.AsBlockValue(args[1])

	// If count is 0, return none without executing
	if count == 0 {
		return value.NewNoneVal(), nil
	}

	// Execute block count times
	var result core.Value
	var err error
	for range count {
		result, err = eval.DoBlock(block.Elements)
		if err != nil {
			return value.NewNoneVal(), err
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
func While(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("while", 2, len(args))
	}

	// First argument must be a block (condition)
	if args[0].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("while", "block for condition", args[0])
	}

	// Second argument must be a block (body)
	if args[1].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("while", "block for body", args[1])
	}

	conditionBlock, _ := value.AsBlockValue(args[0])
	bodyBlock, _ := value.AsBlockValue(args[1])

	result := value.NewNoneVal()

	// Loop while condition is truthy
	for {
		// Evaluate condition block
		conditionResult, err := eval.DoBlock(conditionBlock.Elements)
		if err != nil {
			return value.NewNoneVal(), err
		}

		// Check if condition is truthy
		if !ToTruthy(conditionResult) {
			break
		}

		// Evaluate body block
		result, err = eval.DoBlock(bodyBlock.Elements)
		if err != nil {
			return value.NewNoneVal(), err
		}
	}

	return result, nil
}

// Reduce implements the 'reduce' native.
//
// Contract: reduce value
// - If value is a block, evaluates each element and returns a new block with the results
// - If value is not a block, returns the value as-is
// - Similar to REBOL's reduce function
//
// This enables blocks to be evaluated for their contents, useful for:
// - Creating blocks with computed values
// - String interpolation patterns
// - Building data structures dynamically
func Reduce(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("reduce", 1, len(args))
	}

	if args[0].GetType() != value.TypeBlock {
		return args[0], nil
	}

	block, _ := value.AsBlockValue(args[0])
	vals := block.Elements
	reducedElements := make([]core.Value, 0)

	position := 0

	for position < len(vals) {
		newPos, result, err := eval.EvaluateExpression(vals, position)
		if err != nil {
			return value.NewNoneVal(), err
		}

		reducedElements = append(reducedElements, result)
		position = newPos
	}

	return value.NewBlockVal(reducedElements), nil
}

// ToTruthy converts a value to truthy/falsy per Viro semantics.
//
// Contract per contracts/control-flow.md:
// - none → false
// - false (logic value) → false
// - All other values → true (including 0, "", [])
func ToTruthy(val core.Value) bool {
	switch val.GetType() {
	case value.TypeNone:
		return false
	case value.TypeLogic:
		b, _ := value.AsLogicValue(val)
		return b
	default:
		// All other values are truthy (including 0, "", [])
		return true
	}
}

// Trace implements the 'trace' native for tracing control (Feature 002, FR-020).
//
// Contract: trace --on [--only block] [--exclude block] [--file path] [--append]
//
//	trace --off
//
// T144: Implements trace --on with refinements
// T145: Implements trace --off
func Trace(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// Check for --on or --off refinement
	hasOn := false
	hasOff := false

	if val, ok := refValues["on"]; ok && ToTruthy(val) {
		hasOn = true
	}
	if val, ok := refValues["off"]; ok && ToTruthy(val) {
		hasOff = true
	}

	if !hasOn && !hasOff {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trace requires --on or --off refinement", "", ""},
		)
	}

	if hasOn && hasOff {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trace cannot have both --on and --off", "", ""},
		)
	}

	if hasOff {
		// Disable tracing
		if trace.GlobalTraceSession != nil {
			trace.GlobalTraceSession.Disable()
		}
		return value.NewNoneVal(), nil
	}

	// Handle --on case
	if trace.GlobalTraceSession == nil {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trace session not initialized", "", ""},
		)
	}

	filters := trace.TraceFilters{}

	// Handle --only refinement
	if onlyVal, ok := refValues["only"]; ok && onlyVal.GetType() != value.TypeNone {
		if onlyVal.GetType() != value.TypeBlock {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--only requires block of words", "", ""},
			)
		}
		onlyBlk, _ := value.AsBlockValue(onlyVal)
		for _, elem := range onlyBlk.Elements {
			if elem.GetType() != value.TypeWord {
				return value.NewNoneVal(), verror.NewScriptError(
					verror.ErrIDTypeMismatch,
					[3]string{"--only block must contain only words", "", ""},
				)
			}
			word, _ := value.AsWordValue(elem)
			filters.IncludeWords = append(filters.IncludeWords, word)
		}
	}

	// Handle --exclude refinement
	if excludeVal, ok := refValues["exclude"]; ok && excludeVal.GetType() != value.TypeNone {
		if excludeVal.GetType() != value.TypeBlock {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--exclude requires block of words", "", ""},
			)
		}
		excludeBlk, _ := value.AsBlockValue(excludeVal)
		for _, elem := range excludeBlk.Elements {
			if elem.GetType() != value.TypeWord {
				return value.NewNoneVal(), verror.NewScriptError(
					verror.ErrIDTypeMismatch,
					[3]string{"--exclude block must contain only words", "", ""},
				)
			}
			word, _ := value.AsWordValue(elem)
			filters.ExcludeWords = append(filters.ExcludeWords, word)
		}
	}

	// Handle --file refinement with sandbox validation
	if fileVal, ok := refValues["file"]; ok && fileVal.GetType() != value.TypeNone {
		if fileVal.GetType() != value.TypeString {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--file requires string path", "", ""},
			)
		}
		fileStr, _ := value.AsStringValue(fileVal)
		filePath := fileStr.String()

		// Validate path is within sandbox
		_, err := resolveSandboxPath(filePath)
		if err != nil {
			return value.NewNoneVal(), verror.NewAccessError(
				verror.ErrIDSandboxViolation,
				[3]string{"trace file path escapes sandbox", filePath, ""},
			)
		}

		// Note: Actual file redirection would require reinitializing the trace session
		// For now, we only validate the path is within sandbox
		// Full implementation deferred until trace session supports dynamic reconfiguration
	} // Handle --append refinement (validates file must also be provided)
	if appendVal, ok := refValues["append"]; ok && ToTruthy(appendVal) {
		if _, hasFile := refValues["file"]; !hasFile || refValues["file"].GetType() == value.TypeNone {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"--append requires --file to be specified", "", ""},
			)
		}
	}

	trace.GlobalTraceSession.Enable(filters)
	return value.NewNoneVal(), nil
}

// TraceQuery implements the 'trace?' query native (Feature 002, FR-020).
//
// Contract: trace?
// Returns boolean indicating if tracing is enabled
//
// T146: Implements trace? query
func TraceQuery(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("trace?", 0, len(args))
	}

	// Return simple boolean indicating trace state
	if trace.GlobalTraceSession == nil {
		return value.NewLogicVal(false), nil
	}

	enabled := trace.GlobalTraceSession.IsEnabled()
	return value.NewLogicVal(enabled), nil
}

// Debug implements the 'debug' native for debugger control (Feature 002, FR-021).
//
// Contract: debug --on | --off | --breakpoint word | --remove id |
//
//	--step | --next | --finish | --continue | --locals | --stack
//
// T148-T152: Implements debug commands
func Debug(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if debug.GlobalDebugger == nil {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"debugger not initialized", "", ""},
		)
	}

	// Check which refinement is present
	if val, ok := refValues["on"]; ok && ToTruthy(val) {
		// Enable debugger
		debug.GlobalDebugger.Enable()
		return value.NewNoneVal(), nil
	}

	if val, ok := refValues["off"]; ok && ToTruthy(val) {
		// Disable debugger
		debug.GlobalDebugger.Disable()
		return value.NewNoneVal(), nil
	}

	// For all other operations, debugger must be enabled
	// Check before processing any other refinement
	if debug.GlobalDebugger.Mode() == debug.DebugModeOff {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"debugger not enabled - use debug --on first", "", ""},
		)
	}

	if val, ok := refValues["breakpoint"]; ok && val.GetType() != value.TypeNone {
		// Set breakpoint on word (accepts lit-word, get-word, or word)
		var word string
		switch val.GetType() {
		case value.TypeLitWord, value.TypeGetWord, value.TypeWord:
			word, _ = value.AsWordValue(val)
		default:
			return value.NewNoneVal(), typeError("debug --breakpoint", "word", val)
		}

		// Validate word exists in current context (lookup covers all scopes)
		var found bool
		_, found = eval.Lookup(word)
		if !found {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDNoValue,
				[3]string{"cannot set breakpoint on unknown word", word, ""},
			)
		}

		id := debug.GlobalDebugger.SetBreakpoint(word)
		return value.NewIntVal(int64(id)), nil
	}

	if val, ok := refValues["remove"]; ok && val.GetType() != value.TypeNone {
		// Remove breakpoint by ID
		id, ok := value.AsIntValue(val)
		if !ok {
			return value.NewNoneVal(), typeError("debug --remove", "integer ID", val)
		}

		// Find and remove breakpoint by ID
		found := debug.GlobalDebugger.RemoveBreakpointByID(id)

		if !found {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDNoSuchBreakpoint,
				[3]string{"breakpoint not found", "", ""},
			)
		}
		return value.NewNoneVal(), nil
	}

	if val, ok := refValues["step"]; ok && ToTruthy(val) {
		debug.GlobalDebugger.EnableStepping()
		return value.NewNoneVal(), nil
	}

	if val, ok := refValues["next"]; ok && ToTruthy(val) {
		debug.GlobalDebugger.EnableStepping()
		return value.NewNoneVal(), nil
	}

	if val, ok := refValues["finish"]; ok && ToTruthy(val) {
		debug.GlobalDebugger.DisableStepping()
		return value.NewNoneVal(), nil
	}

	if val, ok := refValues["continue"]; ok && ToTruthy(val) {
		debug.GlobalDebugger.DisableStepping()
		return value.NewNoneVal(), nil
	}

	if val, ok := refValues["locals"]; ok && ToTruthy(val) {
		// Create empty object for now (TODO: populate with actual locals)
		fields := []string{}
		initializers := make(map[string][]core.Value)
		return instantiateObject(eval, -1, nil, fields, initializers)
	}

	if val, ok := refValues["stack"]; ok && ToTruthy(val) {
		// Return block with call stack entries
		// For now, return empty block
		return value.NewBlockVal([]core.Value{}), nil
	}

	return value.NewNoneVal(), verror.NewScriptError(
		verror.ErrIDInvalidOperation,
		[3]string{"debug requires a valid refinement", "", ""},
	)
}
