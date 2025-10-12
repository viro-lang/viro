// Package native implements built-in native functions for Viro.
//
// Control flow natives implement conditional execution and iteration.
// Contract per contracts/control-flow.md: when, if, loop, while
package native

import (
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
func When(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("when", 2, len(args))
	}

	// First argument is condition (already evaluated)
	condition := args[0]

	// Second argument must be a block (NOT evaluated yet)
	if args[1].Type != value.TypeBlock {
		return value.NoneVal(), typeError("when", "block", args[1])
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
func If(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 3 {
		return value.NoneVal(), arityError("if", 3, len(args))
	}

	// First argument is condition (already evaluated)
	condition := args[0]

	// Second argument must be a block (true branch)
	if args[1].Type != value.TypeBlock {
		return value.NoneVal(), typeError("if", "block for true branch", args[1])
	}

	// Third argument must be a block (false branch)
	if args[2].Type != value.TypeBlock {
		return value.NoneVal(), typeError("if", "block for false branch", args[2])
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
func Loop(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("loop", 2, len(args))
	}

	// First argument must be an integer
	count, ok := args[0].AsInteger()
	if !ok {
		return value.NoneVal(), typeError("loop", "integer for count", args[0])
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
		return value.NoneVal(), typeError("loop", "block for body", args[1])
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
func While(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("while", 2, len(args))
	}

	// First argument must be a block (condition)
	if args[0].Type != value.TypeBlock {
		return value.NoneVal(), typeError("while", "block for condition", args[0])
	}

	// Second argument must be a block (body)
	if args[1].Type != value.TypeBlock {
		return value.NoneVal(), typeError("while", "block for body", args[1])
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

// Reduce implements the 'reduce' native.
//
// Contract: reduce block
// - Evaluates each element in the block
// - Returns a new block containing the evaluated results
// - Similar to REBOL's reduce function
//
// This enables blocks to be evaluated for their contents, useful for:
// - Creating blocks with computed values
// - String interpolation patterns
// - Building data structures dynamically
func Reduce(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("reduce", 1, len(args))
	}

	// Argument must be a block
	if args[0].Type != value.TypeBlock {
		return value.NoneVal(), typeError("reduce", "block", args[0])
	}

	block, _ := args[0].AsBlock()
	reducedElements := make([]value.Value, len(block.Elements))

	// Evaluate each element in the block
	for i, elem := range block.Elements {
		result, err := eval.Do_Next(elem)
		if err != nil {
			return value.NoneVal(), err
		}
		reducedElements[i] = result
	}

	return value.BlockVal(reducedElements), nil
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

// Trace implements the 'trace' native for tracing control (Feature 002, FR-020).
//
// Contract: trace --on [--only block] [--exclude block] [--file path] [--append]
//
//	trace --off
//
// T144: Implements trace --on with refinements
// T145: Implements trace --off
func Trace(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
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
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trace requires --on or --off refinement", "", ""},
		)
	}

	if hasOn && hasOff {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trace cannot have both --on and --off", "", ""},
		)
	}

	if hasOff {
		// Disable tracing
		if GlobalTraceSession != nil {
			GlobalTraceSession.Disable()
		}
		return value.NoneVal(), nil
	}

	// Handle --on case
	if GlobalTraceSession == nil {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"trace session not initialized", "", ""},
		)
	}

	filters := TraceFilters{}

	// Handle --only refinement
	if onlyVal, ok := refValues["only"]; ok && onlyVal.Type != value.TypeNone {
		if onlyVal.Type != value.TypeBlock {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--only requires block of words", "", ""},
			)
		}
		onlyBlk, _ := onlyVal.AsBlock()
		for _, elem := range onlyBlk.Elements {
			if elem.Type != value.TypeWord {
				return value.NoneVal(), verror.NewScriptError(
					verror.ErrIDTypeMismatch,
					[3]string{"--only block must contain only words", "", ""},
				)
			}
			word, _ := elem.AsWord()
			filters.IncludeWords = append(filters.IncludeWords, word)
		}
	}

	// Handle --exclude refinement
	if excludeVal, ok := refValues["exclude"]; ok && excludeVal.Type != value.TypeNone {
		if excludeVal.Type != value.TypeBlock {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--exclude requires block of words", "", ""},
			)
		}
		excludeBlk, _ := excludeVal.AsBlock()
		for _, elem := range excludeBlk.Elements {
			if elem.Type != value.TypeWord {
				return value.NoneVal(), verror.NewScriptError(
					verror.ErrIDTypeMismatch,
					[3]string{"--exclude block must contain only words", "", ""},
				)
			}
			word, _ := elem.AsWord()
			filters.ExcludeWords = append(filters.ExcludeWords, word)
		}
	}

	// Handle --file refinement with sandbox validation
	if fileVal, ok := refValues["file"]; ok && fileVal.Type != value.TypeNone {
		if fileVal.Type != value.TypeString {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--file requires string path", "", ""},
			)
		}
		fileStr, _ := fileVal.AsString()
		filePath := fileStr.String()

		// Validate path is within sandbox
		_, err := resolveSandboxPath(filePath)
		if err != nil {
			return value.NoneVal(), verror.NewAccessError(
				verror.ErrIDSandboxViolation,
				[3]string{"trace file path escapes sandbox", filePath, ""},
			)
		}

		// Note: Actual file redirection would require reinitializing the trace session
		// For now, we only validate the path is within sandbox
		// Full implementation deferred until trace session supports dynamic reconfiguration
	} // Handle --append refinement (validates file must also be provided)
	if appendVal, ok := refValues["append"]; ok && ToTruthy(appendVal) {
		if _, hasFile := refValues["file"]; !hasFile || refValues["file"].Type == value.TypeNone {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"--append requires --file to be specified", "", ""},
			)
		}
	}

	GlobalTraceSession.Enable(filters)
	return value.NoneVal(), nil
}

// TraceQuery implements the 'trace?' query native (Feature 002, FR-020).
//
// Contract: trace?
// Returns boolean indicating if tracing is enabled
//
// T146: Implements trace? query
func TraceQuery(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 0 {
		return value.NoneVal(), arityError("trace?", 0, len(args))
	}

	// Return simple boolean indicating trace state
	if GlobalTraceSession == nil {
		return value.LogicVal(false), nil
	}

	enabled := GlobalTraceSession.IsEnabled()
	return value.LogicVal(enabled), nil
}

// Debug implements the 'debug' native for debugger control (Feature 002, FR-021).
//
// Contract: debug --on | --off | --breakpoint word | --remove id |
//
//	--step | --next | --finish | --continue | --locals | --stack
//
// T148-T152: Implements debug commands
func Debug(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if GlobalDebugger == nil {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"debugger not initialized", "", ""},
		)
	}

	// Check which refinement is present
	if val, ok := refValues["on"]; ok && ToTruthy(val) {
		// Enable debugger
		GlobalDebugger.mu.Lock()
		GlobalDebugger.mode = DebugModeActive
		GlobalDebugger.mu.Unlock()
		return value.NoneVal(), nil
	}

	if val, ok := refValues["off"]; ok && ToTruthy(val) {
		// Disable debugger
		GlobalDebugger.mu.Lock()
		GlobalDebugger.mode = DebugModeOff
		GlobalDebugger.breakpoints = make(map[string]int)
		GlobalDebugger.stepping = false
		GlobalDebugger.mu.Unlock()
		return value.NoneVal(), nil
	}

	// For all other operations, debugger must be enabled
	// Check before processing any other refinement
	if GlobalDebugger.Mode() == DebugModeOff {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidOperation,
			[3]string{"debugger not enabled - use debug --on first", "", ""},
		)
	}

	if val, ok := refValues["breakpoint"]; ok && val.Type != value.TypeNone {
		// Set breakpoint on word (accepts lit-word, get-word, or word)
		var word string
		switch val.Type {
		case value.TypeLitWord, value.TypeGetWord, value.TypeWord:
			word, _ = val.AsWord()
		default:
			return value.NoneVal(), typeError("debug --breakpoint", "word", val)
		}

		// Validate word exists in current context
		// Check both native registry (in root frame) and user-defined words
		rootFrame := eval.GetFrameByIndex(0)
		_, isNative := rootFrame.Get(word)
		var isUserDefined bool
		if lookup, ok := eval.(wordLookup); ok {
			_, isUserDefined = lookup.Lookup(word)
		}

		if !isNative && !isUserDefined {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDNoValue,
				[3]string{"cannot set breakpoint on unknown word", word, ""},
			)
		}

		id := GlobalDebugger.SetBreakpoint(word)
		return value.IntVal(int64(id)), nil
	}

	if val, ok := refValues["remove"]; ok && val.Type != value.TypeNone {
		// Remove breakpoint by ID
		id, ok := val.AsInteger()
		if !ok {
			return value.NoneVal(), typeError("debug --remove", "integer ID", val)
		}

		// Find and remove breakpoint by ID
		found := false
		GlobalDebugger.mu.Lock()
		for word, bpID := range GlobalDebugger.breakpoints {
			if int64(bpID) == id {
				delete(GlobalDebugger.breakpoints, word)
				found = true
				break
			}
		}
		GlobalDebugger.mu.Unlock()

		if !found {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDNoSuchBreakpoint,
				[3]string{"breakpoint not found", "", ""},
			)
		}
		return value.NoneVal(), nil
	}

	if val, ok := refValues["step"]; ok && ToTruthy(val) {
		GlobalDebugger.EnableStepping()
		return value.NoneVal(), nil
	}

	if val, ok := refValues["next"]; ok && ToTruthy(val) {
		GlobalDebugger.EnableStepping()
		return value.NoneVal(), nil
	}

	if val, ok := refValues["finish"]; ok && ToTruthy(val) {
		GlobalDebugger.DisableStepping()
		return value.NoneVal(), nil
	}

	if val, ok := refValues["continue"]; ok && ToTruthy(val) {
		GlobalDebugger.DisableStepping()
		return value.NoneVal(), nil
	}

	if val, ok := refValues["locals"]; ok && ToTruthy(val) {
		// Return object with local bindings from current frame
		// This requires access to the evaluator's current frame
		// For now, return empty object
		mgr, ok := eval.(frameManager)
		if !ok {
			return value.NoneVal(), verror.NewInternalError(
				"internal-error",
				[3]string{"debug --locals", "frame-manager-unavailable", ""},
			)
		}

		// Create empty object for now (TODO: populate with actual locals)
		fields := []string{}
		initializers := make(map[string][]value.Value)
		return instantiateObject(mgr, eval, -1, nil, fields, initializers)
	}

	if val, ok := refValues["stack"]; ok && ToTruthy(val) {
		// Return block with call stack entries
		// For now, return empty block
		return value.BlockVal([]value.Value{}), nil
	}

	return value.NoneVal(), verror.NewScriptError(
		verror.ErrIDInvalidOperation,
		[3]string{"debug requires a valid refinement", "", ""},
	)
}
