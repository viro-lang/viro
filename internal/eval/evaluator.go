package eval

import (
	"fmt"
	"io"
	"os"
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
	Stack        *stack.Stack
	Frames       []core.Frame
	frameStore   []core.Frame
	captured     map[int]bool
	callStack    []string
	OutputWriter io.Writer
	ErrorWriter  io.Writer
	InputReader  io.Reader

	// Cached trace state fields for performance optimization.
	// These fields are synchronized with the global trace session and must be updated via UpdateTraceCache().
	// Call UpdateTraceCache() after any change to the global trace session (e.g., enabling/disabling tracing,
	// or modifying trace filters) to ensure cache consistency.
	traceEnabled         bool
	traceShouldTraceExpr bool
}

// NewEvaluator creates a new evaluator instance with an initialized global frame.
// The global frame serves as the top-level context for variable bindings and function definitions.
func NewEvaluator() *Evaluator {
	global := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)
	global.Name = "(top level)"
	global.Index = 0
	e := &Evaluator{
		Stack:        stack.NewStack(1024),
		Frames:       []core.Frame{global},
		frameStore:   []core.Frame{global},
		captured:     make(map[int]bool),
		callStack:    []string{"(top level)"},
		OutputWriter: os.Stdout,
		ErrorWriter:  os.Stderr,
		InputReader:  os.Stdin,
	}
	e.captured[0] = true

	frame.InitTypeFrames()

	e.UpdateTraceCache()

	return e
}

// SetOutputWriter sets the output writer for this evaluator.
// This allows redirecting output (e.g., from print function) to different destinations.
func (e *Evaluator) SetOutputWriter(w io.Writer) {
	if w == nil {
		e.OutputWriter = os.Stdout
	} else {
		e.OutputWriter = w
	}
}

// GetOutputWriter returns the current output writer.
func (e *Evaluator) GetOutputWriter() io.Writer {
	return e.OutputWriter
}

// SetErrorWriter sets the error writer for this evaluator.
// This allows redirecting error output to different destinations.
func (e *Evaluator) SetErrorWriter(w io.Writer) {
	if w == nil {
		e.ErrorWriter = os.Stderr
	} else {
		e.ErrorWriter = w
	}
}

// GetErrorWriter returns the current error writer.
func (e *Evaluator) GetErrorWriter() io.Writer {
	return e.ErrorWriter
}

// SetInputReader sets the input reader for this evaluator.
// This allows redirecting input (e.g., from input function) from different sources.
func (e *Evaluator) SetInputReader(r io.Reader) {
	if r == nil {
		e.InputReader = os.Stdin
	} else {
		e.InputReader = r
	}
}

// GetInputReader returns the current input reader.
func (e *Evaluator) GetInputReader() io.Reader {
	return e.InputReader
}

// UpdateTraceCache refreshes the cached trace settings from the global trace session.
// This should be called whenever trace settings change (enable/disable/filter updates).
func (e *Evaluator) UpdateTraceCache() {
	if trace.GlobalTraceSession == nil {
		e.traceEnabled = false
		e.traceShouldTraceExpr = false
		return
	}
	e.traceEnabled = trace.GlobalTraceSession.IsEnabled()
	e.traceShouldTraceExpr = e.traceEnabled && trace.GlobalTraceSession.ShouldTraceExpression()
}

// currentFrame returns the currently active frame in the evaluation context.
// Returns nil if no frames are available.
func (e *Evaluator) currentFrame() core.Frame {
	if len(e.Frames) == 0 {
		return nil
	}
	return e.Frames[len(e.Frames)-1]
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
	return value.NewNoneVal(), false
}

// DoBlock evaluates a sequence of values as a block using position tracking.
// Returns the result of the last expression or none value for empty blocks.
func (e *Evaluator) DoBlock(vals []core.Value) (core.Value, error) {
	var traceStart time.Time
	if e.traceEnabled {
		traceStart = time.Now()
		e.emitTraceResult("block-enter", "", fmt.Sprintf("[%d expressions]", len(vals)), value.NewNoneVal(), 0, traceStart, nil)
	}

	if len(vals) == 0 {
		if e.traceEnabled {
			e.emitTraceResult("block-exit", "", "[]", value.NewNoneVal(), 0, time.Now(), nil)
		}
		return value.NewNoneVal(), nil
	}

	position := 0
	lastResult := value.NewNoneVal()

	for position < len(vals) {
		newPos, result, err := e.EvaluateExpression(vals, position)
		if err != nil {
			if e.traceEnabled {
				e.emitTraceResult("block-exit", "", fmt.Sprintf("[error at position %d]", position), value.NewNoneVal(), position, time.Now(), err)
			}
			return value.NewNoneVal(), e.annotateError(err, vals, position)
		}
		position = newPos
		lastResult = result
	}

	if e.traceEnabled {
		e.emitTraceResult("block-exit", "", fmt.Sprintf("[%d expressions]", len(vals)), lastResult, len(vals), time.Now(), nil)
	}

	return lastResult, nil
}

// isNextInfixOperator checks if the next element is an infix operator
func (e *Evaluator) isNextInfixOperator(block []core.Value, position int) bool {
	if position >= len(block) {
		return false
	}

	nextElement := block[position]
	if nextElement.GetType() != value.TypeWord {
		return false
	}

	word, ok := value.AsWordValue(nextElement)
	if !ok {
		return false
	}

	resolved, found := e.Lookup(word)
	if !found {
		return false
	}

	fn, ok := value.AsFunctionValue(resolved)
	return ok && fn.Infix
}

func (e *Evaluator) consumeInfixOperator(block []core.Value, position int, leftOperand core.Value) (int, core.Value, error) {
	wordElement := block[position]
	word, _ := value.AsWordValue(wordElement)
	resolved, _ := e.Lookup(word)
	fn, _ := value.AsFunctionValue(resolved)

	name := functionDisplayName(fn)
	e.pushCall(name)
	defer e.popCall()

	positional, _ := e.separateParameters(fn)
	if len(positional) == 0 {
		return position, value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{functionDisplayName(fn), "0", "1 (infix requires at least one parameter)"},
		)
	}

	posArgs, refValues, newPos, err := e.collectFunctionArgs(fn, block, position+1, 1, true)
	if err != nil {
		return position, value.NewNoneVal(), e.annotateError(err, block, position)
	}

	posArgs[0] = leftOperand

	if fn.Type == value.FuncNative {
		result, err := e.callNative(fn, posArgs, refValues)
		if err != nil {
			return position, value.NewNoneVal(), e.annotateError(err, block, position)
		}
		return newPos, result, nil
	}

	result, err := e.executeFunction(fn, posArgs, refValues)
	if err != nil {
		return position, value.NewNoneVal(), err
	}
	return newPos, result, nil
}

// evaluateElement evaluates a single element without lookahead
func (e *Evaluator) evaluateElement(block []core.Value, position int) (int, core.Value, error) {
	if position >= len(block) {
		return position, value.NewNoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"missing expression", "", ""})
	}

	element := block[position]

	var traceStart time.Time
	if e.traceEnabled {
		traceStart = time.Now()
	}

	shouldTraceExpr := e.traceShouldTraceExpr

	switch element.GetType() {
	case value.TypeInteger, value.TypeString, value.TypeLogic,
		value.TypeNone, value.TypeDecimal, value.TypeObject,
		value.TypePort, value.TypeDatatype, value.TypeBlock,
		value.TypeFunction:
		if shouldTraceExpr {
			e.emitTraceResult("eval", "", element.Form(), element, position, traceStart, nil)
		}
		return position + 1, element, nil

	case value.TypeParen:
		block, _ := value.AsBlockValue(element)
		if shouldTraceExpr {
			e.emitTraceResult("eval", "paren", fmt.Sprintf("(%s)", block.Form()), value.NewNoneVal(), position, traceStart, nil)
		}
		result, err := e.DoBlock(block.Elements)
		if shouldTraceExpr && err == nil {
			e.emitTraceResult("eval", "paren", fmt.Sprintf("(%s)", block.Form()), result, position, traceStart, nil)
		}
		return position + 1, result, err

	case value.TypeLitWord:
		wordStr, _ := value.AsWordValue(element)
		result := value.NewWordVal(wordStr)
		if shouldTraceExpr {
			e.emitTraceResult("eval", wordStr, fmt.Sprintf("'%s", wordStr), result, position, traceStart, nil)
		}
		return position + 1, result, nil

	case value.TypeGetWord:
		wordStr, _ := value.AsWordValue(element)
		result, ok := e.Lookup(wordStr)
		if !ok {
			err := verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
			if shouldTraceExpr {
				e.emitTraceResult("eval", wordStr, fmt.Sprintf(":%s", wordStr), value.NewNoneVal(), position, traceStart, err)
			}
			return position, value.NewNoneVal(), err
		}
		if shouldTraceExpr {
			e.emitTraceResult("eval", wordStr, fmt.Sprintf(":%s", wordStr), result, position, traceStart, nil)
		}
		return position + 1, result, nil

	case value.TypeGetPath:
		getPath, _ := value.AsGetPath(element)
		result, err := e.evalGetPathValue(getPath)
		if shouldTraceExpr {
			e.emitTraceResult("eval", "", getPath.Mold(), result, position, traceStart, err)
		}
		return position + 1, result, err

	case value.TypeSetPath:
		setPath, _ := value.AsSetPath(element)
		newPos, result, err := e.evalSetPathValue(block, position, setPath)
		if shouldTraceExpr {
			e.emitTraceResult("eval", "", setPath.Mold(), result, position, traceStart, err)
		}
		return newPos, result, err

	case value.TypeSetWord:
		wordStr, _ := value.AsWordValue(element)

		if position+1 >= len(block) {
			err := verror.NewScriptError(
				verror.ErrIDNoValue,
				[3]string{wordStr, "set-word-without-value", wordStr},
			)
			if shouldTraceExpr {
				e.emitTraceResult("eval", wordStr, fmt.Sprintf("%s:", wordStr), value.NewNoneVal(), position, traceStart, err)
			}
			return position, value.NewNoneVal(), err
		}

		newPos, result, err := e.EvaluateExpression(block, position+1)
		if err != nil {
			if shouldTraceExpr {
				e.emitTraceResult("eval", wordStr, fmt.Sprintf("%s:", wordStr), value.NewNoneVal(), position, traceStart, err)
			}
			return position, value.NewNoneVal(), e.annotateError(err, block, position)
		}

		if result.GetType() == value.TypeFunction {
			if fnVal, ok := value.AsFunctionValue(result); ok && fnVal.Name == "" {
				fnVal.Name = wordStr
			}
		}

		currentFrame := e.currentFrame()
		currentFrame.Bind(wordStr, result)

		if shouldTraceExpr {
			e.emitTraceResult("eval", wordStr, fmt.Sprintf("%s:", wordStr), result, position, traceStart, nil)
		}

		return newPos, result, nil

	case value.TypeWord:
		wordStr, _ := value.AsWordValue(element)

		if debug.GlobalDebugger != nil && debug.GlobalDebugger.HasBreakpoint(wordStr) {
			if e.traceEnabled {
				trace.GlobalTraceSession.Emit(trace.TraceEvent{
					Timestamp: time.Now(),
					Word:      "debug",
					Value:     fmt.Sprintf("breakpoint hit: %s", wordStr),
					Duration:  0,
				})
			}
		}

		resolved, found := e.Lookup(wordStr)
		if !found {
			err := verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
			if shouldTraceExpr {
				e.emitTraceResult("eval", wordStr, wordStr, value.NewNoneVal(), position, traceStart, err)
			}
			return position, value.NewNoneVal(), err
		}

		if resolved.GetType() == value.TypeFunction {
			fn, _ := value.AsFunctionValue(resolved)
			newPos, result, err := e.invokeFunctionExpression(block, position, fn)
			return newPos, result, err
		} else {
			if shouldTraceExpr {
				e.emitTraceResult("eval", wordStr, wordStr, resolved, position, traceStart, nil)
			}
			return position + 1, resolved, nil
		}

	case value.TypePath:
		path, _ := value.AsPath(element)
		result, err := e.evalPathValue(path)
		if err != nil {
			if shouldTraceExpr {
				e.emitTraceResult("eval", "", path.Mold(), value.NewNoneVal(), position, traceStart, err)
			}
			return position, value.NewNoneVal(), err
		}

		if result.GetType() == value.TypeFunction {
			fn, _ := value.AsFunctionValue(result)
			newPos, result, err := e.invokeFunctionExpression(block, position, fn)
			return newPos, result, err
		} else {
			if shouldTraceExpr {
				e.emitTraceResult("eval", "", path.Mold(), result, position, traceStart, nil)
			}
			return position + 1, result, nil
		}

	default:
		return position, value.NewNoneVal(), verror.NewInternalError("unknown value type in evaluateExpression", [3]string{})
	}
}

// EvaluateExpression evaluates a single expression from a block starting at the given position.
// Handles infix operator lookahead internally for proper left-to-right evaluation.
// Returns the new position after consuming the expression, the result value, and any error.
func (e *Evaluator) EvaluateExpression(block []core.Value, position int) (int, core.Value, error) {
	newPos, result, err := e.evaluateElement(block, position)
	if err != nil {
		return position, value.NewNoneVal(), err
	}

	for newPos < len(block) {
		if !e.isNextInfixOperator(block, newPos) {
			break
		}

		nextPos, nextResult, err := e.consumeInfixOperator(block, newPos, result)
		if err != nil {
			return position, value.NewNoneVal(), err
		}

		newPos = nextPos
		result = nextResult
	}

	return newPos, result, nil
}

// evalSetPathValue handles set-path assignment in expression evaluation.
func (e *Evaluator) evalSetPathValue(block []core.Value, position int, setPath *value.SetPathExpression) (int, core.Value, error) {
	newPos, result, err := e.EvaluateExpression(block, position+1)
	if err != nil {
		return position, value.NewNoneVal(), e.annotateError(err, block, position)
	}

	tr, err := traversePath(e, setPath.PathExpression, true)
	if err != nil {
		return position, value.NewNoneVal(), err
	}

	result, err = e.assignToPathTarget(tr, result, setPath.Mold())
	if err != nil {
		return position, value.NewNoneVal(), err
	}

	return newPos, result, nil
}

// invokeFunctionExpression invokes a function starting at the given position in a block.
// Returns the new position after consuming the function and its arguments, the result value, and any error.
func (e *Evaluator) invokeFunctionExpression(block []core.Value, position int, fn *value.FunctionValue) (int, core.Value, error) {
	name := functionDisplayName(fn)
	e.pushCall(name)
	defer e.popCall()

	posArgs, refValues, newPos, err := e.collectFunctionArgs(fn, block, position+1, 0, false)
	if err != nil {
		return position, value.NewNoneVal(), e.annotateError(err, block, position)
	}

	var traceStart time.Time
	var args map[string]string
	if e.traceEnabled {
		traceStart = time.Now()
		args = e.captureFunctionArgs(fn, posArgs, refValues)
		event := trace.TraceEvent{
			Timestamp:  traceStart,
			Value:      "",
			Word:       name,
			Duration:   0,
			EventType:  "call",
			Step:       trace.GlobalTraceSession.NextStep(),
			Depth:      len(e.callStack) - 1,
			Position:   position,
			Expression: name,
			Args:       args,
		}
		trace.GlobalTraceSession.Emit(event)
	}

	var result core.Value
	if fn.Type == value.FuncNative {
		result, err = e.callNative(fn, posArgs, refValues)
		if err != nil {
			if e.traceEnabled {
				e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, err)
			}
			return position, value.NewNoneVal(), e.annotateError(err, block, position)
		}
	} else {
		result, err = e.executeFunction(fn, posArgs, refValues)
		if err != nil {
			if e.traceEnabled {
				e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, err)
			}
			return position, value.NewNoneVal(), err
		}
	}

	if e.traceEnabled {
		e.emitTraceResult("return", name, name, result, position, traceStart, nil)
	}

	return newPos, result, nil
}

func (e *Evaluator) collectParameter(block []core.Value, position int, paramSpec value.ParamSpec, useElementEval bool) (int, core.Value, error) {
	if position >= len(block) {
		return position, value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDNoValue,
			[3]string{"parameter expected", "", ""},
		)
	}

	if paramSpec.Eval {
		if useElementEval {
			return e.evaluateElement(block, position)
		}
		return e.EvaluateExpression(block, position)
	}

	return position + 1, block[position], nil
}

func (e *Evaluator) separateParameters(fn *value.FunctionValue) ([]value.ParamSpec, map[string]value.ParamSpec) {
	positional := make([]value.ParamSpec, 0, len(fn.Params))
	refinements := make(map[string]value.ParamSpec)

	for _, spec := range fn.Params {
		if spec.Refinement {
			refinements[spec.Name] = spec
		} else {
			positional = append(positional, spec)
		}
	}

	return positional, refinements
}

func (e *Evaluator) initializeRefinements(refSpecs map[string]value.ParamSpec) map[string]core.Value {
	refValues := make(map[string]core.Value, len(refSpecs))
	for name, spec := range refSpecs {
		if spec.TakesValue {
			refValues[name] = value.NewNoneVal()
		} else {
			refValues[name] = value.NewLogicVal(false)
		}
	}
	return refValues
}

func (e *Evaluator) collectFunctionArgs(fn *value.FunctionValue, block []core.Value, startPosition int, startParamIndex int, useElementEval bool) ([]core.Value, map[string]core.Value, int, error) {
	positional, refSpecs := e.separateParameters(fn)
	refValues := e.initializeRefinements(refSpecs)
	refProvided := make(map[string]bool)

	posArgs := make([]core.Value, len(positional))
	position := startPosition
	paramIndex := startParamIndex

	for paramIndex < len(positional) {
		var err error
		position, err = e.readRefinements(block, position, refSpecs, refValues, refProvided)
		if err != nil {
			return nil, nil, position, err
		}

		if position >= len(block) {
			return nil, nil, position, verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{functionDisplayName(fn), strconv.Itoa(len(positional)), strconv.Itoa(paramIndex)},
			)
		}

		paramSpec := positional[paramIndex]
		position, posArgs[paramIndex], err = e.collectParameter(block, position, paramSpec, useElementEval)
		if err != nil {
			return nil, nil, position, err
		}

		paramIndex++
	}

	var err error
	position, err = e.readRefinements(block, position, refSpecs, refValues, refProvided)
	if err != nil {
		return nil, nil, position, err
	}

	return posArgs, refValues, position, nil
}

// evalPathValue evaluates a path expression and returns its value.
func (e *Evaluator) evalPathValue(path *value.PathExpression) (core.Value, error) {
	tr, err := traversePath(e, path, false)
	if err != nil {
		return value.NewNoneVal(), err
	}
	result := tr.values[len(tr.values)-1]

	// Path expressions return functions without invoking them
	// (they can be invoked later with arguments like: obj.method arg1 arg2)

	return result, nil
}

// evalGetPathValue evaluates a get-path expression and returns its value.
// Get-paths NEVER invoke functions - they just return the result.
func (e *Evaluator) evalGetPathValue(getPath *value.GetPathExpression) (core.Value, error) {
	tr, err := traversePath(e, getPath.PathExpression, false)
	if err != nil {
		return value.NewNoneVal(), err
	}
	// Get-paths NEVER invoke functions - just return the result
	return tr.values[len(tr.values)-1], nil
}

// callNative invokes a native function with positional and refinement arguments.
// Native functions are implemented in Go and called directly.
// Returns the result of the native function call or an error.
func (e *Evaluator) callNative(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value) (core.Value, error) {
	if fn.Type != value.FuncNative {
		return value.NewNoneVal(), verror.NewInternalError("callNative expects native function", [3]string{})
	}

	result, err := fn.Native(posArgs, refValues, e)

	if err == nil {
		return result, nil
	}
	if verr, ok := err.(*verror.Error); ok {
		return result, verr
	}
	return value.NewNoneVal(), verror.NewInternalError(err.Error(), [3]string{})
}

// isRefinement checks if a value represents a function refinement.
// Refinements are words that start with "--" and modify function behavior.
func isRefinement(val core.Value) bool {
	if val.GetType() != value.TypeWord {
		return false
	}
	wordStr, ok := value.AsWordValue(val)
	if !ok {
		return false
	}
	return strings.HasPrefix(wordStr, "--")
}

// refinementError creates a script error for refinement-related issues.
func refinementError(kind, refName string) error {
	var msg string
	switch kind {
	case "unknown":
		msg = fmt.Sprintf("Unknown refinement: --%s", refName)
	case "duplicate":
		msg = fmt.Sprintf("Duplicate refinement: --%s", refName)
	case "missing-value":
		msg = fmt.Sprintf("Refinement --%s requires a value", refName)
	}
	return verror.NewScriptError(verror.ErrIDInvalidOperation, [3]string{msg, "", ""})
}

// readRefinements parses refinement arguments from a token sequence.
// Refinements can be flags (boolean) or take values.
// Updates the refinement values map and tracks which refinements have been provided.
func (e *Evaluator) readRefinements(tokens []core.Value, pos int, refSpecs map[string]value.ParamSpec, refValues map[string]core.Value, refProvided map[string]bool) (int, error) {
	for pos < len(tokens) && isRefinement(tokens[pos]) {
		wordStr, _ := value.AsWordValue(tokens[pos])
		refName := strings.TrimPrefix(wordStr, "--")

		spec, exists := refSpecs[refName]
		if !exists {
			return pos, refinementError("unknown", refName)
		}

		if refProvided[refName] {
			return pos, refinementError("duplicate", refName)
		}

		// Handle refinements that take values vs. boolean flags
		if spec.TakesValue {
			if pos+1 >= len(tokens) {
				return pos, refinementError("missing-value", refName)
			}
			var arg core.Value
			var err error
			pos, arg, err = e.evaluateElement(tokens, pos+1)
			if err != nil {
				return pos, err
			}
			refValues[refName] = arg
		} else {
			refValues[refName] = value.NewLogicVal(true)
			pos++
		}

		refProvided[refName] = true
	}

	return pos, nil
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
	e.PushFrameContext(frame)
	defer e.popFrame()

	posIndex := 0
	for _, spec := range fn.Params {
		if spec.Refinement {
			val, ok := refinements[spec.Name]
			if !ok {
				if spec.TakesValue {
					val = value.NewNoneVal()
				} else {
					val = value.NewLogicVal(false)
				}
			}
			frame.Bind(spec.Name, val)
			continue
		}

		frame.Bind(spec.Name, posArgs[posIndex])
		posIndex++
	}

	if fn.Body == nil {
		return value.NewNoneVal(), verror.NewInternalError("function body missing", [3]string{})
	}

	result, err := e.DoBlock(fn.Body.Elements)
	if err != nil {
		return value.NewNoneVal(), err
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
		base = value.NewIntVal(num)
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

			// Search field in object and prototype chain using owned frames
			fieldVal, found := obj.GetFieldWithProto(fieldName)
			if !found {
				return nil, verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{fieldName, "", ""})
			}

			tr.values = append(tr.values, fieldVal)

		case value.PathSegmentIndex:
			// Collection indexing
			index, ok := seg.Value.(int64)
			if !ok {
				return nil, verror.NewInternalError("index segment does not contain int64", [3]string{})
			}

			if current.GetType() == value.TypeBlock {
				block, _ := value.AsBlockValue(current)
				if err := checkIndexBounds(index, int64(len(block.Elements)), "block"); err != nil {
					return nil, err
				}
				tr.values = append(tr.values, block.Elements[index-1])

			} else if current.GetType() == value.TypeString {
				str, _ := value.AsStringValue(current)
				runes := []rune(str.String())
				if err := checkIndexBounds(index, int64(len(runes)), "string"); err != nil {
					return nil, err
				}
				tr.values = append(tr.values, value.NewStrVal(string(runes[index-1])))

			} else if current.GetType() == value.TypeBinary {
				bin, _ := value.AsBinaryValue(current)
				if err := checkIndexBounds(index, int64(bin.Length()), "binary"); err != nil {
					return nil, err
				}
				tr.values = append(tr.values, value.NewIntVal(int64(bin.At(int(index-1)))))

			} else {
				return nil, verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index requires block, string, or binary type", "", ""})
			}

		default:
			return nil, verror.NewInternalError("unsupported path segment type", [3]string{})
		}
	}

	return tr, nil
}

// checkIndexBounds validates that an index is within bounds (1-based indexing).
// Returns nil if valid, error if out of bounds.
func checkIndexBounds(index, length int64, typeName string) error {
	if index < 1 || index > length {
		return verror.NewScriptError(verror.ErrIDIndexOutOfRange,
			[3]string{fmt.Sprintf("index %d out of range for %s of length %d", index, typeName, length), "", ""})
	}
	return nil
}

// assignToPathTarget assigns a value to the target location identified by a path traversal.
// Supports assignment to object fields and block/string indices.
// Validates that the target is assignable and within bounds.
func (e *Evaluator) assignToPathTarget(tr *pathTraversal, newVal core.Value, pathStr string) (core.Value, error) {
	if len(tr.segments) < 2 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"set-path requires at least 2 segments", "", ""})
	}

	container := tr.values[len(tr.values)-1]
	finalSeg := tr.segments[len(tr.segments)-1]

	// Cannot assign to paths starting with numeric literals
	if tr.segments[0].Type == value.PathSegmentIndex {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{pathStr, "", ""})
	}

	if container.GetType() == value.TypeNone {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDNonePath, [3]string{"cannot assign to none value", "", ""})
	}

	switch finalSeg.Type {
	case value.PathSegmentIndex:
		// Assign to collection element
		index, ok := finalSeg.Value.(int64)
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError("index segment does not contain int64", [3]string{})
		}

		if container.GetType() != value.TypeBlock {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{"index assignment requires block type", "", ""})
		}

		block, _ := value.AsBlockValue(container)
		if err := checkIndexBounds(index, int64(len(block.Elements)), "block"); err != nil {
			return value.NewNoneVal(), err
		}
		block.Elements[index-1] = newVal

	case value.PathSegmentWord:
		// Assign to object field
		fieldName, ok := finalSeg.Value.(string)
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError("word segment does not contain string", [3]string{})
		}

		if container.GetType() != value.TypeObject {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDImmutableTarget, [3]string{"cannot assign field to non-object", "", ""})
		}

		obj, _ := value.AsObject(container)

		// Set field using owned frame (creates field if it doesn't exist)
		obj.SetField(fieldName, newVal)

	default:
		return value.NewNoneVal(), verror.NewInternalError("unsupported path segment type for assignment", [3]string{})
	}

	return newVal, nil
}

// captureFrameState returns a map of variable bindings in the current frame for tracing.
func (e *Evaluator) captureFrameState() map[string]string {
	if !e.traceEnabled || trace.GlobalTraceSession == nil || !trace.GlobalTraceSession.GetVerbose() {
		return nil
	}

	currentFrame := e.currentFrame()
	bindings := currentFrame.GetAll()
	result := make(map[string]string, len(bindings))

	for _, binding := range bindings {
		result[binding.Symbol] = binding.Value.Form()
	}

	return result
}

// captureFunctionArgs extracts function argument names and values for tracing.
func (e *Evaluator) captureFunctionArgs(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value) map[string]string {
	if !e.traceEnabled || trace.GlobalTraceSession == nil || !trace.GlobalTraceSession.GetIncludeArgs() {
		return nil
	}

	result := make(map[string]string)

	positional, _ := e.separateParameters(fn)
	for i, param := range positional {
		if i < len(posArgs) {
			result[param.Name] = posArgs[i].Form()
		}
	}

	for name, val := range refValues {
		result[name] = val.Form()
	}

	return result
}

// emitTraceResult emits a trace event with comprehensive debugging information.
func (e *Evaluator) emitTraceResult(eventType string, word string, expr string, result core.Value, position int, traceStart time.Time, err error) {
	if !e.traceEnabled || trace.GlobalTraceSession == nil {
		return
	}

	depth := len(e.callStack) - 1
	if !trace.GlobalTraceSession.ShouldTraceAtDepth(depth) {
		return
	}

	var frameState map[string]string
	if trace.GlobalTraceSession.GetVerbose() {
		frameState = e.captureFrameState()
	}

	event := trace.TraceEvent{
		Timestamp:  traceStart,
		Value:      result.Form(),
		Word:       word,
		Duration:   time.Since(traceStart).Nanoseconds(),
		EventType:  eventType,
		Step:       trace.GlobalTraceSession.NextStep(),
		Depth:      depth,
		Position:   position,
		Expression: expr,
		Frame:      frameState,
	}

	if err != nil {
		event.Error = err.Error()
	}

	trace.GlobalTraceSession.Emit(event)
}
