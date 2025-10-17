package eval

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/stack"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Evaluator represents the core evaluation engine for the Vi programming language.
// It manages execution context including frames, call stack, and value evaluation.
type Evaluator struct {
	Stack      *stack.Stack
	Frames     []core.Frame
	frameStore []core.Frame
	captured   map[int]bool
	callStack  []string
}

// NewEvaluator creates a new evaluator instance with an initialized global frame.
// The global frame serves as the top-level context for variable bindings and function definitions.
func NewEvaluator() *Evaluator {
	global := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)
	global.Name = "(top level)"
	global.Index = 0
	e := &Evaluator{
		Stack:      stack.NewStack(1024),
		Frames:     []core.Frame{global},
		frameStore: []core.Frame{global},
		captured:   make(map[int]bool),
		callStack:  []string{"(top level)"},
	}
	e.captured[0] = true

	frame.InitTypeFrames()

	return e
}

// Callstack returns the current call stack as a slice of function names.
// The call stack represents the execution context, with the most recent call at the end.
func (e *Evaluator) Callstack() []string {
	return e.callStack
}

// currentFrame returns the currently active frame in the evaluation context.
// Returns nil if no frames are available.
func (e *Evaluator) currentFrame() core.Frame {
	if len(e.Frames) == 0 {
		return nil
	}
	return e.Frames[len(e.Frames)-1]
}

// pushFrame adds a new frame to the evaluation context and returns its index.
// If the frame doesn't have an index assigned, it will be stored in the frame store.
func (e *Evaluator) pushFrame(f core.Frame) int {
	idx := f.GetIndex()
	if idx < 0 {
		idx = len(e.frameStore)
		e.frameStore = append(e.frameStore, f)
		f.SetIndex(idx)
	}
	e.Frames = append(e.Frames, f)
	return idx
}

// currentFrameIndex returns the index of the currently active frame.
// Returns -1 if no frames are available.
func (e *Evaluator) currentFrameIndex() int {
	if len(e.Frames) == 0 {
		return -1
	}
	current := e.Frames[len(e.Frames)-1]
	return current.GetIndex()
}

// evalFunc represents a function type for evaluating different value types.
// It takes an evaluator and a value, returning the evaluated result and any error.
type evalFunc func(core.Evaluator, core.Value) (core.Value, error)

// evalDispatch maps value types to their corresponding evaluation functions.
// This dispatch table enables polymorphic evaluation of different Vi value types.
var evalDispatch map[core.ValueType]evalFunc

// init initializes the evaluation dispatch table with handlers for all supported value types.
func init() {
	evalDispatch = map[core.ValueType]evalFunc{
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

// evalLiteral handles evaluation of literal values (integers, strings, logic, none, decimal, object, port, datatype).
// Literal values evaluate to themselves without any computation.
func evalLiteral(e core.Evaluator, val core.Value) (core.Value, error) {
	return val, nil
}

// evalBlock handles evaluation of block values.
// Block values evaluate to themselves without execution of their contents.
func evalBlock(e core.Evaluator, val core.Value) (core.Value, error) {
	return val, nil
}

// evalFunction handles evaluation of function values.
// Function values evaluate to themselves without invocation.
func evalFunction(e core.Evaluator, val core.Value) (core.Value, error) {
	return val, nil
}

// evalParenDispatch handles evaluation of parenthesized expressions.
// Parentheses contain a block that gets executed immediately.
func evalParenDispatch(e core.Evaluator, val core.Value) (core.Value, error) {
	block, ok := value.AsBlock(val)
	if !ok {
		return value.NoneVal(), verror.NewInternalError("paren value does not contain BlockValue", [3]string{})
	}

	return e.DoBlock(block.Elements)
}

// evalWordDispatch handles evaluation of word values with debugging support.
// Words are looked up in the current context and may trigger breakpoints.
func evalWordDispatch(e core.Evaluator, val core.Value) (core.Value, error) {
	if debug.GlobalDebugger != nil {
		wordStr, ok := value.AsWord(val)
		if ok && debug.GlobalDebugger.HasBreakpoint(wordStr) {
			if trace.GlobalTraceSession != nil && trace.GlobalTraceSession.IsEnabled() {
				trace.GlobalTraceSession.Emit(trace.TraceEvent{
					Timestamp: time.Now(),
					Word:      "debug",
					Value:     fmt.Sprintf("breakpoint hit: %s", wordStr),
					Duration:  0,
					Depth:     len(e.Callstack()),
				})
			}
		}
	}
	return evalWord(e, val)
}

// evalSetWordDispatch handles evaluation of set-word values without assignment.
// Set-words without values are invalid and result in an error.
func evalSetWordDispatch(e core.Evaluator, val core.Value) (core.Value, error) {
	wordStr, _ := value.AsWord(val)
	return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})
}

// evalGetWordDispatch handles evaluation of get-word values.
// Get-words return the value bound to the word without evaluation.
func evalGetWordDispatch(e core.Evaluator, val core.Value) (core.Value, error) {
	return evalGetWord(e, val)
}

// evalLitWordDispatch handles evaluation of lit-word values.
// Lit-words evaluate to word values containing the same string.
func evalLitWordDispatch(e core.Evaluator, val core.Value) (core.Value, error) {
	return value.WordVal(val.GetPayload().(string)), nil
}

// evalPathDispatch handles evaluation of path values.
// Paths navigate through object fields and collection indices.
func evalPathDispatch(e core.Evaluator, val core.Value) (core.Value, error) {
	return evalPath(e, val)
}

// popFrame removes the current frame from the evaluation context and returns its index.
// If the frame is not captured, it may be cleared from the frame store.
// Captured frames are converted to closure type for potential reuse.
func (e *Evaluator) popFrame() int {
	if len(e.Frames) == 0 {
		return -1
	}
	frm := e.Frames[len(e.Frames)-1]
	e.Frames = e.Frames[:len(e.Frames)-1]
	idx := frm.GetIndex()
	if !e.captured[idx] {
		e.frameStore[idx] = nil
	} else if frm.GetType() != frame.FrameClosure {
		frm.ChangeType(frame.FrameClosure)
	}
	return idx
}

// pushCall adds a function name to the call stack for debugging and error reporting.
// Anonymous functions are labeled as "(anonymous)".
func (e *Evaluator) pushCall(name string) {
	if name == "" {
		name = "(anonymous)"
	}
	e.callStack = append(e.callStack, name)
}

// popCall removes the most recent function name from the call stack.
// Does nothing if the call stack has only one entry (top level).
func (e *Evaluator) popCall() {
	if len(e.callStack) <= 1 {
		return
	}
	e.callStack = e.callStack[:len(e.callStack)-1]
}

// captureCallStack returns a copy of the current call stack with proper ordering.
// The returned slice has the top-level call first and the most recent call last.
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

// annotateError enhances error information with context from the current evaluation state.
// Adds "near" information and call stack "where" information to script errors.
func (e *Evaluator) annotateError(err error, vals []core.Value, idx int) error {
	if err == nil {
		return nil
	}
	if verr, ok := err.(*verror.Error); ok {
		if idx >= 0 && idx < len(vals) && verr.Near == "" {
			verr.SetNear(verror.CaptureNear(vals, idx))
		}
		if len(verr.Where) == 0 {
			where := e.captureCallStack()
			if len(where) > 0 {
				verr.SetWhere(where)
			}
		}
	}
	return err
}

// functionDisplayName returns a display name for a function value.
// Returns "(anonymous)" for functions without names, otherwise returns the function name.
func functionDisplayName(fn *value.FunctionValue) string {
	if fn == nil || fn.Name == "" {
		return "(anonymous)"
	}
	return fn.Name
}

// MarkFrameCaptured marks a frame as captured to prevent it from being cleared.
// Captured frames are retained for closure access and potential reuse.
func (e *Evaluator) MarkFrameCaptured(idx int) {
	if idx >= 0 {
		e.captured[idx] = true
	}
}

// CurrentFrameIndex returns the index of the currently active frame.
// This is a public accessor for the currentFrameIndex method.
func (e *Evaluator) CurrentFrameIndex() int {
	return e.currentFrameIndex()
}

// RegisterFrame registers a frame in the frame store and returns its index.
// If the frame already has an index, returns it; otherwise assigns a new one.
func (e *Evaluator) RegisterFrame(f core.Frame) int {
	if f.GetIndex() >= 0 {
		return f.GetIndex()
	}

	idx := len(e.frameStore)
	e.frameStore = append(e.frameStore, f)
	f.SetIndex(idx)
	return idx
}

// GetFrameByIndex retrieves a frame from the frame store by its index.
// Returns nil if the index is out of bounds.
func (e *Evaluator) GetFrameByIndex(idx int) core.Frame {
	if idx < 0 || idx >= len(e.frameStore) {
		return nil
	}
	return e.frameStore[idx]
}

// PushFrameContext adds a frame to the evaluation context and returns its index.
// This is similar to pushFrame but may be used for different context management.
func (e *Evaluator) PushFrameContext(f core.Frame) int {
	idx := f.GetIndex()
	if idx < 0 {
		idx = len(e.frameStore)
		e.frameStore = append(e.frameStore, f)
		f.SetIndex(idx)
	}
	e.Frames = append(e.Frames, f)
	return idx
}

// PopFrameContext removes the current frame from the evaluation context.
// This is a public accessor for the popFrame method.
func (e *Evaluator) PopFrameContext() {
	e.popFrame()
}

// Lookup searches for a symbol in the current frame hierarchy.
// Starts with the current frame and walks up the parent chain until found or top level reached.
// Returns the bound value and true if found, otherwise returns none value and false.
func (e *Evaluator) Lookup(symbol string) (core.Value, bool) {
	frame := e.currentFrame()
	for frame != nil {
		if val, ok := frame.Get(symbol); ok {
			return val, true
		}
		if frame.GetParent() == -1 {
			break
		}
		frame = e.GetFrameByIndex(frame.GetParent())
	}
	return value.NoneVal(), false
}

// DoNext evaluates a single value using the appropriate evaluation function based on its type.
// Supports tracing for performance monitoring and debugging.
// Returns the evaluated result and any error encountered during evaluation.
func (e *Evaluator) DoNext(val core.Value) (core.Value, error) {
	var traceStart time.Time
	var traceWord string
	if trace.GlobalTraceSession != nil && trace.GlobalTraceSession.IsEnabled() {
		traceStart = time.Now()
		if value.IsWord(val.GetType()) {
			if w, ok := value.AsWord(val); ok {
				traceWord = w
			}
		}
	}

	evalFn, found := evalDispatch[val.GetType()]
	if !found {
		result := value.NoneVal()
		err := verror.NewInternalError("unknown value type in Do_Next", [3]string{})
		return result, err
	}

	result, err := evalFn(e, val)

	if trace.GlobalTraceSession != nil && trace.GlobalTraceSession.IsEnabled() && traceWord != "" {
		duration := time.Since(traceStart)
		trace.GlobalTraceSession.Emit(trace.TraceEvent{
			Timestamp: traceStart,
			Value:     result.String(),
			Word:      traceWord,
			Duration:  duration.Nanoseconds(),
			Depth:     len(e.callStack),
		})
	}

	return result, err
}

// DoBlock evaluates a sequence of values as a block.
// Handles set-word assignments and regular evaluations, returning the result of the last expression.
// Empty blocks return none value.
func (e *Evaluator) DoBlock(vals []core.Value) (core.Value, error) {
	if len(vals) == 0 {
		return value.NoneVal(), nil
	}

	var lastResult core.Value = value.NoneVal()
	var err error

	for i := 0; i < len(vals); i++ {
		val := vals[i]

		// Handle set-word assignments (x: value)
		if val.GetType() == value.TypeSetWord {
			lastResult, err = e.evalSetWord(val, vals, &i)
			if err != nil {
				return value.NoneVal(), e.annotateError(err, vals, i)
			}
			continue
		}

		// Evaluate regular expressions, potentially with function calls
		startIdx := i
		lastResult, err = e.evaluateWithFunctionCall(val, vals, &i, lastResult)
		if err != nil {
			return value.NoneVal(), e.annotateError(err, vals, startIdx)
		}
	}

	return lastResult, nil
}

// evalExpressionFromTokens evaluates a single expression from a token sequence starting at the given position.
// Handles both set-word assignments and regular evaluations.
// Returns the result, new position, and any error encountered.
func (e *Evaluator) evalExpressionFromTokens(tokens []core.Value, pos int, lastResult core.Value) (core.Value, int, error) {
	if pos >= len(tokens) {
		return value.NoneVal(), pos, verror.NewScriptError(verror.ErrIDNoValue, [3]string{"missing expression", "", ""})
	}

	idx := pos
	current := tokens[idx]

	if current.GetType() == value.TypeSetWord {
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

// invokeFunction invokes a function with the given arguments and context.
// Handles both native and user-defined functions, managing the call stack and argument collection.
// Supports infix function calls where the last result is used as the first argument.
func (e *Evaluator) invokeFunction(fn *value.FunctionValue, vals []core.Value, idx *int, lastResult core.Value) (core.Value, error) {
	startIdx := *idx
	name := functionDisplayName(fn)

	// Track function call for debugging
	e.pushCall(name)
	defer e.popCall()

	// Parse arguments including refinements and infix handling
	tokens := vals[*idx+1:]
	posArgs, refValues, consumed, err := e.collectFunctionArgsWithInfix(fn, tokens, lastResult)
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, startIdx)
	}

	*idx += consumed

	// Dispatch to native or user-defined function
	if fn.Type == value.FuncNative {
		result, callErr := e.callNative(fn, posArgs, refValues)
		if callErr != nil {
			return value.NoneVal(), e.annotateError(callErr, vals, startIdx)
		}
		return result, nil
	}

	result, execErr := e.executeFunction(fn, posArgs, refValues)
	if execErr != nil {
		return value.NoneVal(), execErr
	}
	return result, nil
}

// callNative invokes a native function with positional and refinement arguments.
// Native functions are implemented in Go and called directly.
// Returns the result of the native function call or an error.
func (e *Evaluator) callNative(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value) (core.Value, error) {
	if fn.Type != value.FuncNative {
		return value.NoneVal(), verror.NewInternalError("callNative expects native function", [3]string{})
	}

	result, err := fn.Native(posArgs, refValues, e)

	if err == nil {
		return result, nil
	}
	if verr, ok := err.(*verror.Error); ok {
		return result, verr
	}
	return value.NoneVal(), verror.NewInternalError(err.Error(), [3]string{})
}

// evaluateWithFunctionCall evaluates a value, potentially invoking a function if the value is a word.
// If the word resolves to a function, it invokes the function with appropriate arguments.
// Otherwise, evaluates the value normally.
func (e *Evaluator) evaluateWithFunctionCall(val core.Value, seq []core.Value, idx *int, lastResult core.Value) (core.Value, error) {
	if val.GetType() != value.TypeWord {
		return e.DoNext(val)
	}

	wordStr, ok := value.AsWord(val)
	if !ok {
		return e.DoNext(val)
	}

	resolved, found := e.Lookup(wordStr)
	if !found {
		return e.DoNext(val)
	}

	if resolved.GetType() == value.TypeFunction {
		fn, _ := value.AsFunction(resolved)
		return e.invokeFunction(fn, seq, idx, lastResult)
	}

	return e.DoNext(val)
}

// isRefinement checks if a value represents a function refinement.
// Refinements are words that start with "--" and modify function behavior.
func isRefinement(val core.Value) bool {
	if val.GetType() != value.TypeWord {
		return false
	}
	wordStr, ok := value.AsWord(val)
	if !ok {
		return false
	}
	return strings.HasPrefix(wordStr, "--")
}

// readRefinements parses refinement arguments from a token sequence.
// Refinements can be flags (boolean) or take values.
// Updates the refinement values map and tracks which refinements have been provided.
func (e *Evaluator) readRefinements(tokens []core.Value, pos int, refSpecs map[string]value.ParamSpec, refValues map[string]core.Value, refProvided map[string]bool) (int, error) {
	for pos < len(tokens) && isRefinement(tokens[pos]) {
		wordStr, _ := value.AsWord(tokens[pos])
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

		// Handle refinements that take values vs. boolean flags
		if spec.TakesValue {
			if pos+1 >= len(tokens) {
				return pos, verror.NewScriptError(
					verror.ErrIDInvalidOperation,
					[3]string{fmt.Sprintf("Refinement --%s requires a value", refName), "", ""},
				)
			}
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

// collectFunctionArgsWithInfix collects positional and refinement arguments for function invocation.
// Handles infix functions where the last result becomes the first argument.
// Parses refinements and validates argument counts.
func (e *Evaluator) collectFunctionArgsWithInfix(fn *value.FunctionValue, tokens []core.Value, lastResult core.Value) ([]core.Value, map[string]core.Value, int, error) {
	displayName := functionDisplayName(fn)

	// Separate positional and refinement parameters
	positional := make([]value.ParamSpec, 0, len(fn.Params))
	refSpecs := make(map[string]value.ParamSpec)
	refValues := make(map[string]core.Value)
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

	posArgs := make([]core.Value, len(positional))
	pos := 0
	paramIndex := 0

	// Handle infix functions (last result becomes first argument)
	useInfix := fn.Infix && lastResult.GetType() != value.TypeNone
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

	// Parse each positional argument, allowing refinements to interleave
	for paramIndex < len(positional) {
		paramSpec := positional[paramIndex]

		newPos, err := e.readRefinements(tokens, pos, refSpecs, refValues, refProvided)
		if err != nil {
			return nil, nil, 0, err
		}
		pos = newPos

		if pos >= len(tokens) {
			return nil, nil, 0, verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{displayName, strconv.Itoa(len(positional)), strconv.Itoa(paramIndex)},
			)
		}

		// Evaluate or use literal depending on parameter spec
		var arg core.Value
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

	// Handle any remaining refinements after positional args
	newPos, err := e.readRefinements(tokens, pos, refSpecs, refValues, refProvided)
	if err != nil {
		return nil, nil, 0, err
	}
	pos = newPos

	return posArgs, refValues, pos, nil
}

// executeFunction executes a user-defined function with the given arguments and refinements.
// Creates a new frame for the function's local scope, binds parameters, and evaluates the function body.
// Returns the result of the last expression in the function body.
func (e *Evaluator) executeFunction(fn *value.FunctionValue, posArgs []core.Value, refinements map[string]core.Value) (core.Value, error) {
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

	result, err := e.DoBlock(fn.Body.Elements)
	if err != nil {
		return value.NoneVal(), err
	}

	return result, nil
}

// evalWord evaluates a word by looking it up in the current context.
// Words represent variable names and function names in Vi.
// Returns an error if the word is not found in any accessible frame.
func evalWord(e core.Evaluator, val core.Value) (core.Value, error) {
	wordStr, ok := value.AsWord(val)
	if !ok {
		return value.NoneVal(), verror.NewInternalError("word value does not contain string", [3]string{})
	}

	result, ok := e.Lookup(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	return result, nil
}

// evalSetWord handles assignment of values to words (set-word evaluation).
// Supports both simple word assignment and path-based assignment (e.g., obj.field = value).
// Creates a new frame if none exists and binds the value to the word.
func (e *Evaluator) evalSetWord(val core.Value, vals []core.Value, i *int) (core.Value, error) {
	wordStr, ok := value.AsWord(val)
	if !ok {
		return value.NoneVal(), verror.NewInternalError("set-word value does not contain string", [3]string{})
	}

	if *i+1 >= len(vals) {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})
	}

	// Handle path-based assignment (e.g., obj.field:)
	if strings.Contains(wordStr, ".") {
		return e.evalSetPath(wordStr, vals, i)
	}

	// Ensure we have a frame to bind to
	currentFrame := e.currentFrame()
	if currentFrame == nil {
		currentFrame = frame.NewFrame(frame.FrameFunctionArgs, -1)
		e.pushFrame(currentFrame)
	}

	*i++
	nextVal := vals[*i]

	// Evaluate the value to assign
	result, err := e.evaluateWithFunctionCall(nextVal, vals, i, value.NoneVal())
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, *i)
	}

	// Auto-name anonymous functions
	if result.GetType() == value.TypeFunction {
		if fnVal, ok := value.AsFunction(result); ok && fnVal.Name == "" {
			fnVal.Name = wordStr
		}
	}

	currentFrame.Bind(wordStr, result)

	return result, nil
}

// evalGetWord evaluates a get-word by looking up its value without evaluation.
// Get-words return the raw bound value, useful for accessing functions and literals.
func evalGetWord(e core.Evaluator, val core.Value) (core.Value, error) {
	wordStr, ok := value.AsWord(val)
	if !ok {
		return value.NoneVal(), verror.NewInternalError("get-word value does not contain string", [3]string{})
	}

	result, ok := e.Lookup(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	return result, nil
}

// pathTraversal represents the result of traversing a path expression.
// Contains the path segments and the values encountered during traversal.
type pathTraversal struct {
	segments []value.PathSegment
	values   []core.Value
}

// traversePath navigates through a path expression to access nested values.
// Supports object field access and collection indexing.
// The stopBeforeLast parameter controls whether to stop before the final segment.
func traversePath(e core.Evaluator, path *value.PathExpression, stopBeforeLast bool) (*pathTraversal, error) {
	if len(path.Segments) == 0 {
		return nil, verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"empty path", "", ""})
	}

	tr := &pathTraversal{
		segments: path.Segments,
		values:   make([]core.Value, 0, len(path.Segments)),
	}

	// Resolve the base value (first segment)
	firstSeg := path.Segments[0]
	var base core.Value

	switch firstSeg.Type {
	case value.PathSegmentWord:
		wordStr, ok := firstSeg.Value.(string)
		if !ok {
			return nil, verror.NewInternalError("word segment does not contain string", [3]string{})
		}

		base, ok = e.Lookup(wordStr)
		if !ok {
			return nil, verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
		}
	case value.PathSegmentIndex:
		// Numeric literals as base (e.g., 1.2.3)
		num, ok := firstSeg.Value.(int64)
		if !ok {
			return nil, verror.NewInternalError("index segment does not contain int64", [3]string{})
		}
		base = value.IntVal(num)
	default:
		return nil, verror.NewInternalError("unexpected first segment type", [3]string{fmt.Sprintf("%v", firstSeg.Type), "", ""})
	}

	tr.values = append(tr.values, base)

	// Traverse remaining segments
	endIdx := len(path.Segments)
	if stopBeforeLast && len(path.Segments) > 1 {
		endIdx = len(path.Segments) - 1
	}

	for i := 1; i < endIdx; i++ {
		seg := path.Segments[i]
		current := tr.values[len(tr.values)-1]

		if current.GetType() == value.TypeNone {
			return nil, verror.NewScriptError(verror.ErrIDNonePath, [3]string{})
		}

		switch seg.Type {
		case value.PathSegmentWord:
			// Object field access
			if current.GetType() != value.TypeObject {
				return nil, verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{value.TypeToString(current.GetType()), "", ""})
			}

			obj, _ := value.AsObject(current)
			fieldName, ok := seg.Value.(string)
			if !ok {
				return nil, verror.NewInternalError("word segment does not contain string", [3]string{})
			}

			objFrame := e.GetFrameByIndex(obj.FrameIndex)
			if objFrame == nil {
				return nil, verror.NewInternalError("object frame not found", [3]string{})
			}

			// Search field in object and prototype chain
			fieldVal, found := objFrame.Get(fieldName)
			if !found {
				currentProto := obj.ParentProto
				for currentProto != nil && !found {
					protoFrame := e.GetFrameByIndex(currentProto.FrameIndex)
					if protoFrame == nil {
						break
					}
					fieldVal, found = protoFrame.Get(fieldName)
					if found {
						break
					}

					currentProto = currentProto.ParentProto
				}

				if !found {
					return nil, verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{fieldName, "", ""})
				}
			}

			tr.values = append(tr.values, fieldVal)

		case value.PathSegmentIndex:
			// Collection indexing
			index, ok := seg.Value.(int64)
			if !ok {
				return nil, verror.NewInternalError("index segment does not contain int64", [3]string{})
			}

			if current.GetType() == value.TypeBlock {
				block, _ := value.AsBlock(current)
				if index < 1 || index > int64(len(block.Elements)) {
					return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range for block of length %d", index, len(block.Elements)), "", ""})
				}
				tr.values = append(tr.values, block.Elements[index-1])

			} else if current.GetType() == value.TypeString {
				str, _ := value.AsString(current)
				runes := []rune(str.String())
				if index < 1 || index > int64(len(runes)) {
					return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range for string of length %d", index, len(runes)), "", ""})
				}
				tr.values = append(tr.values, value.StrVal(string(runes[index-1])))

			} else if current.GetType() == value.TypeBinary {
				bin, _ := value.AsBinary(current)
				if index < 1 || index > int64(bin.Length()) {
					return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range for binary of length %d", index, bin.Length()), "", ""})
				}
				tr.values = append(tr.values, value.IntVal(int64(bin.At(int(index-1)))))

			} else {
				return nil, verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index requires block, string, or binary type", "", ""})
			}

		default:
			return nil, verror.NewInternalError("unsupported path segment type", [3]string{})
		}
	}

	return tr, nil
}

// parsePathString converts a dot-separated string path into a PathExpression.
// Supports both word segments (field names) and numeric segments (indices).
// Requires at least two segments for a valid path.
func parsePathString(pathStr string) (*value.PathExpression, error) {
	parts := strings.Split(pathStr, ".")
	if len(parts) < 2 {
		return nil, verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"set-path requires at least 2 segments", "", ""})
	}

	segments := make([]value.PathSegment, len(parts))
	for i, part := range parts {
		if idx, err := strconv.ParseInt(part, 10, 64); err == nil {
			segments[i] = value.PathSegment{
				Type:  value.PathSegmentIndex,
				Value: idx,
			}
		} else {
			segments[i] = value.PathSegment{
				Type:  value.PathSegmentWord,
				Value: part,
			}
		}
	}

	return value.NewPath(segments, value.NoneVal()), nil
}

// assignToPathTarget assigns a value to the target location identified by a path traversal.
// Supports assignment to object fields and block/string indices.
// Validates that the target is assignable and within bounds.
func (e *Evaluator) assignToPathTarget(tr *pathTraversal, newVal core.Value, pathStr string) (core.Value, error) {
	if len(tr.segments) < 2 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"set-path requires at least 2 segments", "", ""})
	}

	container := tr.values[len(tr.values)-1]
	finalSeg := tr.segments[len(tr.segments)-1]

	// Cannot assign to paths starting with numeric literals
	if tr.segments[0].Type == value.PathSegmentIndex {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{pathStr, "", ""})
	}

	if container.GetType() == value.TypeNone {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNonePath, [3]string{"cannot assign to none value", "", ""})
	}

	switch finalSeg.Type {
	case value.PathSegmentIndex:
		// Assign to collection element
		index, ok := finalSeg.Value.(int64)
		if !ok {
			return value.NoneVal(), verror.NewInternalError("index segment does not contain int64", [3]string{})
		}

		if container.GetType() != value.TypeBlock {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index assignment requires block type", "", ""})
		}

		block, _ := value.AsBlock(container)
		if index < 1 || index > int64(len(block.Elements)) {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{fmt.Sprintf("index %d out of range", index), "", ""})
		}
		block.Elements[index-1] = newVal

	case value.PathSegmentWord:
		// Assign to object field
		fieldName, ok := finalSeg.Value.(string)
		if !ok {
			return value.NoneVal(), verror.NewInternalError("word segment does not contain string", [3]string{})
		}

		if container.GetType() != value.TypeObject {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{"cannot assign field to non-object", "", ""})
		}

		obj, _ := value.AsObject(container)
		objFrame := e.GetFrameByIndex(obj.FrameIndex)
		if objFrame == nil {
			return value.NoneVal(), verror.NewInternalError("object frame not found", [3]string{})
		}

		// Check if field exists in object or prototype chain
		_, found := objFrame.Get(fieldName)
		if !found {
			if obj.ParentProto != nil {
				parentFrame := e.GetFrameByIndex(obj.ParentProto.FrameIndex)
				if parentFrame != nil {
					_, found = parentFrame.Get(fieldName)
				}
			}

			if !found {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{fieldName, "", ""})
			}
		}
		objFrame.Bind(fieldName, newVal)

	default:
		return value.NoneVal(), verror.NewInternalError("unsupported path segment type for assignment", [3]string{})
	}

	return newVal, nil
}

// evalPath evaluates a path expression to access nested values.
// Traverses the path to return the value at the final location.
// Supports object field access and collection indexing.
func evalPath(e core.Evaluator, val core.Value) (core.Value, error) {
	if val.GetType() != value.TypePath {
		typeString := value.TypeToString(val.GetType())
		return value.NoneVal(), verror.NewInternalError("evalPath called with non-path type", [3]string{typeString, "", ""})
	}

	path, ok := value.AsPath(val)
	if !ok {
		return value.NoneVal(), verror.NewInternalError("path value does not contain PathExpression - payload type mismatch", [3]string{fmt.Sprintf("payload=%T", val.GetPayload()), "", ""})
	}

	tr, err := traversePath(e, path, false)
	if err != nil {
		return value.NoneVal(), err
	}

	return tr.values[len(tr.values)-1], nil
}

// evalSetPath handles assignment to path expressions (set-path evaluation).
// Parses the path string, evaluates the value to assign, and performs the assignment.
// Supports assignment to nested object fields and collection elements.
func (e *Evaluator) evalSetPath(pathStr string, vals []core.Value, i *int) (core.Value, error) {
	path, err := parsePathString(pathStr)
	if err != nil {
		return value.NoneVal(), err
	}

	*i++
	if *i >= len(vals) {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-path without value", "", ""})
	}

	var result core.Value
	nextVal := vals[*i]

	result, err = e.evaluateWithFunctionCall(nextVal, vals, i, value.NoneVal())
	if err != nil {
		return value.NoneVal(), e.annotateError(err, vals, *i)
	}

	tr, err := traversePath(e, path, true)
	if err != nil {
		return value.NoneVal(), err
	}

	return e.assignToPathTarget(tr, result, pathStr)
}
