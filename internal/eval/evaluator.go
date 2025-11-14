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

func (e *Evaluator) SetOutputWriter(w io.Writer) {
	if w == nil {
		e.OutputWriter = os.Stdout
	} else {
		e.OutputWriter = w
	}
}

func (e *Evaluator) GetOutputWriter() io.Writer {
	return e.OutputWriter
}

func (e *Evaluator) SetErrorWriter(w io.Writer) {
	if w == nil {
		e.ErrorWriter = os.Stderr
	} else {
		e.ErrorWriter = w
	}
}

func (e *Evaluator) GetErrorWriter() io.Writer {
	return e.ErrorWriter
}

func (e *Evaluator) SetInputReader(r io.Reader) {
	if r == nil {
		e.InputReader = os.Stdin
	} else {
		e.InputReader = r
	}
}

func (e *Evaluator) GetInputReader() io.Reader {
	return e.InputReader
}

func (e *Evaluator) UpdateTraceCache() {
	if trace.GlobalTraceSession == nil {
		e.traceEnabled = false
		e.traceShouldTraceExpr = false
		return
	}
	e.traceEnabled = trace.GlobalTraceSession.IsEnabled()
	e.traceShouldTraceExpr = e.traceEnabled && trace.GlobalTraceSession.ShouldTraceExpression()
}

func (e *Evaluator) currentFrame() core.Frame {
	if len(e.Frames) == 0 {
		return nil
	}
	return e.Frames[len(e.Frames)-1]
}

func (e *Evaluator) currentFrameIndex() int {
	if len(e.Frames) == 0 {
		return -1
	}
	current := e.Frames[len(e.Frames)-1]
	return current.GetIndex()
}

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

func (e *Evaluator) annotateError(err error, vals []core.Value, locations []core.SourceLocation, idx int) error {
	if err == nil {
		return nil
	}
	if verr, ok := err.(*verror.Error); ok {
		if idx >= 0 && idx < len(vals) && verr.Near == "" {
			verr.SetNear(verror.CaptureNear(vals, idx))
		}
		if idx >= 0 && idx < len(locations) && verr.File == "" {
			loc := locations[idx]
			if loc.Line != 0 || loc.Column != 0 || loc.File != "" {
				file := loc.File
				verr.SetLocation(file, loc.Line, loc.Column)
			}
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

func functionDisplayName(fn *value.FunctionValue) string {
	if fn == nil || fn.Name == "" {
		return "(anonymous)"
	}
	return fn.Name
}

func (e *Evaluator) MarkFrameCaptured(idx int) {
	if idx >= 0 {
		e.captured[idx] = true
	}
}

func (e *Evaluator) CurrentFrameIndex() int {
	return e.currentFrameIndex()
}

func (e *Evaluator) RegisterFrame(f core.Frame) int {
	if f.GetIndex() >= 0 {
		return f.GetIndex()
	}

	idx := len(e.frameStore)
	e.frameStore = append(e.frameStore, f)
	f.SetIndex(idx)
	return idx
}

func (e *Evaluator) GetFrameByIndex(idx int) core.Frame {
	if idx < 0 || idx >= len(e.frameStore) {
		return nil
	}
	return e.frameStore[idx]
}

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

func (e *Evaluator) PopFrameContext() {
	e.popFrame()
}

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

func (e *Evaluator) GetCallStack() []string {
	if len(e.callStack) == 0 {
		return []string{}
	}
	stack := make([]string, len(e.callStack))
	copy(stack, e.callStack)
	return stack
}

func (e *Evaluator) DoBlock(vals []core.Value, locations []core.SourceLocation) (core.Value, error) {
	var traceStart time.Time
	if e.traceEnabled {
		traceStart = time.Now()
		e.emitTraceResult("block-enter", "", fmt.Sprintf("[%d expressions]", len(vals)), value.NewNoneVal(), 0, traceStart, nil)
	}

	if len(locations) != len(vals) {
		locations = nil
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
		newPos, result, err := e.EvaluateExpression(vals, locations, position)
		if err != nil {
			if e.traceEnabled {
				e.emitTraceResult("block-exit", "", fmt.Sprintf("[error at position %d]", position), value.NewNoneVal(), position, time.Now(), err)
			}
			return value.NewNoneVal(), e.annotateError(err, vals, locations, position)
		}
		position = newPos
		lastResult = result
	}

	if e.traceEnabled {
		e.emitTraceResult("block-exit", "", fmt.Sprintf("[%d expressions]", len(vals)), lastResult, len(vals), time.Now(), nil)
	}

	return lastResult, nil
}

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

func (e *Evaluator) consumeInfixOperator(block []core.Value, locations []core.SourceLocation, position int, leftOperand core.Value) (int, core.Value, error) {
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

	posArgs, refValues, newPos, err := e.collectFunctionArgs(fn, block, locations, position+1, 1, true)
	if err != nil {
		return position, value.NewNoneVal(), e.annotateError(err, block, locations, position)
	}

	posArgs[0] = leftOperand

	if fn.Type == value.FuncNative {
		result, err := e.callNative(fn, posArgs, refValues)
		if err != nil {
			return position, value.NewNoneVal(), e.annotateError(err, block, locations, position)
		}
		return newPos, result, nil
	}

	result, err := e.executeFunction(fn, posArgs, refValues)
	if err != nil {
		return position, value.NewNoneVal(), err
	}
	return newPos, result, nil
}

func (e *Evaluator) evaluateSetWord(block []core.Value, locations []core.SourceLocation, element core.Value, position int, traceStart time.Time, shouldTraceExpr bool) (int, core.Value, error) {
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

	newPos, result, err := e.EvaluateExpression(block, locations, position+1)
	if err != nil {
		if shouldTraceExpr {
			e.emitTraceResult("eval", wordStr, fmt.Sprintf("%s:", wordStr), value.NewNoneVal(), position, traceStart, err)
		}
		return position, value.NewNoneVal(), e.annotateError(err, block, locations, position)
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
}

func (e *Evaluator) evaluateWord(block []core.Value, locations []core.SourceLocation, element core.Value, position int, traceStart time.Time, shouldTraceExpr bool) (int, core.Value, error) {
	wordStr, _ := value.AsWordValue(element)

	if debug.GlobalDebugger != nil {
		debug.GlobalDebugger.HandleBreakpoint(wordStr, position, len(e.callStack)-1)
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
		newPos, result, err := e.invokeFunctionExpression(block, locations, position, fn)
		return newPos, result, err
	}

	if shouldTraceExpr {
		e.emitTraceResult("eval", wordStr, wordStr, resolved, position, traceStart, nil)
	}

	return position + 1, resolved, nil
}

func (e *Evaluator) evaluatePath(block []core.Value, locations []core.SourceLocation, element core.Value, position int, traceStart time.Time, shouldTraceExpr bool) (int, core.Value, error) {
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
		newPos, result, err := e.invokeFunctionExpression(block, locations, position, fn)
		return newPos, result, err
	}

	if shouldTraceExpr {
		e.emitTraceResult("eval", "", path.Mold(), result, position, traceStart, nil)
	}

	return position + 1, result, nil
}

func (e *Evaluator) evaluateElement(block []core.Value, locations []core.SourceLocation, position int) (int, core.Value, error) {
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
	case value.TypeInteger, value.TypeLogic,
		value.TypeNone, value.TypeDecimal, value.TypeObject,
		value.TypePort, value.TypeDatatype,
		value.TypeFunction:
		if shouldTraceExpr {
			e.emitTraceResult("eval", "", element.Form(), element, position, traceStart, nil)
		}
		return position + 1, element, nil
	case value.TypeBlock, value.TypeBinary, value.TypeString:
		if element.GetType() == value.TypeBlock {
			if blockVal, ok := value.AsBlockValue(element); ok && blockVal.Length() == 0 {
				cloned := blockVal.Clone()
				if shouldTraceExpr {
					e.emitTraceResult("eval", "", element.Form(), cloned, position, traceStart, nil)
				}
				return position + 1, cloned, nil
			}
		} else if element.GetType() == value.TypeBinary {
			if binaryVal, ok := value.AsBinaryValue(element); ok && binaryVal.Length() == 0 {
				cloned := binaryVal.Clone()
				if shouldTraceExpr {
					e.emitTraceResult("eval", "", element.Form(), cloned, position, traceStart, nil)
				}
				return position + 1, cloned, nil
			}
		} else if element.GetType() == value.TypeString {
			if stringVal, ok := value.AsStringValue(element); ok && stringVal.Length() == 0 {
				cloned := stringVal.Clone()
				if shouldTraceExpr {
					e.emitTraceResult("eval", "", element.Form(), cloned, position, traceStart, nil)
				}
				return position + 1, cloned, nil
			}
		}
		if shouldTraceExpr {
			e.emitTraceResult("eval", "", element.Form(), element, position, traceStart, nil)
		}
		return position + 1, element, nil

	case value.TypeParen:
		parenBlock, _ := value.AsBlockValue(element)
		if shouldTraceExpr {
			e.emitTraceResult("eval", "paren", fmt.Sprintf("(%s)", parenBlock.Form()), value.NewNoneVal(), position, traceStart, nil)
		}
		nestedLocations := []core.SourceLocation{}
		if parenBlock != nil {
			nestedLocations = parenBlock.Locations()
		}
		result, err := e.DoBlock(parenBlock.Elements, nestedLocations)
		if shouldTraceExpr && err == nil {
			e.emitTraceResult("eval", "paren", fmt.Sprintf("(%s)", parenBlock.Form()), result, position, traceStart, nil)
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
		newPos, result, err := e.evalSetPathValue(block, locations, position, setPath)
		if shouldTraceExpr {
			e.emitTraceResult("eval", "", setPath.Mold(), result, position, traceStart, err)
		}
		return newPos, result, err

	case value.TypeSetWord:
		return e.evaluateSetWord(block, locations, element, position, traceStart, shouldTraceExpr)

	case value.TypeWord:
		return e.evaluateWord(block, locations, element, position, traceStart, shouldTraceExpr)

	case value.TypePath:
		return e.evaluatePath(block, locations, element, position, traceStart, shouldTraceExpr)

	default:
		return position, value.NewNoneVal(), verror.NewInternalError("unknown value type in evaluateExpression", [3]string{})
	}
}

func (e *Evaluator) EvaluateExpression(block []core.Value, locations []core.SourceLocation, position int) (int, core.Value, error) {
	if len(locations) != len(block) {
		locations = nil
	}

	newPos, result, err := e.evaluateElement(block, locations, position)
	if err != nil {
		return position, value.NewNoneVal(), err
	}

	for newPos < len(block) {
		if !e.isNextInfixOperator(block, newPos) {
			break
		}

		nextPos, nextResult, err := e.consumeInfixOperator(block, locations, newPos, result)
		if err != nil {
			return position, value.NewNoneVal(), err
		}

		newPos = nextPos
		result = nextResult
	}

	return newPos, result, nil
}

func (e *Evaluator) evalSetPathValue(block []core.Value, locations []core.SourceLocation, position int, setPath *value.SetPathExpression) (int, core.Value, error) {
	newPos, result, err := e.EvaluateExpression(block, locations, position+1)
	if err != nil {
		return position, value.NewNoneVal(), e.annotateError(err, block, locations, position)
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

func (e *Evaluator) setupFunctionCallTracing(name string, position int, posArgs []core.Value, refValues map[string]core.Value) (time.Time, map[string]string) {
	var traceStart time.Time
	var args map[string]string
	if e.traceEnabled {
		traceStart = time.Now()
		args = e.captureFunctionArgs(nil, posArgs, refValues) // fn is not needed for arg capture
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
	return traceStart, args
}

func (e *Evaluator) callNativeFunction(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value, name string, position int, traceStart time.Time) (core.Value, error) {
	result, err := e.callNative(fn, posArgs, refValues)
	if err != nil {
		if e.traceEnabled {
			e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, err)
		}
		return value.NewNoneVal(), e.annotateError(err, nil, nil, position)
	}
	return result, nil
}

func (e *Evaluator) callUserDefinedFunction(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value, name string, position int, traceStart time.Time) (core.Value, error) {
	result, err := e.executeFunction(fn, posArgs, refValues)
	if err != nil {
		convertedErr := verror.ConvertLoopControlSignal(err)
		if convertedErr != err {
			if e.traceEnabled {
				e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, convertedErr)
			}
			return value.NewNoneVal(), convertedErr
		}
		if e.traceEnabled {
			e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, err)
		}
		return value.NewNoneVal(), err
	}
	return result, nil
}

func (e *Evaluator) invokeFunctionExpression(block []core.Value, locations []core.SourceLocation, position int, fn *value.FunctionValue) (int, core.Value, error) {
	name := functionDisplayName(fn)
	e.pushCall(name)
	defer e.popCall()

	posArgs, refValues, newPos, err := e.collectFunctionArgs(fn, block, locations, position+1, 0, false)
	if err != nil {
		return position, value.NewNoneVal(), e.annotateError(err, block, locations, position)
	}

	traceStart, _ := e.setupFunctionCallTracing(name, position, posArgs, refValues)

	var result core.Value
	if fn.Type == value.FuncNative {
		result, err = e.callNativeFunction(fn, posArgs, refValues, name, position, traceStart)
	} else {
		result, err = e.callUserDefinedFunction(fn, posArgs, refValues, name, position, traceStart)
	}

	if err != nil {
		return position, value.NewNoneVal(), err
	}

	if e.traceEnabled {
		e.emitTraceResult("return", name, name, result, position, traceStart, nil)
	}

	return newPos, result, nil
}

func (e *Evaluator) collectParameter(block []core.Value, locations []core.SourceLocation, position int, paramSpec value.ParamSpec, useElementEval bool) (int, core.Value, error) {
	if position >= len(block) {
		return position, value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDNoValue,
			[3]string{"parameter expected", "", ""},
		)
	}

	if paramSpec.Eval {
		if useElementEval {
			return e.evaluateElement(block, locations, position)
		}
		return e.EvaluateExpression(block, locations, position)
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

func (e *Evaluator) collectFunctionArgs(fn *value.FunctionValue, block []core.Value, locations []core.SourceLocation, startPosition int, startParamIndex int, useElementEval bool) ([]core.Value, map[string]core.Value, int, error) {
	positional, refSpecs := e.separateParameters(fn)
	refValues := e.initializeRefinements(refSpecs)
	refProvided := make(map[string]bool)

	posArgs := make([]core.Value, len(positional))
	position := startPosition
	paramIndex := startParamIndex

	position, err := e.collectPositionalArgs(fn, block, locations, positional, posArgs, refSpecs, refValues, refProvided, position, paramIndex, useElementEval)
	if err != nil {
		return nil, nil, position, err
	}

	position, err = e.readRefinements(block, locations, position, refSpecs, refValues, refProvided)
	if err != nil {
		return nil, nil, position, err
	}

	return posArgs, refValues, position, nil
}

func (e *Evaluator) collectPositionalArgs(fn *value.FunctionValue, block []core.Value, locations []core.SourceLocation, positional []value.ParamSpec, posArgs []core.Value, refSpecs map[string]value.ParamSpec, refValues map[string]core.Value, refProvided map[string]bool, startPosition, startParamIndex int, useElementEval bool) (int, error) {
	position := startPosition
	paramIndex := startParamIndex

	for paramIndex < len(positional) {
		var err error
		position, err = e.readRefinements(block, locations, position, refSpecs, refValues, refProvided)
		if err != nil {
			return position, err
		}

		if position >= len(block) {
			for i := paramIndex; i < len(positional); i++ {
				if !positional[i].Optional {
					return position, verror.NewScriptError(
						verror.ErrIDArgCount,
						[3]string{functionDisplayName(fn), strconv.Itoa(len(positional)), strconv.Itoa(paramIndex)},
					)
				}
				posArgs[i] = value.NewNoneVal()
			}
			return position, nil
		}

		paramSpec := positional[paramIndex]
		position, posArgs[paramIndex], err = e.collectParameter(block, locations, position, paramSpec, useElementEval)
		if err != nil {
			return position, err
		}

		paramIndex++
	}

	return position, nil
}

func (e *Evaluator) evalPathValue(path *value.PathExpression) (core.Value, error) {
	tr, err := traversePath(e, path, false)
	if err != nil {
		return value.NewNoneVal(), err
	}
	if len(tr.values) == 0 {
		return value.NewNoneVal(), verror.NewInternalError("path traversal returned no values", [3]string{})
	}
	return tr.values[len(tr.values)-1], nil
}

func (e *Evaluator) evalGetPathValue(getPath *value.GetPathExpression) (core.Value, error) {
	return e.evalPathValue(getPath.PathExpression)
}

func (e *Evaluator) callNative(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value) (core.Value, error) {
	if fn.Type != value.FuncNative {
		return value.NewNoneVal(), verror.NewInternalError("callNative expects native function", [3]string{})
	}

	result, err := fn.Native(posArgs, refValues, e)

	if err == nil {
		return result, nil
	}
	if _, ok := err.(*ReturnSignal); ok {
		return result, err
	}
	if verr, ok := err.(*verror.Error); ok {
		return result, verr
	}
	return value.NewNoneVal(), verror.NewInternalError(err.Error(), [3]string{})
}

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

func makePathTypeError(expected, got string, context string) error {
	msg := fmt.Sprintf("%s (got %s)", expected, got)
	return verror.NewScriptError(verror.ErrIDPathTypeMismatch, [3]string{msg, context, ""})
}

func (e *Evaluator) readRefinements(tokens []core.Value, locations []core.SourceLocation, pos int, refSpecs map[string]value.ParamSpec, refValues map[string]core.Value, refProvided map[string]bool) (int, error) {
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
			pos, arg, err = e.EvaluateExpression(tokens, locations, pos+1)
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

func (e *Evaluator) executeFunction(fn *value.FunctionValue, posArgs []core.Value, refinements map[string]core.Value) (core.Value, error) {
	parent := fn.Parent
	if parent == -1 {
		parent = 0
	}

	frame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, parent, len(fn.Params))
	frame.Name = functionDisplayName(fn)
	e.PushFrameContext(frame)
	defer e.popFrame()

	e.bindFunctionParameters(frame, fn, posArgs, refinements)

	if fn.Body == nil {
		return value.NewNoneVal(), verror.NewInternalError("function body missing", [3]string{})
	}

	result, err := e.DoBlock(fn.Body.Elements, fn.Body.Locations())
	if err != nil {
		if returnSig, ok := err.(*ReturnSignal); ok {
			return returnSig.Value(), nil
		}
		return value.NewNoneVal(), err
	}

	return result, nil
}

func (e *Evaluator) bindFunctionParameters(frame core.Frame, fn *value.FunctionValue, posArgs []core.Value, refinements map[string]core.Value) {
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
}

type pathTraversal struct {
	segments []value.PathSegment
	values   []core.Value
}

func (e *Evaluator) materializeSegment(seg value.PathSegment) (value.PathSegment, error) {
	if seg.Type != value.PathSegmentEval {
		return seg, nil
	}

	block, ok := seg.AsEvalBlock()
	if !ok {
		return value.PathSegment{}, verror.NewInternalError("eval segment missing block", [3]string{})
	}

	result, err := e.DoBlock(block.Elements, block.Locations())
	if err != nil {
		return value.PathSegment{}, err
	}

	if name, ok := value.AsWordValue(result); ok {
		return value.NewWordSegment(name), nil
	}

	if strVal, ok := value.AsStringValue(result); ok {
		if strVal.String() == "" {
			return value.PathSegment{}, verror.NewScriptError(
				verror.ErrIDEmptyPathSegment,
				[3]string{"", "eval-empty-segment", ""},
			)
		}
		return value.NewWordSegment(strVal.String()), nil
	}

	if num, ok := value.AsIntValue(result); ok {
		return value.NewIndexSegment(num), nil
	}

	return value.PathSegment{}, verror.NewScriptError(
		verror.ErrIDInvalidPath,
		[3]string{
			fmt.Sprintf("eval segment must evaluate to word, string, or integer (got %s)", value.TypeToString(result.GetType())),
			"eval-segment-type",
			"",
		},
	)
}

func (e *Evaluator) resolvePathBase(firstSeg value.PathSegment) (core.Value, error) {
	switch firstSeg.Type {
	case value.PathSegmentWord:
		wordStr, ok := firstSeg.AsWord()
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError("word segment does not contain string", [3]string{})
		}

		base, ok := e.Lookup(wordStr)
		if !ok {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
		}
		return base, nil
	case value.PathSegmentIndex:
		num, ok := firstSeg.AsIndex()
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError("index segment does not contain int64", [3]string{})
		}
		return value.NewIntVal(num), nil
	default:
		return value.NewNoneVal(), verror.NewInternalError("unexpected first segment type", [3]string{fmt.Sprintf("%v", firstSeg.Type), "", ""})
	}
}

func (e *Evaluator) traverseWordSegment(tr *pathTraversal, seg value.PathSegment, current core.Value, lenient bool) error {
	if current.GetType() != value.TypeObject {
		return makePathTypeError("word segment requires object", value.TypeToString(current.GetType()), "")
	}

	obj, ok := value.AsObject(current)
	if !ok {
		return verror.NewInternalError("failed to cast object value", [3]string{})
	}

	fieldName, ok := seg.AsWord()
	if !ok {
		return verror.NewInternalError("word segment does not contain string", [3]string{})
	}

	fieldVal, found := obj.GetFieldWithProto(fieldName)
	if !found {
		if lenient {
			tr.values = append(tr.values, value.NewNoneVal())
			return nil
		}
		return verror.NewScriptError(verror.ErrIDNoSuchField, [3]string{fieldName, "", ""})
	}

	tr.values = append(tr.values, fieldVal)
	return nil
}

func (e *Evaluator) traverseIndexSegment(tr *pathTraversal, seg value.PathSegment, current core.Value, lenient bool) error {
	index, ok := seg.AsIndex()
	if !ok {
		return verror.NewInternalError("index segment does not contain int64", [3]string{})
	}

	switch current.GetType() {
	case value.TypeBlock:
		block, ok := value.AsBlockValue(current)
		if !ok {
			return verror.NewInternalError("failed to cast block value", [3]string{})
		}
		if index < 1 || index > int64(len(block.Elements)) {
			if lenient {
				tr.values = append(tr.values, value.NewNoneVal())
				return nil
			}
			return checkIndexBounds(index, int64(len(block.Elements)), "block")
		}
		tr.values = append(tr.values, block.Elements[index-1])

	case value.TypeString:
		str, ok := value.AsStringValue(current)
		if !ok {
			return verror.NewInternalError("failed to cast string value", [3]string{})
		}
		runes := []rune(str.String())
		if index < 1 || index > int64(len(runes)) {
			if lenient {
				tr.values = append(tr.values, value.NewNoneVal())
				return nil
			}
			return checkIndexBounds(index, int64(len(runes)), "string")
		}
		tr.values = append(tr.values, value.NewStrVal(string(runes[index-1])))

	case value.TypeBinary:
		bin, ok := value.AsBinaryValue(current)
		if !ok {
			return verror.NewInternalError("failed to cast binary value", [3]string{})
		}
		if index < 1 || index > int64(bin.Length()) {
			if lenient {
				tr.values = append(tr.values, value.NewNoneVal())
				return nil
			}
			return checkIndexBounds(index, int64(bin.Length()), "binary")
		}
		tr.values = append(tr.values, value.NewIntVal(int64(bin.At(int(index-1)))))

	default:
		return makePathTypeError("index requires block, string, or binary", value.TypeToString(current.GetType()), "")
	}

	return nil
}

func traversePath(e core.Evaluator, path *value.PathExpression, stopBeforeLast bool) (*pathTraversal, error) {
	if len(path.Segments) == 0 {
		return nil, verror.NewScriptError(verror.ErrIDInvalidPath, [3]string{"empty path", "empty", ""})
	}

	eval, ok := e.(*Evaluator)
	if !ok {
		return nil, verror.NewInternalError("evaluator type mismatch", [3]string{})
	}

	resolved := make([]value.PathSegment, len(path.Segments))
	copy(resolved, path.Segments)

	tr := &pathTraversal{
		segments: resolved,
		values:   make([]core.Value, 0, len(path.Segments)),
	}

	firstSeg, err := eval.materializeSegment(resolved[0])
	if err != nil {
		return nil, err
	}
	resolved[0] = firstSeg

	base, err := eval.resolvePathBase(firstSeg)
	if err != nil {
		return nil, err
	}

	tr.values = append(tr.values, base)

	lastIndex := len(path.Segments) - 1

	for i := 1; i < len(path.Segments); i++ {
		if stopBeforeLast && i == lastIndex {
			break
		}
		seg, err := eval.materializeSegment(resolved[i])
		if err != nil {
			return nil, err
		}
		resolved[i] = seg
		current := tr.values[len(tr.values)-1]

		if current.GetType() == value.TypeNone {
			return nil, verror.NewScriptError(verror.ErrIDNonePath, [3]string{"cannot traverse through none", "", ""})
		}

		switch seg.Type {
		case value.PathSegmentWord:
			if err := eval.traverseWordSegment(tr, seg, current, !stopBeforeLast); err != nil {
				return nil, err
			}

		case value.PathSegmentIndex:
			if err := eval.traverseIndexSegment(tr, seg, current, !stopBeforeLast); err != nil {
				return nil, err
			}

		default:
			return nil, verror.NewScriptError(
				verror.ErrIDInvalidPath,
				[3]string{
					fmt.Sprintf("unsupported path segment type: %v", seg.Type),
					"invalid-segment",
					"",
				},
			)
		}
	}

	return tr, nil
}

func checkIndexBounds(index, length int64, typeName string) error {
	if index < 1 || index > length {
		return verror.NewScriptError(verror.ErrIDOutOfBounds,
			[3]string{fmt.Sprintf("%d", index), fmt.Sprintf("%d", length), ""})
	}
	return nil
}

func (e *Evaluator) assignToPathTarget(tr *pathTraversal, newVal core.Value, pathStr string) (core.Value, error) {
	if len(tr.segments) < 2 {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidPath,
			[3]string{"set-path requires at least 2 segments", "set-path-too-short", ""},
		)
	}

	if len(tr.values) == 0 {
		return value.NewNoneVal(), verror.NewInternalError("path traversal returned no values", [3]string{})
	}

	if tr.segments[0].Type == value.PathSegmentIndex {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDImmutableTarget,
			[3]string{pathStr, "numeric-literal-base", ""},
		)
	}

	container := tr.values[len(tr.values)-1]
	if container.GetType() == value.TypeNone {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDNonePath,
			[3]string{"cannot assign to none value", pathStr, ""},
		)
	}

	finalSeg := tr.segments[len(tr.segments)-1]
	seg := finalSeg
	if finalSeg.Type == value.PathSegmentEval {
		var err error
		seg, err = e.materializeSegment(finalSeg)
		if err != nil {
			return value.NewNoneVal(), err
		}
		tr.segments[len(tr.segments)-1] = seg
	}
	switch seg.Type {
	case value.PathSegmentIndex:
		return e.assignToIndexTarget(container, seg, newVal, pathStr)
	case value.PathSegmentWord:
		return e.assignToWordTarget(container, seg, newVal, pathStr)
	default:
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDInvalidPath,
			[3]string{
				fmt.Sprintf("unsupported segment type for assignment: %v", seg.Type),
				"invalid-assignment-target",
				"",
			},
		)
	}
}

func (e *Evaluator) assignToIndexTarget(container core.Value, finalSeg value.PathSegment, newVal core.Value, pathStr string) (core.Value, error) {
	index, ok := finalSeg.AsIndex()
	if !ok {
		return value.NewNoneVal(), verror.NewInternalError("index segment does not contain int64", [3]string{})
	}

	switch container.GetType() {
	case value.TypeBlock:
		block, ok := value.AsBlockValue(container)
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError("failed to cast block value", [3]string{})
		}

		if err := checkIndexBounds(index, int64(len(block.Elements)), "block"); err != nil {
			return value.NewNoneVal(), err
		}
		block.Elements[index-1] = newVal

		return newVal, nil

	case value.TypeBinary:
		if newVal.GetType() != value.TypeInteger {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch,
				[3]string{"binary index assignment", "integer", value.TypeToString(newVal.GetType())})
		}

		intVal, ok := value.AsIntValue(newVal)
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError("failed to cast integer value", [3]string{})
		}

		if intVal < 0 || intVal > 255 {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds,
				[3]string{fmt.Sprintf("%d", intVal), "0-255", "binary byte value"})
		}

		bin, ok := value.AsBinaryValue(container)
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError("failed to cast binary value", [3]string{})
		}

		if err := checkIndexBounds(index, int64(bin.Length()), "binary"); err != nil {
			return value.NewNoneVal(), err
		}
		bin.Bytes()[index-1] = byte(intVal)

		return newVal, nil

	default:
		return value.NewNoneVal(), makePathTypeError("index assignment requires block or binary type", value.TypeToString(container.GetType()), pathStr)
	}
}

func (e *Evaluator) assignToWordTarget(container core.Value, finalSeg value.PathSegment, newVal core.Value, pathStr string) (core.Value, error) {
	fieldName, ok := finalSeg.AsWord()
	if !ok {
		return value.NewNoneVal(), verror.NewInternalError("word segment does not contain string", [3]string{})
	}

	if container.GetType() != value.TypeObject {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDImmutableTarget,
			[3]string{
				fmt.Sprintf("cannot assign field to %s (must be object)", value.TypeToString(container.GetType())),
				pathStr,
				"",
			},
		)
	}

	obj, ok := value.AsObject(container)
	if !ok {
		return value.NewNoneVal(), verror.NewInternalError("failed to cast object value", [3]string{})
	}

	obj.SetField(fieldName, newVal)
	return newVal, nil
}

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

func (e *Evaluator) NewReturnSignal(val core.Value) error {
	return NewReturnSignal(val)
}
