// Package native implements built-in native functions for Viro.
//
// Control flow natives implement conditional execution and iteration.
// Contract per contracts/control-flow.md: when, if, loop, while
package native

import (
	"fmt"
	"io"
	"strconv"
	"time"

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
		return eval.DoBlock(block.Elements, block.Locations())
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
		return eval.DoBlock(block.Elements, block.Locations())
	}

	// Evaluate false-block
	block, _ := value.AsBlockValue(args[2])
	return eval.DoBlock(block.Elements, block.Locations())
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

	// Check for --with-index refinement
	indexVal, hasIndexRef := refValues["with-index"]
	var indexWord string
	if hasIndexRef && indexVal.GetType() != value.TypeNone {
		if !value.IsWord(indexVal.GetType()) {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--with-index requires a word", "", ""},
			)
		}
		indexWord, _ = value.AsWordValue(indexVal)
	}

	block, _ := value.AsBlockValue(args[1])

	// If count is 0, return none without executing
	if count == 0 {
		return value.NewNoneVal(), nil
	}

	currentFrameIdx := eval.CurrentFrameIndex()
	currentFrame := eval.GetFrameByIndex(currentFrameIdx)

	// Execute block count times
	var result core.Value
	var err error
	for i := 0; i < int(count); i++ {
		if hasIndexRef && indexVal.GetType() != value.TypeNone {
			currentFrame.Bind(indexWord, value.NewIntVal(int64(i)))
		}

		result, err = eval.DoBlock(block.Elements, block.Locations())
		if err != nil {
			shouldExit, shouldContinue, propagateErr := handleLoopControlSignal(err)
			if propagateErr != nil {
				return value.NewNoneVal(), propagateErr
			}
			if shouldExit {
				return value.NewNoneVal(), nil
			}
			if shouldContinue {
				continue
			}
		}
	}

	return result, nil
}

// While implements the 'while' conditional loop native.
//
// Contract: while condition [body]
// - Condition can be any value or block
// - If condition is a block, it is re-evaluated before each iteration
// - If condition is not a block, it is evaluated once and must remain constant
// - Body must be a block
// - Loops while condition evaluates to truthy
// - Returns result of last iteration, or none if never executed
func While(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("while", 2, len(args))
	}

	// First argument is the condition (already evaluated if not a block)
	condition := args[0]

	// Second argument must be a block (body)
	if args[1].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("while", "block for body", args[1])
	}

	bodyBlock, _ := value.AsBlockValue(args[1])

	result := value.NewNoneVal()

	// Check if condition is a block (needs re-evaluation each iteration)
	if condition.GetType() == value.TypeBlock {
		conditionBlock, _ := value.AsBlockValue(condition)

		// Loop while condition block evaluates to truthy
		for {
			// Evaluate condition block
			conditionResult, err := eval.DoBlock(conditionBlock.Elements, conditionBlock.Locations())
			if err != nil {
				return value.NewNoneVal(), err
			}

			// Check if condition is truthy
			if !ToTruthy(conditionResult) {
				break
			}

			// Evaluate body block
			result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
			if err != nil {
				shouldExit, shouldContinue, propagateErr := handleLoopControlSignal(err)
				if propagateErr != nil {
					return value.NewNoneVal(), propagateErr
				}
				if shouldExit {
					return value.NewNoneVal(), nil
				}
				if shouldContinue {
					continue
				}
			}
		}
	} else {
		// Condition is not a block, it's already evaluated and constant
		// Loop while condition is truthy (will be infinite if condition is always truthy)
		for ToTruthy(condition) {
			// Evaluate body block
			var err error
			result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
			if err != nil {
				shouldExit, shouldContinue, propagateErr := handleLoopControlSignal(err)
				if propagateErr != nil {
					return value.NewNoneVal(), propagateErr
				}
				if shouldExit {
					return value.NewNoneVal(), nil
				}
				if shouldContinue {
					continue
				}
			}
		}
	}

	return result, nil
}

// Reduce implements the 'reduce' native.
//
// Contract: reduce value
// - If value is a block, evaluates each element and returns a new block with the results
// - If value is not a block, returns the value as-is
// - Evaluates block elements
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
	locations := block.Locations()
	reducedElements := make([]core.Value, 0)

	position := 0

	for position < len(vals) {
		newPos, result, err := eval.EvaluateExpression(vals, locations, position)
		if err != nil {
			return value.NewNoneVal(), err
		}

		reducedElements = append(reducedElements, result)
		position = newPos
	}

	return value.NewBlockVal(reducedElements), nil
}

func Compose(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("compose", 1, len(args))
	}

	if args[0].GetType() == value.TypeParen {
		parenBlock, _ := value.AsBlockValue(args[0])
		return eval.DoBlock(parenBlock.Elements, parenBlock.Locations())
	}

	if args[0].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("compose", "block", args[0])
	}

	block, _ := value.AsBlockValue(args[0])
	vals := block.Elements
	composedElements := make([]core.Value, 0)

	for _, element := range vals {
		if element.GetType() == value.TypeParen {
			parenBlock, _ := value.AsBlockValue(element)
			result, err := eval.DoBlock(parenBlock.Elements, parenBlock.Locations())
			if err != nil {
				return value.NewNoneVal(), err
			}
			composedElements = append(composedElements, result)
		} else {
			composedElements = append(composedElements, element)
		}
	}

	return value.NewBlockVal(composedElements), nil
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

func isLoopControlSignal(err error) (isControl bool, signalType string) {
	if err == nil {
		return false, ""
	}
	verr, ok := err.(*verror.Error)
	if !ok {
		return false, ""
	}
	if verr.Category != verror.ErrThrow {
		return false, ""
	}
	if verr.ID == verror.ErrIDBreak {
		return true, "break"
	}
	if verr.ID == verror.ErrIDContinue {
		return true, "continue"
	}
	return false, ""
}

func extractLevels(err error) int64 {
	verr, ok := err.(*verror.Error)
	if !ok || verr.Args[0] == "" {
		return 1
	}

	levels, parseErr := strconv.ParseInt(verr.Args[0], 10, 64)
	if parseErr != nil {
		return 1
	}

	if levels < 1 {
		return 1
	}

	return levels
}

func handleLoopControlSignal(err error) (shouldExit bool, shouldContinue bool, propagateErr error) {
	isControl, signalType := isLoopControlSignal(err)
	if !isControl {
		return false, false, err
	}

	levels := extractLevels(err)

	if levels > 1 {
		verr, _ := err.(*verror.Error)
		newErr := verror.NewError(
			verror.ErrThrow,
			verr.ID,
			[3]string{fmt.Sprintf("%d", levels-1), "", ""},
		)
		return false, false, newErr
	}

	if signalType == "break" {
		return true, false, nil
	}
	return false, true, nil
}

func parseLevelsRefinement(refValues map[string]core.Value) (int64, error) {
	levels := int64(1)
	if levelsVal, ok := refValues["levels"]; ok && levelsVal.GetType() != value.TypeNone {
		if levelsVal.GetType() != value.TypeInteger {
			return 0, verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--levels requires integer", "", ""},
			)
		}
		levels, _ = value.AsIntValue(levelsVal)
		if levels < 1 {
			return 0, verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"--levels must be >= 1", "", ""},
			)
		}
	}
	return levels, nil
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
			eval.UpdateTraceCache()
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

	// Handle --verbose refinement (Phase 3)
	if verboseVal, ok := refValues["verbose"]; ok && ToTruthy(verboseVal) {
		filters.Verbose = true
	}

	// Handle --step-level refinement (Phase 3)
	if stepLevelVal, ok := refValues["step-level"]; ok && stepLevelVal.GetType() != value.TypeNone {
		if stepLevelVal.GetType() != value.TypeInteger {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--step-level requires integer (0=calls, 1=expressions, 2=all)", "", ""},
			)
		}
		stepLevel, _ := value.AsIntValue(stepLevelVal)
		if stepLevel < 0 || stepLevel > 2 {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"--step-level must be 0, 1, or 2", "", ""},
			)
		}
		filters.StepLevel = int(stepLevel)
	}

	// Handle --include-args refinement (Phase 3)
	if includeArgsVal, ok := refValues["include-args"]; ok && ToTruthy(includeArgsVal) {
		filters.IncludeArgs = true
	}

	// Handle --max-depth refinement (Phase 3)
	if maxDepthVal, ok := refValues["max-depth"]; ok && maxDepthVal.GetType() != value.TypeNone {
		if maxDepthVal.GetType() != value.TypeInteger {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--max-depth requires integer", "", ""},
			)
		}
		maxDepth, _ := value.AsIntValue(maxDepthVal)
		if maxDepth < 0 {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"--max-depth must be non-negative", "", ""},
			)
		}
		filters.MaxDepth = int(maxDepth)
	}

	// Reset step counter when enabling trace (Phase 3)
	trace.GlobalTraceSession.ResetStepCounter()

	trace.GlobalTraceSession.Enable(filters)
	eval.UpdateTraceCache()
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

func Foreach(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("foreach", 3, len(args))
	}

	indexVal, hasIndexRef := refValues["with-index"]
	var indexWord string
	if hasIndexRef && indexVal.GetType() != value.TypeNone {
		if !value.IsWord(indexVal.GetType()) {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--with-index requires a word", "", ""},
			)
		}
		indexWord, _ = value.AsWordValue(indexVal)
	}

	seriesVal := args[0]

	if !value.IsSeries(seriesVal.GetType()) && seriesVal.GetType() != value.TypeObject {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"foreach requires series or object type (block!, string!, binary!, object!)", "", ""},
		)
	}

	varsArg := args[1]

	if args[2].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("foreach", "block for body", args[2])
	}

	bodyBlock, _ := value.AsBlockValue(args[2])

	var varNames []string

	if value.IsWord(varsArg.GetType()) {
		varName, _ := value.AsWordValue(varsArg)
		varNames = []string{varName}
	} else if varsArg.GetType() == value.TypeBlock {
		wordBlock, _ := value.AsBlockValue(varsArg)
		if len(wordBlock.Elements) == 0 {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"foreach vars block must contain at least one word", "", ""},
			)
		}
		varNames = make([]string, len(wordBlock.Elements))
		for i, varElement := range wordBlock.Elements {
			if !value.IsWord(varElement.GetType()) {
				return value.NewNoneVal(), verror.NewScriptError(
					verror.ErrIDInvalidOperation,
					[3]string{"foreach vars must be a word or block of words", "", ""},
				)
			}
			varName, _ := value.AsWordValue(varElement)
			varNames[i] = varName
		}
	} else {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"foreach vars must be a word or block of words", "", ""},
		)
	}

	currentFrameIdx := eval.CurrentFrameIndex()
	currentFrame := eval.GetFrameByIndex(currentFrameIdx)

	numVars := len(varNames)
	var result core.Value
	var err error
	var iteration int

	if seriesVal.GetType() == value.TypeObject {
		obj, ok := seriesVal.(*value.ObjectInstance)
		if !ok {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"value is not a valid object", "", ""},
			)
		}

		bindings := obj.GetAllFieldsWithProto()
		if len(bindings) == 0 {
			return value.NewNoneVal(), nil
		}

		for _, binding := range bindings {
			for j := 0; j < numVars; j++ {
				if j == 0 {
					currentFrame.Bind(varNames[j], value.NewStrVal(binding.Symbol))
				} else if j == 1 {
					fieldVal, _ := obj.GetFieldWithProto(binding.Symbol)
					currentFrame.Bind(varNames[j], fieldVal)
				} else {
					currentFrame.Bind(varNames[j], value.NewNoneVal())
				}
			}

			if hasIndexRef && indexVal.GetType() != value.TypeNone {
				currentFrame.Bind(indexWord, value.NewIntVal(int64(iteration)))
			}

			result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())

			if err != nil {
				shouldExit, shouldContinue, propagateErr := handleLoopControlSignal(err)
				if propagateErr != nil {
					return value.NewNoneVal(), propagateErr
				}
				if shouldExit {
					return value.NewNoneVal(), nil
				}
				if shouldContinue {
					iteration++
					continue
				}
			}
			iteration++
		}
	} else {
		series, ok := seriesVal.(value.Series)
		if !ok {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"value does not implement Series interface", "", ""},
			)
		}

		startIndex := series.GetIndex()
		length := series.Length()

		if length == 0 || startIndex >= length {
			return value.NewNoneVal(), nil
		}

		for i := startIndex; i < length; {
			for j := 0; j < numVars; j++ {
				if i < length {
					element := series.ElementAt(i)
					currentFrame.Bind(varNames[j], element)
					i++
				} else {
					currentFrame.Bind(varNames[j], value.NewNoneVal())
				}
			}

			if hasIndexRef && indexVal.GetType() != value.TypeNone {
				currentFrame.Bind(indexWord, value.NewIntVal(int64(iteration)))
			}

			result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())

			if err != nil {
				shouldExit, shouldContinue, propagateErr := handleLoopControlSignal(err)
				if propagateErr != nil {
					return value.NewNoneVal(), propagateErr
				}
				if shouldExit {
					return value.NewNoneVal(), nil
				}
				if shouldContinue {
					iteration++
					continue
				}
			}
			iteration++
		}
	}

	return result, nil
}

func Do(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("do", 1, len(args))
	}

	val := args[0]

	nextVal, hasNext := refValues["next"]
	if hasNext && nextVal.GetType() != value.TypeNone {
		var wordName string
		switch nextVal.GetType() {
		case value.TypeLitWord, value.TypeGetWord, value.TypeWord:
			wordName, _ = value.AsWordValue(nextVal)
		default:
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--next requires a word", "", ""},
			)
		}

		if val.GetType() != value.TypeBlock {
			newPos, result, err := eval.EvaluateExpression([]core.Value{val}, nil, 0)
			if err != nil {
				return value.NewNoneVal(), err
			}
			if newPos > 0 {
				return result, nil
			}
			return value.NewNoneVal(), nil
		}

		block, _ := value.AsBlockValue(val)
		vals := block.Elements
		locations := block.Locations()
		startIndex := block.Index

		currentFrameIdx := eval.CurrentFrameIndex()
		currentFrame := eval.GetFrameByIndex(currentFrameIdx)

		if startIndex >= len(vals) {
			nextBlock := block.Clone()
			nextBlock.SetIndex(startIndex)
			currentFrame.Bind(wordName, nextBlock.(core.Value))
			return value.NewNoneVal(), nil
		}

		newPos, result, err := eval.EvaluateExpression(vals, locations, startIndex)
		if err != nil {
			return value.NewNoneVal(), err
		}

		nextBlock := block.Clone()
		nextBlock.SetIndex(newPos)
		currentFrame.Bind(wordName, nextBlock.(core.Value))

		return result, nil
	}

	if val.GetType() == value.TypeBlock {
		block, _ := value.AsBlockValue(val)
		startIndex := block.Index
		if startIndex >= len(block.Elements) {
			return value.NewNoneVal(), nil
		}
		locations := block.Locations()
		if len(locations) > startIndex {
			locations = locations[startIndex:]
		} else {
			locations = nil
		}
		return eval.DoBlock(block.Elements[startIndex:], locations)
	}

	newPos, result, err := eval.EvaluateExpression([]core.Value{val}, nil, 0)
	if err != nil {
		return value.NewNoneVal(), err
	}
	if newPos > 0 {
		return result, nil
	}

	return value.NewNoneVal(), nil
}

// Debug implements the 'debug' native for debugger control (Feature 002, FR-021).
//
// Contract: debug --on | --off | --breakpoint word | --remove id
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

	return value.NewNoneVal(), verror.NewScriptError(
		verror.ErrIDInvalidOperation,
		[3]string{"debug requires a valid refinement", "", ""},
	)
}

func Break(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("break", 0, len(args))
	}

	levels, err := parseLevelsRefinement(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	return value.NewNoneVal(), verror.NewError(
		verror.ErrThrow,
		verror.ErrIDBreak,
		[3]string{fmt.Sprintf("%d", levels), "", ""},
	)
}

func Continue(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("continue", 0, len(args))
	}

	levels, err := parseLevelsRefinement(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	return value.NewNoneVal(), verror.NewError(
		verror.ErrThrow,
		verror.ErrIDContinue,
		[3]string{fmt.Sprintf("%d", levels), "", ""},
	)
}

func Return(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) > 1 {
		return value.NewNoneVal(), arityError("return", 1, len(args))
	}

	returnVal := value.NewNoneVal()
	if len(args) == 1 {
		returnVal = args[0]
	}

	return value.NewNoneVal(), eval.NewReturnSignal(returnVal)
}

func Probe(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("probe", 1, len(args))
	}

	val := args[0]
	molded := val.Mold()

	writer := eval.GetOutputWriter()
	if writer != io.Discard {
		fmt.Fprintf(writer, "== %s\n", molded)
	} else {
		if trace.GlobalTraceSession != nil {
			event := trace.TraceEvent{
				Timestamp: time.Now(),
				Value:     molded,
				Word:      "probe",
				EventType: "debug",
			}
			trace.GlobalTraceSession.Emit(event)
		} else {
			fmt.Fprintf(eval.GetErrorWriter(), "== %s\n", molded)
		}
	}

	return val, nil
}
