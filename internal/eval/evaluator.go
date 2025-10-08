// Package eval provides the core evaluation engine for Viro.
//
// Architecture:
// - Evaluator struct holds the stack and execution state
// - Do_Next evaluates a single value
// - Do_Blk evaluates a block of values in sequence
// - Type-based dispatch routes evaluation by value type
//
// Per Constitution Principle III: Explicit type dispatch, no polymorphism.
package eval

import (
	"fmt"
	"strings"

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
	// Type-based dispatch per Constitution Principle III
	switch val.Type {
	case value.TypeInteger, value.TypeString, value.TypeLogic, value.TypeNone:
		// Literals evaluate to themselves
		return val, nil

	case value.TypeBlock:
		// Blocks return self (deferred evaluation)
		return val, nil

	case value.TypeParen:
		// Parens evaluate contents
		return e.evalParen(val)

	case value.TypeWord:
		// Words look up in frame
		return e.evalWord(val)

	case value.TypeSetWord:
		// Set-words are handled in Do_Blk (need next value from sequence)
		// If we reach here, it's a set-word in isolation (error)
		wordStr, _ := val.AsWord()
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})

	case value.TypeGetWord:
		// Get-words fetch without evaluation
		return e.evalGetWord(val)

	case value.TypeLitWord:
		// Lit-words return as word
		// Extract the word symbol and return as a word
		return value.WordVal(val.Payload.(string)), nil

	case value.TypeFunction:
		// Functions return self (they're values)
		return val, nil

	default:
		return value.NoneVal(), verror.NewInternalError("unknown value type in Do_Next", [3]string{})
	}
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
