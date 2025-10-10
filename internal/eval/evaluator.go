// Package eval provides the core evaluation engine for Viro.
//
// Architecture (inspired by Lua implementation):
// - Evaluator struct holds the stack and execution state
// - Do_Next evaluates a single value (like eval.eval_expr in Lua)
// - Do_Blk evaluates a block of values in sequence (like eval_block in Lua)
// - Type-based dispatch routes evaluation by value type
//
// Function call handling (similar to Lua's handle_fn_call):
// - collectFunctionArgsWithInfix: reads arguments with refinements between them
// - readRefinements: reads refinements from current position (like read_refinements in Lua)
// - isRefinement: checks if token is a refinement (like is_refinement in Lua)
//
// Key improvements from Lua port:
// - Refinements can appear anywhere: fn arg1 --ref1 arg2 --ref2 arg3
// - Infix operators supported for both native and user functions
// - Simple loop through parameters (like Lua's while i <= #fn.args)
//
// Per Constitution Principle III: Explicit type dispatch, no polymorphism.
// Package eval implements the core evaluation engine for the Viro interpreter.
//
// The evaluator uses type-based dispatch to evaluate REBOL-style expressions.
// It supports literals, words, functions, blocks, and parens with left-to-right
// evaluation (no operator precedence) and proper scoping rules.
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
	captured   map[int]bool
	callStack  []string
}

// NewEvaluator creates a new evaluation engine with an empty stack.
func NewEvaluator() *Evaluator {
	global := frame.NewFrame(frame.FrameClosure, -1)
	global.Name = "(top level)"
	global.Index = 0
	e := &Evaluator{
		Stack:      stack.NewStack(1024),
		Frames:     []*frame.Frame{global},
		frameStore: []*frame.Frame{global},
		captured:   make(map[int]bool),
		callStack:  []string{"(top level)"},
	}
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

func (e *Evaluator) pushFrame(f *frame.Frame) int {
	idx := f.Index
	if idx < 0 {
		// Frame not yet in frameStore, add it
		idx = len(e.frameStore)
		e.frameStore = append(e.frameStore, f)
		f.Index = idx
	}
	e.Frames = append(e.Frames, f)
	return idx
}

func (e *Evaluator) currentFrameIndex() int {
	if len(e.Frames) == 0 {
		return -1
	}
	current := e.Frames[len(e.Frames)-1]
	return current.Index
}

func (e *Evaluator) getFrameByIndex(idx int) *frame.Frame {
	if idx < 0 || idx >= len(e.frameStore) {
		return nil
	}
	return e.frameStore[idx]
}

// evalFunc is a type-specific evaluation function
type evalFunc func(*Evaluator, value.Value) (value.Value, *verror.Error)

// evalDispatch maps value types to their evaluation functions.
// Initialized at package load time.
var evalDispatch map[value.ValueType]evalFunc

func init() {
	evalDispatch = map[value.ValueType]evalFunc{
		value.TypeInteger:  evalLiteral,
		value.TypeString:   evalLiteral,
		value.TypeLogic:    evalLiteral,
		value.TypeNone:     evalLiteral,
		value.TypeDecimal:  evalLiteral,
		value.TypeObject:   evalLiteral,
		value.TypePort:     evalLiteral,
		value.TypeDatatype: evalLiteral,
		value.TypeBlock:    evalBlock,
		value.TypeFunction: evalFunction,
		value.TypeParen:    evalParenDispatch,
		value.TypeWord:     evalWordDispatch,
		value.TypeSetWord:  evalSetWordDispatch,
		value.TypeGetWord:  evalGetWordDispatch,
		value.TypeLitWord:  evalLitWordDispatch,
		value.TypePath:     evalPathDispatch,
	}
}

// evalLiteral handles all literal types that evaluate to themselves
func evalLiteral(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return val, nil
}

// evalBlock handles block evaluation (deferred - returns self)
func evalBlock(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return val, nil
}

// evalFunction handles function evaluation (returns self)
func evalFunction(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return val, nil
}

// evalParenDispatch handles paren evaluation
func evalParenDispatch(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return e.evalParen(val)
}

// evalWordDispatch handles word evaluation
func evalWordDispatch(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return e.evalWord(val)
}

// evalSetWordDispatch handles set-word evaluation (error in isolation)
func evalSetWordDispatch(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	wordStr, _ := val.AsWord()
	return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})
}

// evalGetWordDispatch handles get-word evaluation
func evalGetWordDispatch(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return e.evalGetWord(val)
}

// evalLitWordDispatch handles lit-word evaluation
// Lit-word ('word) evaluates to the corresponding word (word)
func evalLitWordDispatch(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return value.WordVal(val.Payload.(string)), nil
}

// evalPathDispatch handles path evaluation
func evalPathDispatch(e *Evaluator, val value.Value) (value.Value, *verror.Error) {
	return e.evalPath(val)
}

// popFrame removes the active frame and returns its store index.
func (e *Evaluator) popFrame() int {
	if len(e.Frames) == 0 {
		return -1
	}
	frm := e.Frames[len(e.Frames)-1]
	e.Frames = e.Frames[:len(e.Frames)-1]
	idx := frm.Index
	if !e.captured[idx] {
		// Release non-captured frames for GC by clearing store entry
		e.frameStore[idx] = nil
		frm.Index = -1
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
	if f.Index >= 0 {
		return f.Index
	}

	// Add to store
	idx := len(e.frameStore)
	e.frameStore = append(e.frameStore, f)
	f.Index = idx
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
func (e *Evaluator) Lookup(symbol string) (value.Value, bool) {
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

	// Type-based dispatch using dispatch table
	evalFn, found := evalDispatch[val.Type]
	if !found {
		result := value.NoneVal()
		err := verror.NewInternalError("unknown value type in Do_Next", [3]string{})
		return result, err
	}

	result, err := evalFn(e, val)

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

	var lastResult value.Value = value.NoneVal()
	var err *verror.Error

	// Simple loop: evaluate each value, passing lastResult forward
	for i := 0; i < len(vals); i++ {
		val := vals[i]

		// Special case: set-word consumes next value
		if val.Type == value.TypeSetWord {
			lastResult, err = e.evalSetWord(val, vals, &i)
			if err != nil {
				return value.NoneVal(), e.annotateError(err, vals, i)
			}
			continue
		}

		startIdx := i
		lastResult, err = e.evaluateWithFunctionCall(val, vals, &i, lastResult)
		if err != nil {
			return value.NoneVal(), e.annotateError(err, vals, startIdx)
		}
	}

	return lastResult, nil
}

// evalExpressionFromTokens evaluates a single expression starting at the given
// position in the token slice and returns the resulting value along with the
// next position to continue reading from. This mirrors the Lua evaluator's
// eval_expr helper used when collecting function arguments and refinement
// values.
func (e *Evaluator) evalExpressionFromTokens(tokens []value.Value, pos int, lastResult value.Value) (value.Value, int, *verror.Error) {
	if pos >= len(tokens) {
		return value.NoneVal(), pos, verror.NewScriptError(verror.ErrIDNoValue, [3]string{"missing expression", "", ""})
	}

	idx := pos
	current := tokens[idx]

	if current.Type == value.TypeSetWord {
		result, err := e.evalSetWord(current, tokens, &idx)
		if err != nil {
			return value.NoneVal(), pos, err
		}
		return result, idx + 1, nil
	}

	result, err := e.evaluateWithFunctionCall(current, tokens, &idx, lastResult)
	if err != nil {
		return value.NoneVal(), pos, err
	}

	return result, idx + 1, nil
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

			if resolved, found := e.Lookup(wordStr); found && resolved.Type == value.TypeFunction {
				fn, _ := resolved.AsFunction()
				return e.invokeFunctionWithTokens(fn, block.Elements[1:])
			}
		}
	}

	return e.Do_Blk(block.Elements)
}

func (e *Evaluator) callNative(name string, info *native.NativeInfo, vals []value.Value, idx *int, lastResult value.Value) (value.Value, *verror.Error) {
	startIdx := *idx
	args := make([]value.Value, 0, info.Arity)

	// Infix handling: if info.Infix && lastResult is not none, use lastResult as first arg
	// This matches Lua: if fn.infix then table.insert(args, last_value)
	useInfix := info.Infix && lastResult.Type != value.TypeNone
	argIndex := 0 // tracks which parameter we're collecting
	tokensConsumed := 0

	if useInfix {
		args = append(args, lastResult)
		argIndex = 1
	}

	// Collect remaining arguments (simple loop like Lua)
	for argIndex < info.Arity {
		tokenIdx := *idx + 1 + tokensConsumed
		if tokenIdx >= len(vals) {
			err := verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{name, strconv.Itoa(info.Arity), strconv.Itoa(len(args))},
			)
			return value.NoneVal(), e.annotateError(err, vals, startIdx)
		}

		var arg value.Value
		var argErr *verror.Error
		// Check if this argument should be evaluated
		if info.EvalArgs != nil && argIndex < len(info.EvalArgs) && !info.EvalArgs[argIndex] {
			arg = vals[tokenIdx]
			tokensConsumed++
		} else {
			var nextPos int
			arg, nextPos, argErr = e.evalExpressionFromTokens(vals, tokenIdx, value.NoneVal())
			if argErr != nil {
				return value.NoneVal(), e.annotateError(argErr, vals, tokenIdx)
			}
			tokensConsumed += nextPos - tokenIdx
		}
		args = append(args, arg)
		argIndex++
	}

	// Advance index by number of tokens consumed from sequence
	*idx += tokensConsumed

	result, err := native.Call(info, args, e)
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, startIdx)
	}
	return result, nil
}

func (e *Evaluator) callNativeFromSlice(name string, info *native.NativeInfo, tokens []value.Value) (value.Value, *verror.Error) {
	context := append([]value.Value{value.WordVal(name)}, tokens...)

	args := make([]value.Value, 0, info.Arity)
	pos := 0
	for argIndex := 0; argIndex < info.Arity; argIndex++ {
		if pos >= len(tokens) {
			err := verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{name, strconv.Itoa(info.Arity), strconv.Itoa(len(args))},
			)
			return value.NoneVal(), e.annotateError(err, context, 0)
		}

		if info.EvalArgs != nil && argIndex < len(info.EvalArgs) && !info.EvalArgs[argIndex] {
			args = append(args, tokens[pos])
			pos++
			continue
		}

		arg, nextPos, err := e.evalExpressionFromTokens(tokens, pos, value.NoneVal())
		if err != nil {
			return value.NoneVal(), e.annotateError(err, context, pos+1)
		}
		args = append(args, arg)
		pos = nextPos
	}

	if pos != len(tokens) {
		err := verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{name, strconv.Itoa(info.Arity), strconv.Itoa(info.Arity + (len(tokens) - pos))},
		)
		return value.NoneVal(), e.annotateError(err, context, pos)
	}

	result, err := native.Call(info, args, e)
	if err != nil {
		return value.NoneVal(), e.annotateError(err, context, 0)
	}
	return result, nil
}

// evaluateWithFunctionCall resolves a value that might represent a callable.
//
// If the value is a word referring to a native or user-defined function, this
// helper dispatches to the appropriate call helper (advancing the index when
// arguments are consumed). Otherwise it falls back to Do_Next.
//
// lastResult is used for infix operators - if the function is infix, lastResult
// becomes the first argument.
func (e *Evaluator) evaluateWithFunctionCall(val value.Value, seq []value.Value, idx *int, lastResult value.Value) (value.Value, *verror.Error) {
	if val.Type != value.TypeWord {
		return e.Do_Next(val)
	}

	wordStr, ok := val.AsWord()
	if !ok {
		return e.Do_Next(val)
	}

	if nativeInfo, found := native.Lookup(wordStr); found {
		return e.callNative(wordStr, nativeInfo, seq, idx, lastResult)
	}

	if resolved, found := e.Lookup(wordStr); found && resolved.Type == value.TypeFunction {
		fn, _ := resolved.AsFunction()
		return e.invokeFunctionFromSequence(fn, seq, idx, lastResult)
	}

	return e.Do_Next(val)
}

func (e *Evaluator) invokeFunctionFromSequence(fn *value.FunctionValue, vals []value.Value, idx *int, lastResult value.Value) (value.Value, *verror.Error) {
	name := functionDisplayName(fn)
	e.pushCall(name)
	defer e.popCall()

	tokens := vals[*idx+1:]
	posArgs, refValues, consumed, err := e.collectFunctionArgsWithInfix(fn, tokens, lastResult)
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
			[3]string{name, strconv.Itoa(len(posArgs)), strconv.Itoa(len(posArgs) + (len(tokens) - consumed))},
		)
		return value.NoneVal(), e.annotateError(err, context, 0)
	}

	result, execErr := e.executeFunction(fn, posArgs, refValues)
	if execErr != nil {
		return value.NoneVal(), execErr
	}
	return result, nil
}

// isRefinement checks if a value is a refinement (word starting with "--")
// Corresponds to is_refinement in Lua implementation
func isRefinement(val value.Value) bool {
	if val.Type != value.TypeWord {
		return false
	}
	wordStr, ok := val.AsWord()
	if !ok {
		return false
	}
	return strings.HasPrefix(wordStr, "--")
}

// readRefinements reads refinements from current position in tokens.
// Corresponds to read_refinements in Lua implementation.
//
// Returns:
//   - newPos: position after consuming refinements
//   - error: if unknown refinement or missing value for value-taking refinement
//
// Modifies refValues map in-place with found refinements.
func (e *Evaluator) readRefinements(
	fn *value.FunctionValue,
	tokens []value.Value,
	pos int,
	refSpecs map[string]value.ParamSpec,
	refValues map[string]value.Value,
	refProvided map[string]bool,
) (int, *verror.Error) {
	for pos < len(tokens) && isRefinement(tokens[pos]) {
		wordStr, _ := tokens[pos].AsWord()
		refName := strings.TrimPrefix(wordStr, "--")

		spec, exists := refSpecs[refName]
		if !exists {
			return pos, verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{fmt.Sprintf("Unknown refinement: --%s", refName), "", ""},
			)
		}

		if refProvided[refName] {
			return pos, verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{fmt.Sprintf("Duplicate refinement: --%s", refName), "", ""},
			)
		}

		if spec.TakesValue {
			if pos+1 >= len(tokens) {
				return pos, verror.NewScriptError(
					verror.ErrIDInvalidOperation,
					[3]string{fmt.Sprintf("Refinement --%s requires a value", refName), "", ""},
				)
			}
			// Evaluate the value for the refinement using expression-aware evaluation
			arg, nextPos, err := e.evalExpressionFromTokens(tokens, pos+1, value.NoneVal())
			if err != nil {
				return pos, err
			}
			refValues[refName] = arg
			pos = nextPos
		} else {
			refValues[refName] = value.LogicVal(true)
			pos++
		}

		refProvided[refName] = true
	}

	return pos, nil
}

// collectFunctionArgsWithInfix collects function arguments with optional infix support.
// Corresponds to handle_fn_call in Lua implementation.
//
// If fn.Infix is true AND lastResult is not none, lastResult becomes the first argument.
// This matches the Lua behavior: if fn.infix then table.insert(args, last_value)
func (e *Evaluator) collectFunctionArgsWithInfix(fn *value.FunctionValue, tokens []value.Value, lastResult value.Value) ([]value.Value, map[string]value.Value, int, *verror.Error) {
	displayName := functionDisplayName(fn)

	// Separate positional params from refinements
	positional := make([]value.ParamSpec, 0, len(fn.Params))
	refSpecs := make(map[string]value.ParamSpec)
	refValues := make(map[string]value.Value)
	refProvided := make(map[string]bool)

	for _, spec := range fn.Params {
		if spec.Refinement {
			refSpecs[spec.Name] = spec
			// Initialize refinement values
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
	pos := 0
	paramIndex := 0

	// Infix handling (like Lua: if fn.infix then add last_value as first arg)
	useInfix := fn.Infix && lastResult.Type != value.TypeNone
	if useInfix {
		if len(positional) == 0 {
			return nil, nil, 0, verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{displayName, "0", "1 (infix requires at least one parameter)"},
			)
		}
		posArgs[0] = lastResult
		paramIndex = 1
	}

	// Main loop: iterate through positional parameters (like Lua's while i <= #fn.args)
	for paramIndex < len(positional) {
		paramSpec := positional[paramIndex]

		// Read refinements BEFORE this argument (like Lua)
		newPos, err := e.readRefinements(fn, tokens, pos, refSpecs, refValues, refProvided)
		if err != nil {
			return nil, nil, 0, err
		}
		pos = newPos

		// Check if we've run out of tokens
		if pos >= len(tokens) {
			return nil, nil, 0, verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{displayName, strconv.Itoa(len(positional)), strconv.Itoa(paramIndex)},
			)
		}

		// Read the argument (eval or raw, based on paramSpec.Eval)
		var arg value.Value
		if paramSpec.Eval {
			var nextPos int
			arg, nextPos, err = e.evalExpressionFromTokens(tokens, pos, value.NoneVal())
			if err != nil {
				return nil, nil, 0, err
			}
			pos = nextPos
		} else {
			token := tokens[pos]
			arg = token
			pos++
		}

		posArgs[paramIndex] = arg
		paramIndex++
	}

	// Read remaining refinements AFTER all arguments (like Lua)
	newPos, err := e.readRefinements(fn, tokens, pos, refSpecs, refValues, refProvided)
	if err != nil {
		return nil, nil, 0, err
	}
	pos = newPos

	return posArgs, refValues, pos, nil
}

func (e *Evaluator) collectFunctionArgs(fn *value.FunctionValue, tokens []value.Value) ([]value.Value, map[string]value.Value, int, *verror.Error) {
	// Delegate to the infix-aware version with none as lastResult
	return e.collectFunctionArgsWithInfix(fn, tokens, value.NoneVal())
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

	result, ok := e.Lookup(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	// Return the value without re-evaluation
	// (Functions are called via evaluateWithFunctionCall, not here)
	return result, nil
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

	// Evaluate with lastResult = none (set-word doesn't use lastResult for infix)
	result, err := e.evaluateWithFunctionCall(nextVal, vals, i, value.NoneVal())
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

	result, ok := e.Lookup(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	// Return without re-evaluation (key difference from evalWord)
	return result, nil
}

// pathTraversal represents a path resolution result with intermediate values
type pathTraversal struct {
	segments []value.PathSegment // original segments from path expression
	values   []value.Value       // resolved values at each level (includes base)
}

// traversePath performs path segment resolution, optionally stopping before the last segment.
// This centralizes path logic for both read (full traversal) and write (stop before last).
//
// Parameters:
//   - path: the PathExpression to traverse
//   - stopBeforeLast: if true, stops after resolving penultimate segment (for writes)
//
// Returns:
//   - pathTraversal with segments and resolved values at each level
//   - error if traversal fails (no-value, none-path, type mismatch, out of range)
func (e *Evaluator) traversePath(path *value.PathExpression, stopBeforeLast bool) (*pathTraversal, *verror.Error) {
	if len(path.Segments) == 0 {
		return nil, verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"empty path", "", ""})
	}

	tr := &pathTraversal{
		segments: path.Segments,
		values:   make([]value.Value, 0, len(path.Segments)),
	}

	// Resolve base (first segment)
	firstSeg := path.Segments[0]
	var base value.Value

	if firstSeg.Type == value.PathSegmentWord {
		wordStr, ok := firstSeg.Value.(string)
		if !ok {
			return nil, verror.NewInternalError("word segment does not contain string", [3]string{})
		}

		base, ok = e.Lookup(wordStr)
		if !ok {
			return nil, verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
		}
	} else if firstSeg.Type == value.PathSegmentIndex {
		// Path starts with a literal number (e.g., 1.field)
		num, ok := firstSeg.Value.(int64)
		if !ok {
			return nil, verror.NewInternalError("index segment does not contain int64", [3]string{})
		}
		base = value.IntVal(num)
	} else {
		return nil, verror.NewInternalError("unexpected first segment type", [3]string{fmt.Sprintf("%v", firstSeg.Type), "", ""})
	}

	tr.values = append(tr.values, base)

	// Determine traversal limit
	endIdx := len(path.Segments)
	if stopBeforeLast && len(path.Segments) > 1 {
		endIdx = len(path.Segments) - 1
	}

	// Traverse segments
	for i := 1; i < endIdx; i++ {
		seg := path.Segments[i]
		current := tr.values[len(tr.values)-1]

		// Check for none mid-path
		if current.Type == value.TypeNone {
			return nil, verror.NewScriptError(verror.ErrIDNonePath, [3]string{})
		}

		switch seg.Type {
		case value.PathSegmentWord:
			// Object field access
			if current.Type != value.TypeObject {
				return nil, verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{current.Type.String(), "", ""})
			}

			obj, _ := current.AsObject()
			fieldName, ok := seg.Value.(string)
			if !ok {
				return nil, verror.NewInternalError("word segment does not contain string", [3]string{})
			}

			// Look up field in object's frame
			objFrame := e.getFrameByIndex(obj.FrameIndex)
			if objFrame == nil {
				return nil, verror.NewInternalError("object frame not found", [3]string{})
			}

			fieldVal, found := objFrame.Get(fieldName)
			if !found {
				// Check parent prototype chain
				currentProto := obj.ParentProto
				for currentProto != nil && !found {
					protoFrame := e.getFrameByIndex(currentProto.FrameIndex)
					if protoFrame == nil {
						break
					}
					fieldVal, found = protoFrame.Get(fieldName)
					if found {
						break
					}

					// Move to next prototype in chain
					currentProto = currentProto.ParentProto
				}

				if !found {
					return nil, verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{fieldName, "", ""})
				}
			}

			tr.values = append(tr.values, fieldVal)

		case value.PathSegmentIndex:
			// Block or string indexing (1-based)
			index, ok := seg.Value.(int64)
			if !ok {
				return nil, verror.NewInternalError("index segment does not contain int64", [3]string{})
			}

			if current.Type == value.TypeBlock {
				block, _ := current.AsBlock()
				// 1-based indexing
				if index < 1 || index > int64(len(block.Elements)) {
					return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range for block of length %d", index, len(block.Elements)), "", ""})
				}
				tr.values = append(tr.values, block.Elements[index-1])

			} else if current.Type == value.TypeString {
				str, _ := current.AsString()
				runes := []rune(str.String())
				if index < 1 || index > int64(len(runes)) {
					return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range for string of length %d", index, len(runes)), "", ""})
				}
				// Return character as string
				tr.values = append(tr.values, value.StrVal(string(runes[index-1])))

			} else {
				return nil, verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index requires block or string type", "", ""})
			}

		default:
			return nil, verror.NewInternalError("unsupported path segment type", [3]string{})
		}
	}

	return tr, nil
}

// parsePathString converts a dot-separated path string into a PathExpression.
// Used by evalSetPath to leverage the centralized traversePath logic.
func parsePathString(pathStr string) (*value.PathExpression, *verror.Error) {
	parts := strings.Split(pathStr, ".")
	if len(parts) < 2 {
		return nil, verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"set-path requires at least 2 segments", "", ""})
	}

	segments := make([]value.PathSegment, len(parts))
	for i, part := range parts {
		// Try to parse as integer index
		if idx, err := strconv.ParseInt(part, 10, 64); err == nil {
			segments[i] = value.PathSegment{
				Type:  value.PathSegmentIndex,
				Value: idx,
			}
		} else {
			// It's a word segment
			segments[i] = value.PathSegment{
				Type:  value.PathSegmentWord,
				Value: part,
			}
		}
	}

	return value.NewPath(segments, value.NoneVal()), nil
}

// assignToPathTarget performs the final assignment operation after path traversal.
// This handles assignment to object fields or block elements.
func (e *Evaluator) assignToPathTarget(tr *pathTraversal, newVal value.Value, pathStr string) (value.Value, *verror.Error) {
	if len(tr.segments) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"set-path requires at least 2 segments", "", ""})
	}

	// Get the penultimate (container) value
	container := tr.values[len(tr.values)-1]
	finalSeg := tr.segments[len(tr.segments)-1]

	// Check for immutable target (literal number as base)
	if tr.segments[0].Type == value.PathSegmentIndex {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{pathStr, "", ""})
	}

	// Check if container can be mutated
	if container.Type == value.TypeNone {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNonePath, [3]string{"cannot assign to none value", "", ""})
	}

	switch finalSeg.Type {
	case value.PathSegmentIndex:
		// Block element assignment
		index, ok := finalSeg.Value.(int64)
		if !ok {
			return value.NoneVal(), verror.NewInternalError("index segment does not contain int64", [3]string{})
		}

		if container.Type != value.TypeBlock {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index assignment requires block type", "", ""})
		}

		block, _ := container.AsBlock()
		if index < 1 || index > int64(len(block.Elements)) {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range", index), "", ""})
		}
		block.Elements[index-1] = newVal

	case value.PathSegmentWord:
		// Object field assignment
		fieldName, ok := finalSeg.Value.(string)
		if !ok {
			return value.NoneVal(), verror.NewInternalError("word segment does not contain string", [3]string{})
		}

		if container.Type != value.TypeObject {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{"cannot assign field to non-object", "", ""})
		}

		obj, _ := container.AsObject()
		objFrame := e.getFrameByIndex(obj.FrameIndex)
		if objFrame == nil {
			return value.NoneVal(), verror.NewInternalError("object frame not found", [3]string{})
		}

		// Check if field exists
		_, found := objFrame.Get(fieldName)
		if !found {
			// Check parent chain to see if it's inherited
			if obj.Parent >= 0 {
				parentFrame := e.getFrameByIndex(obj.Parent)
				if parentFrame != nil {
					_, found = parentFrame.Get(fieldName)
				}
			}

			if !found {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{fieldName, "", ""})
			}
		}

		// Bind to object's frame
		objFrame.Bind(fieldName, newVal)

	default:
		return value.NoneVal(), verror.NewInternalError("unsupported path segment type for assignment", [3]string{})
	}

	return newVal, nil
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

	// Use centralized traversal logic for full path resolution
	tr, err := e.traversePath(path, false)
	if err != nil {
		return value.NoneVal(), err
	}

	// Return the final resolved value
	return tr.values[len(tr.values)-1], nil
}

// evalSetPath handles path assignment (set-path) like obj.field: value (T091).
//
// Contract per contracts/objects.md §3:
// - Parse path string to extract segments
// - Traverse to penultimate segment
// - Update final segment in target container (object frame or block)
// - Error if attempting to assign to literal or immutable target
func (e *Evaluator) evalSetPath(pathStr string, vals []value.Value, i *int) (value.Value, *verror.Error) {
	// Parse path string into PathExpression
	path, err := parsePathString(pathStr)
	if err != nil {
		return value.NoneVal(), err
	}

	// Evaluate the value to assign
	*i++
	if *i >= len(vals) {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-path without value", "", ""})
	}

	var result value.Value
	nextVal := vals[*i]

	result, err = e.evaluateWithFunctionCall(nextVal, vals, i, value.NoneVal())
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, *i)
	}

	// Use centralized traversal to reach the penultimate container
	tr, err := e.traversePath(path, true)
	if err != nil {
		return value.NoneVal(), err
	}

	// Perform the final assignment
	return e.assignToPathTarget(tr, result, pathStr)
}
