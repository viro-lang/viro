// Package eval provides the core evaluation engine for Viro.
//
// Architecture:
// - Evaluator struct holds the stack and execution state
// - Do_Next evaluates a single value
// - Do_Blk evaluates a block of values in sequence
// - Type-based dispatch routes evaluation by value type
//
// Per Constitution Principle III: Explicit type dispatch, no polymorphism.
// Package eval implements the core evaluation engine for the Viro interpreter.
//
// The evaluator uses type-based dispatch to evaluate REBOL-style expressions.
// It supports literals, words, functions, blocks, and parens with proper
// operator precedence and scoping rules.
//
// Key functions:
//   - Do_Next: Evaluate a single value based on its type
//   - Do_Blk: Evaluate a sequence of values (block)
//
// The evaluator maintains a stack for data and frames, providing lexical
// scoping for variables and functions.
package eval

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/stack"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Evaluator is the core evaluation engine.
//
// Design per Constitution Principle IV: Index-based access to stack/frames.
// Stack holds both data and frames in a unified structure.
type Evaluator struct {
	Stack      *stack.Stack
	Frames     []*frame.Frame
	frameStore []*frame.Frame
	frameIndex map[*frame.Frame]int
	captured   map[int]bool
	callStack  []string
}

// NewEvaluator creates a new evaluation engine with an empty stack.
func NewEvaluator() *Evaluator {
	global := frame.NewFrame(frame.FrameClosure, -1)
	global.Name = "(top level)"
	e := &Evaluator{
		Stack:      stack.NewStack(1024),
		Frames:     []*frame.Frame{global},
		frameStore: []*frame.Frame{global},
		frameIndex: make(map[*frame.Frame]int),
		captured:   make(map[int]bool),
		callStack:  []string{"(top level)"},
	}
	e.frameIndex[global] = 0
	e.captured[0] = true
	return e
}

// currentFrame returns the active frame (top of frame stack).
func (e *Evaluator) currentFrame() *frame.Frame {
	if len(e.Frames) == 0 {
		return nil
	}
	return e.Frames[len(e.Frames)-1]
}

// currentFrameIndex returns the store index for the active frame.
func (e *Evaluator) currentFrameIndex() int {
	frame := e.currentFrame()
	if frame == nil {
		return -1
	}
	if idx, ok := e.frameIndex[frame]; ok {
		return idx
	}
	return -1
}

// getFrameByIndex retrieves frame from store by index.
func (e *Evaluator) getFrameByIndex(idx int) *frame.Frame {
	if idx < 0 || idx >= len(e.frameStore) {
		return nil
	}
	return e.frameStore[idx]
}

// pushFrame registers a new frame as active and stores it for closure lookups.
func (e *Evaluator) pushFrame(f *frame.Frame) int {
	idx, ok := e.frameIndex[f]
	if !ok {
		idx = len(e.frameStore)
		e.frameStore = append(e.frameStore, f)
		e.frameIndex[f] = idx
	}
	e.Frames = append(e.Frames, f)
	return idx
}

// popFrame removes the active frame and returns its store index.
func (e *Evaluator) popFrame() int {
	if len(e.Frames) == 0 {
		return -1
	}
	frm := e.Frames[len(e.Frames)-1]
	e.Frames = e.Frames[:len(e.Frames)-1]
	idx := e.frameIndex[frm]
	if !e.captured[idx] {
		// Release non-captured frames for GC by clearing store entry
		e.frameStore[idx] = nil
		delete(e.frameIndex, frm)
	} else if frm.Type != frame.FrameClosure {
		frm.Type = frame.FrameClosure
	}
	return idx
}

func (e *Evaluator) pushCall(name string) {
	if name == "" {
		name = "(anonymous)"
	}
	e.callStack = append(e.callStack, name)
}

func (e *Evaluator) popCall() {
	if len(e.callStack) <= 1 {
		return
	}
	e.callStack = e.callStack[:len(e.callStack)-1]
}

func (e *Evaluator) captureCallStack() []string {
	if len(e.callStack) == 0 {
		return []string{}
	}
	where := make([]string, len(e.callStack))
	for i := range e.callStack {
		where[i] = e.callStack[len(e.callStack)-1-i]
	}
	return where
}

func (e *Evaluator) annotateError(err *verror.Error, vals []value.Value, idx int) *verror.Error {
	if err == nil {
		return nil
	}
	if idx >= 0 && idx < len(vals) && err.Near == "" {
		err.SetNear(verror.CaptureNear(vals, idx))
	}
	if len(err.Where) == 0 {
		where := e.captureCallStack()
		if len(where) > 0 {
			err.SetWhere(where)
		}
	}
	return err
}

func functionDisplayName(fn *value.FunctionValue) string {
	if fn == nil || fn.Name == "" {
		return "(anonymous)"
	}
	return fn.Name
}

// MarkFrameCaptured marks a frame as captured for closure usage.
func (e *Evaluator) MarkFrameCaptured(idx int) {
	if idx >= 0 {
		e.captured[idx] = true
	}
}

// CurrentFrameIndex exposes the active frame index (implements frameProvider).
func (e *Evaluator) CurrentFrameIndex() int {
	return e.currentFrameIndex()
}

// RegisterFrame adds a frame to the evaluator's frame store and returns its index.
// Feature 002: Used by object native to register object frames.
// NOTE: This does NOT push the frame onto the active frame stack (e.Frames).
// Use PushFrameContext to make it active.
func (e *Evaluator) RegisterFrame(f *frame.Frame) int {
	// Check if frame is already registered
	if idx, ok := e.frameIndex[f]; ok {
		return idx
	}

	// Add to store
	idx := len(e.frameStore)
	e.frameStore = append(e.frameStore, f)
	e.frameIndex[f] = idx
	return idx
}

// GetFrameByIndex retrieves a frame from the store by its index.
// Feature 002: Used by path traversal to access object frames.
func (e *Evaluator) GetFrameByIndex(idx int) *frame.Frame {
	return e.getFrameByIndex(idx)
}

// PushFrameContext temporarily makes a frame the active context for evaluation.
// Feature 002: Used by object native to evaluate initializers in object context.
// Returns the frame index.
func (e *Evaluator) PushFrameContext(f *frame.Frame) int {
	return e.pushFrame(f)
}

// PopFrameContext removes the active frame context.
// Feature 002: Used after object initialization completes.
func (e *Evaluator) PopFrameContext() {
	e.popFrame()
}

// lookup searches for a word through the active frame chain.
func (e *Evaluator) lookup(symbol string) (value.Value, bool) {
	frame := e.currentFrame()
	for frame != nil {
		if val, ok := frame.Get(symbol); ok {
			return val, true
		}
		if frame.Parent == -1 {
			break
		}
		frame = e.getFrameByIndex(frame.Parent)
	}
	return value.NoneVal(), false
}

// Do_Next evaluates a single value and returns the result.
//
// Contract per data-model.md §3:
// - Integers, strings, logic, none: Return self (literals)
// - Blocks [...]: Return self (deferred evaluation)
// - Parens (...): Evaluate contents and return result
// - Words: Look up in frame and evaluate result
// - Set-words (word:): Bind next value
// - Get-words (:word): Fetch value without evaluation
// - Lit-words ('word): Return word as-is
//
// Returns the evaluated value and any error encountered.
func (e *Evaluator) Do_Next(val value.Value) (value.Value, *verror.Error) {
	// Trace instrumentation (Feature 002, T025)
	// Per FR-015: emit trace events when tracing is enabled
	var traceStart time.Time
	var traceWord string
	if native.GlobalTraceSession != nil && native.GlobalTraceSession.IsEnabled() {
		traceStart = time.Now()
		if val.Type.IsWord() {
			if w, ok := val.AsWord(); ok {
				traceWord = w
			}
		}
	}

	// Type-based dispatch per Constitution Principle III
	var result value.Value
	var err *verror.Error

	switch val.Type {
	case value.TypeInteger, value.TypeString, value.TypeLogic, value.TypeNone:
		// Literals evaluate to themselves
		result, err = val, nil

	case value.TypeBlock:
		// Blocks return self (deferred evaluation)
		result, err = val, nil

	case value.TypeParen:
		// Parens evaluate contents
		result, err = e.evalParen(val)

	case value.TypeWord:
		// Words look up in frame
		result, err = e.evalWord(val)

	case value.TypeSetWord:
		// Set-words are handled in Do_Blk (need next value from sequence)
		// If we reach here, it's a set-word in isolation (error)
		wordStr, _ := val.AsWord()
		result, err = value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})

	case value.TypeGetWord:
		// Get-words fetch without evaluation
		result, err = e.evalGetWord(val)

	case value.TypeLitWord:
		// Lit-words return as word
		// Extract the word symbol and return as a word
		result, err = value.WordVal(val.Payload.(string)), nil

	case value.TypeFunction:
		// Functions return self (they're values)
		result, err = val, nil

	// Feature 002: New value types (T024)
	case value.TypeDecimal:
		// Decimals evaluate to themselves (literals)
		result, err = val, nil

	case value.TypeObject:
		// Objects evaluate to themselves
		result, err = val, nil

	case value.TypePort:
		// Ports evaluate to themselves (handles)
		result, err = val, nil

	case value.TypePath:
		// Paths are evaluated by path evaluator (T091)
		result, err = e.evalPath(val)

	default:
		result, err = value.NoneVal(), verror.NewInternalError("unknown value type in Do_Next", [3]string{})
	}

	// Emit trace event if tracing is enabled
	if native.GlobalTraceSession != nil && native.GlobalTraceSession.IsEnabled() && traceWord != "" {
		duration := time.Since(traceStart)
		native.GlobalTraceSession.Emit(native.TraceEvent{
			Timestamp: traceStart,
			Value:     result.String(),
			Word:      traceWord,
			Duration:  duration.Nanoseconds(),
			Depth:     len(e.callStack),
		})
	}

	return result, err
}

// Do_Blk evaluates a block of values in sequence.
//
// Contract per data-model.md §3: Evaluates each value left-to-right.
// Returns the last result, or none if block is empty.
//
// Special handling for set-words: evaluates next value and binds result.
func (e *Evaluator) Do_Blk(vals []value.Value) (value.Value, *verror.Error) {
	if len(vals) == 0 {
		return value.NoneVal(), nil
	}

	var result value.Value
	var err *verror.Error

	// Use index-based loop to handle set-words that consume next value
	for i := 0; i < len(vals); i++ {
		val := vals[i]

		// Special case: set-word consumes next value
		if val.Type == value.TypeSetWord {
			result, err = e.evalSetWord(val, vals, &i)
			if err != nil {
				return value.NoneVal(), e.annotateError(err, vals, i)
			}
			continue
		}

		// Special case: word that might be a native or user-defined function call
		if val.Type == value.TypeWord {
			wordStr, ok := val.AsWord()
			if ok {
				if nativeInfo, found := native.Lookup(wordStr); found {
					// This is a native function call - collect arguments
					startIdx := i
					result, err = e.callNative(wordStr, nativeInfo, vals, &i)
					if err != nil {
						return value.NoneVal(), e.annotateError(err, vals, startIdx)
					}
					continue
				}

				if resolved, found := e.lookup(wordStr); found && resolved.Type == value.TypeFunction {
					fn, _ := resolved.AsFunction()
					startIdx := i
					result, err = e.invokeFunctionFromSequence(fn, vals, &i)
					if err != nil {
						return value.NoneVal(), e.annotateError(err, vals, startIdx)
					}
					continue
				}
			}
		}

		result, err = e.Do_Next(val)
		if err != nil {
			return value.NoneVal(), e.annotateError(err, vals, i)
		}
	}

	return result, nil
}

// evalParen evaluates the contents of a paren and returns the result.
// Special handling: if paren starts with a word that's a native, treat it as a function call.
func (e *Evaluator) evalParen(val value.Value) (value.Value, *verror.Error) {
	block, ok := val.AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("paren value does not contain BlockValue", [3]string{})
	}

	if len(block.Elements) >= 1 && block.Elements[0].Type == value.TypeWord {
		wordStr, ok := block.Elements[0].AsWord()
		if ok {
			if nativeInfo, found := native.Lookup(wordStr); found {
				return e.callNativeFromSlice(wordStr, nativeInfo, block.Elements[1:])
			}

			if resolved, found := e.lookup(wordStr); found && resolved.Type == value.TypeFunction {
				fn, _ := resolved.AsFunction()
				return e.invokeFunctionWithTokens(fn, block.Elements[1:])
			}
		}
	}

	return e.Do_Blk(block.Elements)
}

func (e *Evaluator) callNative(name string, info *native.NativeInfo, vals []value.Value, idx *int) (value.Value, *verror.Error) {
	startIdx := *idx
	args := make([]value.Value, 0, info.Arity)
	for j := 0; j < info.Arity; j++ {
		tokenIdx := *idx + 1 + j
		if tokenIdx >= len(vals) {
			err := verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{name, fmt.Sprintf("%d", info.Arity), fmt.Sprintf("%d", len(args))},
			)
			return value.NoneVal(), e.annotateError(err, vals, startIdx)
		}
		arg, argErr := e.Do_Next(vals[tokenIdx])
		if argErr != nil {
			return value.NoneVal(), e.annotateError(argErr, vals, tokenIdx)
		}
		args = append(args, arg)
	}

	*idx += info.Arity
	result, err := native.Call(info, args, e)
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, startIdx)
	}
	return result, nil
}

func (e *Evaluator) callNativeFromSlice(name string, info *native.NativeInfo, tokens []value.Value) (value.Value, *verror.Error) {
	context := append([]value.Value{value.WordVal(name)}, tokens...)
	if len(tokens) != info.Arity {
		err := verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{name, fmt.Sprintf("%d", info.Arity), fmt.Sprintf("%d", len(tokens))},
		)
		return value.NoneVal(), e.annotateError(err, context, 0)
	}

	args := make([]value.Value, 0, info.Arity)
	for k, token := range tokens {
		arg, err := e.Do_Next(token)
		if err != nil {
			return value.NoneVal(), e.annotateError(err, context, k+1)
		}
		args = append(args, arg)
	}

	result, err := native.Call(info, args, e)
	if err != nil {
		return value.NoneVal(), e.annotateError(err, context, 0)
	}
	return result, nil
}

func (e *Evaluator) invokeFunctionFromSequence(fn *value.FunctionValue, vals []value.Value, idx *int) (value.Value, *verror.Error) {
	name := functionDisplayName(fn)
	e.pushCall(name)
	defer e.popCall()

	tokens := vals[*idx+1:]
	posArgs, refValues, consumed, err := e.collectFunctionArgs(fn, tokens)
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, *idx)
	}

	*idx += consumed
	result, execErr := e.executeFunction(fn, posArgs, refValues)
	if execErr != nil {
		return value.NoneVal(), execErr
	}
	return result, nil
}

func (e *Evaluator) invokeFunctionWithTokens(fn *value.FunctionValue, tokens []value.Value) (value.Value, *verror.Error) {
	name := functionDisplayName(fn)
	context := append([]value.Value{value.WordVal(name)}, tokens...)
	e.pushCall(name)
	defer e.popCall()

	posArgs, refValues, consumed, err := e.collectFunctionArgs(fn, tokens)
	if err != nil {
		return value.NoneVal(), e.annotateError(err, context, 0)
	}

	if consumed != len(tokens) {
		err := verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{name, fmt.Sprintf("%d", len(posArgs)), fmt.Sprintf("%d", len(posArgs)+(len(tokens)-consumed))},
		)
		return value.NoneVal(), e.annotateError(err, context, 0)
	}

	result, execErr := e.executeFunction(fn, posArgs, refValues)
	if execErr != nil {
		return value.NoneVal(), execErr
	}
	return result, nil
}

func (e *Evaluator) collectFunctionArgs(fn *value.FunctionValue, tokens []value.Value) ([]value.Value, map[string]value.Value, int, *verror.Error) {
	displayName := functionDisplayName(fn)
	positional := make([]value.ParamSpec, 0, len(tokens))
	refSpecs := make(map[string]value.ParamSpec)
	refValues := make(map[string]value.Value)
	refProvided := make(map[string]bool)

	for _, spec := range fn.Params {
		if spec.Refinement {
			refSpecs[spec.Name] = spec
			if spec.TakesValue {
				refValues[spec.Name] = value.NoneVal()
			} else {
				refValues[spec.Name] = value.LogicVal(false)
			}
			continue
		}
		positional = append(positional, spec)
	}

	posArgs := make([]value.Value, len(positional))
	consumed := 0
	posIndex := 0

	for consumed < len(tokens) {
		token := tokens[consumed]

		if token.Type == value.TypeWord {
			wordName, _ := token.AsWord()
			if strings.HasPrefix(wordName, "--") {
				refName := strings.TrimPrefix(wordName, "--")
				spec, exists := refSpecs[refName]
				if !exists {
					return nil, nil, 0, verror.NewScriptError(
						verror.ErrIDInvalidOperation,
						[3]string{fmt.Sprintf("Unknown refinement: --%s", refName), "", ""},
					)
				}
				if refProvided[refName] {
					return nil, nil, 0, verror.NewScriptError(
						verror.ErrIDInvalidOperation,
						[3]string{fmt.Sprintf("Duplicate refinement: --%s", refName), "", ""},
					)
				}

				if spec.TakesValue {
					if consumed+1 >= len(tokens) {
						return nil, nil, 0, verror.NewScriptError(
							verror.ErrIDInvalidOperation,
							[3]string{fmt.Sprintf("Refinement --%s requires a value", refName), "", ""},
						)
					}
					valueToken := tokens[consumed+1]
					arg, err := e.Do_Next(valueToken)
					if err != nil {
						return nil, nil, 0, err
					}
					refValues[refName] = arg
					consumed += 2
				} else {
					refValues[refName] = value.LogicVal(true)
					consumed++
				}

				refProvided[refName] = true
				continue
			}
		}

		if posIndex >= len(positional) {
			break
		}

		arg, err := e.Do_Next(token)
		if err != nil {
			return nil, nil, 0, err
		}
		posArgs[posIndex] = arg
		posIndex++
		consumed++
	}

	if posIndex < len(positional) {
		return nil, nil, 0, verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{displayName, fmt.Sprintf("%d", len(positional)), fmt.Sprintf("%d", posIndex)},
		)
	}

	return posArgs, refValues, consumed, nil
}

func (e *Evaluator) executeFunction(fn *value.FunctionValue, posArgs []value.Value, refinements map[string]value.Value) (value.Value, *verror.Error) {
	parent := fn.Parent
	if parent == -1 {
		parent = 0
	}

	frame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, parent, len(fn.Params))
	frame.Name = functionDisplayName(fn)
	e.pushFrame(frame)
	defer e.popFrame()

	posIndex := 0
	for _, spec := range fn.Params {
		if spec.Refinement {
			val, ok := refinements[spec.Name]
			if !ok {
				if spec.TakesValue {
					val = value.NoneVal()
				} else {
					val = value.LogicVal(false)
				}
			}
			frame.Bind(spec.Name, val)
			continue
		}

		frame.Bind(spec.Name, posArgs[posIndex])
		posIndex++
	}

	if fn.Body == nil {
		return value.NoneVal(), verror.NewInternalError("function body missing", [3]string{})
	}

	result, err := e.Do_Blk(fn.Body.Elements)
	if err != nil {
		return value.NoneVal(), err
	}

	return result, nil
}

// evalWord looks up a word in the current frame and evaluates the result.
func (e *Evaluator) evalWord(val value.Value) (value.Value, *verror.Error) {
	wordStr, ok := val.AsWord()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("word value does not contain string", [3]string{})
	}

	// Check if it's a native function - if so, return the word itself
	// (it will be called when it appears in function position)
	if _, ok := native.Lookup(wordStr); ok {
		return val, nil // Return the word itself, not evaluated yet
	}

	result, ok := e.lookup(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	// Evaluate the result (words evaluate to their values, then re-evaluate)
	return e.Do_Next(result)
}

// evalSetWord handles set-word evaluation: binds next value to word.
//
// Contract per data-model.md §4:
// - Set-Word: Evaluate next expression, store in Binding frame at Symbol
//
// Parameters:
// - val: The set-word value (contains word symbol)
// - vals: The full sequence of values being evaluated
// - i: Pointer to current index (will be advanced to skip consumed value)
func (e *Evaluator) evalSetWord(val value.Value, vals []value.Value, i *int) (value.Value, *verror.Error) {
	// Extract word symbol from set-word
	wordStr, ok := val.AsWord()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("set-word value does not contain string", [3]string{})
	}

	// Check if there's a next value to evaluate
	if *i+1 >= len(vals) {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})
	}

	// Check if this is a set-path (contains dot separator)
	if strings.Contains(wordStr, ".") {
		return e.evalSetPath(wordStr, vals, i)
	}

	currentFrame := e.currentFrame()
	if currentFrame == nil {
		currentFrame = frame.NewFrame(frame.FrameFunctionArgs, -1)
		e.pushFrame(currentFrame)
	}

	// Advance to next value and evaluate it
	*i++
	nextVal := vals[*i]

	var (
		result value.Value
		err    *verror.Error
	)

	if nextVal.Type == value.TypeWord {
		if wordStr, ok := nextVal.AsWord(); ok {
			if nativeInfo, found := native.Lookup(wordStr); found {
				result, err = e.callNative(wordStr, nativeInfo, vals, i)
			} else if resolved, found := e.lookup(wordStr); found && resolved.Type == value.TypeFunction {
				fn, _ := resolved.AsFunction()
				result, err = e.invokeFunctionFromSequence(fn, vals, i)
			} else {
				result, err = e.Do_Next(nextVal)
			}
		} else {
			result, err = e.Do_Next(nextVal)
		}
	} else {
		result, err = e.Do_Next(nextVal)
	}
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, *i)
	}

	if result.Type == value.TypeFunction {
		if fnVal, ok := result.AsFunction(); ok && fnVal.Name == "" {
			fnVal.Name = wordStr
		}
	}

	// Bind the result to the word in current frame
	currentFrame.Bind(wordStr, result)

	// Return the bound value
	return result, nil
}

// evalGetWord looks up a word in the current frame WITHOUT evaluating the result.
//
// Contract per data-model.md §4:
// - Get-Word: Look up Symbol in Binding frame, return without evaluation
//
// This is the difference from regular word evaluation which re-evaluates the result.
func (e *Evaluator) evalGetWord(val value.Value) (value.Value, *verror.Error) {
	wordStr, ok := val.AsWord()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("get-word value does not contain string", [3]string{})
	}

	result, ok := e.lookup(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	// Return without re-evaluation (key difference from evalWord)
	return result, nil
}

// evalPath evaluates a path expression by traversing segments (T091).
//
// Contract per contracts/objects.md §3:
// 1. Resolve base value from first segment (word lookup)
// 2. For each subsequent segment:
//   - Object field: lookup in frame, check parent chain if not found
//   - Block/String index: 1-based indexing with bounds checking
//   - None encountered: error (none-path)
//
// 3. Return final value or error
//
// Error cases:
// - none-path: path traverses through none value
// - no-such-field: object field not found
// - index-out-of-range: block/string index invalid
// - path-type-mismatch: path applied to unsupported type
func (e *Evaluator) evalPath(val value.Value) (value.Value, *verror.Error) {

	if val.Type != value.TypePath {
		return value.NoneVal(), verror.NewInternalError("evalPath called with non-path type", [3]string{val.Type.String(), "", ""})
	}

	path, ok := val.AsPath()
	if !ok {

		return value.NoneVal(), verror.NewInternalError("path value does not contain PathExpression - payload type mismatch", [3]string{fmt.Sprintf("payload=%T", val.Payload), "", ""})
	}

	if len(path.Segments) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"empty path", "", ""})
	}

	// Start with the base value - resolve first segment
	firstSeg := path.Segments[0]
	var base value.Value

	if firstSeg.Type == value.PathSegmentWord {
		wordStr, ok := firstSeg.Value.(string)
		if !ok {
			return value.NoneVal(), verror.NewInternalError("word segment does not contain string", [3]string{})
		}

		base, ok = e.lookup(wordStr)
		if !ok {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
		}
	} else if firstSeg.Type == value.PathSegmentIndex {
		// Path starts with a literal number (e.g., 1.field)
		// This is valid for reading but will fail for assignment (immutable target)
		num, ok := firstSeg.Value.(int64)
		if !ok {
			return value.NoneVal(), verror.NewInternalError("index segment does not contain int64", [3]string{})
		}
		base = value.IntVal(num)
	} else {
		return value.NoneVal(), verror.NewInternalError("unexpected first segment type", [3]string{fmt.Sprintf("%v", firstSeg.Type), "", ""})
	} // Traverse remaining segments
	for i := 1; i < len(path.Segments); i++ {
		seg := path.Segments[i]

		// Check for none mid-path (none-path error)
		if base.Type == value.TypeNone {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDNonePath, [3]string{})
		}

		switch seg.Type {
		case value.PathSegmentWord:
			// Object field access
			if base.Type != value.TypeObject {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{base.Type.String(), "", ""})
			}

			obj, _ := base.AsObject()
			fieldName, ok := seg.Value.(string)
			if !ok {
				return value.NoneVal(), verror.NewInternalError("word segment does not contain string", [3]string{})
			}

			// Look up field in object's frame
			objFrame := e.getFrameByIndex(obj.FrameIndex)
			if objFrame == nil {
				return value.NoneVal(), verror.NewInternalError("object frame not found", [3]string{})
			}

			fieldVal, found := objFrame.Get(fieldName)
			if !found {
				// Check parent chain
				if obj.Parent >= 0 {
					parentFrame := e.getFrameByIndex(obj.Parent)
					if parentFrame != nil {
						fieldVal, found = parentFrame.Get(fieldName)
					}
				}

				if !found {
					return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{fieldName, "", ""})
				}
			}

			base = fieldVal

		case value.PathSegmentIndex:
			// Block or string indexing (1-based)
			index, ok := seg.Value.(int64)
			if !ok {
				return value.NoneVal(), verror.NewInternalError("index segment does not contain int64", [3]string{})
			}

			if base.Type == value.TypeBlock {
				block, _ := base.AsBlock()
				// 1-based indexing
				if index < 1 || index > int64(len(block.Elements)) {
					return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range for block of length %d", index, len(block.Elements)), "", ""})
				}
				base = block.Elements[index-1]

			} else if base.Type == value.TypeString {
				str, _ := base.AsString()
				runes := []rune(str.String())
				if index < 1 || index > int64(len(runes)) {
					return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range for string of length %d", index, len(runes)), "", ""})
				}
				// Return character as string
				base = value.StrVal(string(runes[index-1]))

			} else {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index requires block or string type", "", ""})
			}

		default:
			return value.NoneVal(), verror.NewInternalError("unsupported path segment type", [3]string{})
		}
	}

	return base, nil
}

// evalSetPath handles path assignment (set-path) like obj.field: value (T091).
//
// Contract per contracts/objects.md §3:
// - Parse path string to extract segments
// - Traverse to penultimate segment
// - Update final segment in target container (object frame or block)
// - Error if attempting to assign to literal or immutable target
func (e *Evaluator) evalSetPath(pathStr string, vals []value.Value, i *int) (value.Value, *verror.Error) {
	// Parse path string into segments
	parts := strings.Split(pathStr, ".")
	if len(parts) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"set-path requires at least 2 segments", "", ""})
	}

	// Evaluate the value to assign
	*i++
	if *i >= len(vals) {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-path without value", "", ""})
	}

	var result value.Value
	var err *verror.Error
	nextVal := vals[*i]

	if nextVal.Type == value.TypeWord {
		if wordStr, ok := nextVal.AsWord(); ok {
			if nativeInfo, found := native.Lookup(wordStr); found {
				result, err = e.callNative(wordStr, nativeInfo, vals, i)
			} else if resolved, found := e.lookup(wordStr); found && resolved.Type == value.TypeFunction {
				fn, _ := resolved.AsFunction()
				result, err = e.invokeFunctionFromSequence(fn, vals, i)
			} else {
				result, err = e.Do_Next(nextVal)
			}
		} else {
			result, err = e.Do_Next(nextVal)
		}
	} else {
		result, err = e.Do_Next(nextVal)
	}
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, *i)
	}

	// Resolve base value
	baseName := parts[0]
	var base value.Value
	var ok bool

	// Check if base is a literal number (e.g., "1" in "1.field: 100")
	if _, parseErr := strconv.ParseInt(baseName, 10, 64); parseErr == nil {
		// Path starts with literal - this is an immutable target
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{pathStr, "", ""})
	}

	base, ok = e.lookup(baseName)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{baseName, "", ""})
	}

	// Traverse to penultimate container
	for j := 1; j < len(parts)-1; j++ {
		part := parts[j]

		// Check for none mid-path
		if base.Type == value.TypeNone {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDNonePath, [3]string{"cannot traverse path through none value", "", ""})
		}

		// Try as index first
		if index, parseErr := fmt.Sscanf(part, "%d", new(int64)); parseErr == nil && index == 1 {
			var idx int64
			fmt.Sscanf(part, "%d", &idx)

			if base.Type == value.TypeBlock {
				block, _ := base.AsBlock()
				if idx < 1 || idx > int64(len(block.Elements)) {
					return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range", idx), "", ""})
				}
				base = block.Elements[idx-1]
			} else {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index requires block type", "", ""})
			}
		} else {
			// It's a word (object field)
			if base.Type != value.TypeObject {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"field access requires object type", "", ""})
			}

			obj, _ := base.AsObject()
			objFrame := e.getFrameByIndex(obj.FrameIndex)
			if objFrame == nil {
				return value.NoneVal(), verror.NewInternalError("object frame not found", [3]string{})
			}

			fieldVal, found := objFrame.Get(part)
			if !found {
				// Check parent chain
				if obj.Parent >= 0 {
					parentFrame := e.getFrameByIndex(obj.Parent)
					if parentFrame != nil {
						fieldVal, found = parentFrame.Get(part)
					}
				}

				if !found {
					return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{part, "", ""})
				}
			}

			base = fieldVal
		}
	}

	// Now handle final segment assignment
	finalPart := parts[len(parts)-1]

	// Check if base can be mutated
	if base.Type == value.TypeNone {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNonePath, [3]string{"cannot assign to none value", "", ""})
	}

	// Try as index first
	if index, parseErr := fmt.Sscanf(finalPart, "%d", new(int64)); parseErr == nil && index == 1 {
		var idx int64
		fmt.Sscanf(finalPart, "%d", &idx)

		if base.Type == value.TypeBlock {
			block, _ := base.AsBlock()
			if idx < 1 || idx > int64(len(block.Elements)) {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range", idx), "", ""})
			}
			block.Elements[idx-1] = result
		} else {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index assignment requires block type", "", ""})
		}
	} else {
		// It's a word (object field)
		if base.Type != value.TypeObject {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{"cannot assign field to non-object", "", ""})
		}

		obj, _ := base.AsObject()
		objFrame := e.getFrameByIndex(obj.FrameIndex)
		if objFrame == nil {
			return value.NoneVal(), verror.NewInternalError("object frame not found", [3]string{})
		}

		// Check if field exists
		_, found := objFrame.Get(finalPart)
		if !found {
			// Check parent chain to see if it's inherited
			if obj.Parent >= 0 {
				parentFrame := e.getFrameByIndex(obj.Parent)
				if parentFrame != nil {
					_, found = parentFrame.Get(finalPart)
				}
			}

			if !found {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{finalPart, "", ""})
			}
		}

		// Bind to object's frame
		objFrame.Bind(finalPart, result)
	}

	return result, nil
}
