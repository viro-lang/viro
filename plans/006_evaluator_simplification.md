# Plan 006: Evaluator Simplification

## Overview

Refactor the evaluator implementation in `internal/eval/evaluator.go` to follow the correct execution model as documented in `docs/execution-model.md`. The current implementation deviates from the fundamental position-tracking pattern required by Viro's sequential evaluation model.

## Critical Issue Identified

**The current evaluator violates the execution model's core principle:**

> "evaluate_expression() returns (position, value)"

**Current implementation (WRONG):**
```go
func (e *Evaluator) DoNext(val core.Value) (core.Value, error)  // ❌ No position returned
func (e *Evaluator) DoBlock(vals []core.Value) (core.Value, error) {
    for i := 0; i < len(vals); i++ {  // ❌ Traditional for-loop
        lastResult, err = e.evaluateWithFunctionCall(val, vals, &i, lastResult)  // ❌ Mutates &i
    }
}
```

**Correct pattern (per execution-model.md):**
```go
func (e *Evaluator) evaluateExpression(block []core.Value, position int, lastResult core.Value) (int, core.Value, error)
func (e *Evaluator) evaluateBlock(block []core.Value) (core.Value, error) {
    position := 0
    lastResult := value.NoneVal()
    
    for position < len(block) {
        newPos, result, err := e.evaluateExpression(block, position, lastResult)
        if err != nil {
            return value.NoneVal(), err
        }
        position = newPos
        lastResult = result
    }
    
    return lastResult, nil
}
```

**Why this matters:**
- Function calls consume multiple positions (function word + arguments)
- Set-words consume multiple positions (set-word + expression)
- Position tracking enables proper sequential consumption
- Index mutation via pointer is an anti-pattern that obscures control flow

## Goals

- Refactor evaluator to follow position-returning pattern
- Reduce evaluator.go complexity while fixing architectural issues
- Maintain 100% test coverage
- Keep all existing functionality working

## Current State Analysis

### File: internal/eval/evaluator.go (1,114 lines)

**Architectural problems:**
1. **DoNext doesn't return position** - violates execution model (lines 389-421)
2. **DoBlock uses traditional for-loop** - should use position tracking (lines 426-454)
3. **Index mutation via &i pointer** - passed to evalSetWord, evaluateWithFunctionCall
4. **Multiple evaluation entry points** - DoNext, evaluateWithFunctionCall, evalExpressionFromTokens doing similar things
5. **Dispatch map instead of switch** - unnecessary indirection (lines 138-161)

**Complexity breakdown:**
- Core evaluation dispatch: ~420 lines (DoNext, DoBlock, dispatch functions)
- Function invocation: ~230 lines (invokeFunction, collectFunctionArgsWithInfix)
- Path evaluation: ~180 lines (traversePath, evalPath, evalSetPath)
- Frame management: ~150 lines (push/pop/lookup/capture)
- Helper functions: ~134 lines (error annotation, call stack, etc.)

## Refactoring Strategy

### Phase 1: Fix Core Execution Model (CRITICAL)

This is the foundational fix that enables all other simplifications.

#### 1.1 Create New `evaluateExpression` Method

**Signature:**
```go
func (e *Evaluator) evaluateExpression(
    block []core.Value,
    position int,
    lastResult core.Value,
) (newPosition int, result core.Value, err error)
```

**Implementation pattern (per execution-model.md lines 56-91):**
```go
func (e *Evaluator) evaluateExpression(block []core.Value, position int, lastResult core.Value) (int, core.Value, error) {
    if position >= len(block) {
        return position, value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"missing expression", "", ""})
    }
    
    element := block[position]
    
    // Add tracing if enabled
    var traceStart time.Time
    var traceWord string
    if trace.GlobalTraceSession != nil && trace.GlobalTraceSession.IsEnabled() {
        traceStart = time.Now()
        if value.IsWord(element.GetType()) {
            if w, ok := value.AsWord(element); ok {
                traceWord = w
            }
        }
    }
    
    switch element.GetType() {
    // Literals evaluate to themselves, advance position by 1
    case value.TypeInteger, value.TypeString, value.TypeLogic, 
         value.TypeNone, value.TypeDecimal, value.TypeObject,
         value.TypePort, value.TypeDatatype, value.TypeBlock, 
         value.TypeFunction:
        return position + 1, element, nil
    
    // Paren: evaluate block immediately
    case value.TypeParen:
        block, _ := value.AsBlock(element)
        result, err := e.evaluateBlock(block.Elements)
        return position + 1, result, err
    
    // LitWord: return word value
    case value.TypeLitWord:
        return position + 1, value.WordVal(element.GetPayload().(string)), nil
    
    // GetWord: lookup without invocation
    case value.TypeGetWord:
        wordStr, _ := value.AsWord(element)
        result, ok := e.Lookup(wordStr)
        if !ok {
            return position, value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
        }
        return position + 1, result, nil
    
    // SetWord: evaluate next expression and bind
    case value.TypeSetWord:
        wordStr, _ := value.AsWord(element)
        
        // Handle path-based assignment
        if strings.Contains(wordStr, ".") {
            return e.evalSetPathExpression(block, position, wordStr)
        }
        
        // Ensure we have a frame to bind to
        currentFrame := e.currentFrame()
        if currentFrame == nil {
            currentFrame = frame.NewFrame(frame.FrameFunctionArgs, -1)
            e.pushFrame(currentFrame)
        }
        
        // Evaluate next expression
        newPos, result, err := e.evaluateExpression(block, position+1, lastResult)
        if err != nil {
            return position, value.NoneVal(), e.annotateError(err, block, position)
        }
        
        // Auto-name anonymous functions
        if result.GetType() == value.TypeFunction {
            if fnVal, ok := value.AsFunction(result); ok && fnVal.Name == "" {
                fnVal.Name = wordStr
            }
        }
        
        currentFrame.Bind(wordStr, result)
        return newPos, result, nil
    
    // Word: lookup and potentially invoke
    case value.TypeWord:
        wordStr, _ := value.AsWord(element)
        
        // Check for breakpoint (debugging)
        if debug.GlobalDebugger != nil && debug.GlobalDebugger.HasBreakpoint(wordStr) {
            if trace.GlobalTraceSession != nil && trace.GlobalTraceSession.IsEnabled() {
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
            return position, value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
        }
        
        // If it's a function, invoke it
        if resolved.GetType() == value.TypeFunction {
            fn, _ := value.AsFunction(resolved)
            newPos, result, err := e.invokeFunctionExpression(block, position, fn, lastResult)
            
            // Emit trace
            if trace.GlobalTraceSession != nil && trace.GlobalTraceSession.IsEnabled() && traceWord != "" {
                duration := time.Since(traceStart)
                trace.GlobalTraceSession.Emit(trace.TraceEvent{
                    Timestamp: traceStart,
                    Value:     result.Form(),
                    Word:      traceWord,
                    Duration:  duration.Nanoseconds(),
                })
            }
            
            return newPos, result, err
        }
        
        // Not a function, just return the value
        return position + 1, resolved, nil
    
    // Path: evaluate path expression
    case value.TypePath:
        path, _ := value.AsPath(element)
        result, err := e.evalPathValue(path)
        return position + 1, result, err
    
    default:
        return position, value.NoneVal(), verror.NewInternalError("unknown value type in evaluateExpression", [3]string{})
    }
}
```

#### 1.2 Create New `evaluateBlock` Method

**Signature:**
```go
func (e *Evaluator) evaluateBlock(block []core.Value) (core.Value, error)
```

**Implementation (per execution-model.md lines 102-111):**
```go
func (e *Evaluator) evaluateBlock(block []core.Value) (core.Value, error) {
    if len(block) == 0 {
        return value.NoneVal(), nil
    }
    
    position := 0
    lastResult := value.NoneVal()
    
    for position < len(block) {
        newPos, result, err := e.evaluateExpression(block, position, lastResult)
        if err != nil {
            return value.NoneVal(), e.annotateError(err, block, position)
        }
        position = newPos
        lastResult = result
    }
    
    return lastResult, nil
}
```

#### 1.3 Update Function Invocation to Return Position

**Current:** `invokeFunction` mutates `*idx` pointer
**Target:** Return new position

```go
func (e *Evaluator) invokeFunctionExpression(
    block []core.Value,
    position int,  // Position of function word
    fn *value.FunctionValue,
    lastResult core.Value,
) (newPosition int, result core.Value, err error) {
    
    name := functionDisplayName(fn)
    e.pushCall(name)
    defer e.popCall()
    
    // Collect arguments starting from position+1
    posArgs, refValues, newPos, err := e.collectFunctionArgs(fn, block, position+1, lastResult)
    if err != nil {
        return position, value.NoneVal(), e.annotateError(err, block, position)
    }
    
    // Invoke function
    if fn.Type == value.FuncNative {
        result, err := e.callNative(fn, posArgs, refValues)
        if err != nil {
            return position, value.NoneVal(), e.annotateError(err, block, position)
        }
        return newPos, result, nil
    }
    
    result, err := e.executeFunction(fn, posArgs, refValues)
    if err != nil {
        return position, value.NoneVal(), err
    }
    return newPos, result, nil
}
```

#### 1.4 Update Argument Collection to Return Position

**Current:** `collectFunctionArgsWithInfix` returns `consumed` count
**Target:** Return actual new position

```go
func (e *Evaluator) collectFunctionArgs(
    fn *value.FunctionValue,
    block []core.Value,
    startPosition int,  // Position after function word
    lastResult core.Value,
) (posArgs []core.Value, refValues map[string]core.Value, newPosition int, err error) {
    
    // Separate positional and refinement parameters
    positional := make([]value.ParamSpec, 0, len(fn.Params))
    refSpecs := make(map[string]value.ParamSpec)
    refValues = make(map[string]core.Value)
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
    
    posArgs = make([]core.Value, len(positional))
    position := startPosition
    paramIndex := 0
    
    // Handle infix
    useInfix := fn.Infix && lastResult.GetType() != value.TypeNone
    if useInfix {
        if len(positional) == 0 {
            return nil, nil, position, verror.NewScriptError(
                verror.ErrIDArgCount,
                [3]string{functionDisplayName(fn), "0", "1 (infix requires at least one parameter)"},
            )
        }
        posArgs[0] = lastResult
        paramIndex = 1
    }
    
    // Collect positional arguments
    for paramIndex < len(positional) {
        paramSpec := positional[paramIndex]
        
        // Read refinements before this argument
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
        
        // Evaluate or use literal
        var arg core.Value
        if paramSpec.Eval {
            var newPos int
            newPos, arg, err = e.evaluateExpression(block, position, value.NoneVal())
            if err != nil {
                return nil, nil, position, err
            }
            position = newPos
        } else {
            arg = block[position]
            position++
        }
        
        posArgs[paramIndex] = arg
        paramIndex++
    }
    
    // Read trailing refinements
    position, err = e.readRefinements(block, position, refSpecs, refValues, refProvided)
    if err != nil {
        return nil, nil, position, err
    }
    
    return posArgs, refValues, position, nil
}
```

#### 1.5 Update `readRefinements` to Return Position

**Current:** Returns `int` for new position
**Target:** Keep same signature but ensure it's used correctly

```go
func (e *Evaluator) readRefinements(
    block []core.Value,
    position int,
    refSpecs map[string]value.ParamSpec,
    refValues map[string]core.Value,
    refProvided map[string]bool,
) (newPosition int, err error) {
    
    for position < len(block) && isRefinement(block[position]) {
        wordStr, _ := value.AsWord(block[position])
        refName := strings.TrimPrefix(wordStr, "--")
        
        spec, exists := refSpecs[refName]
        if !exists {
            return position, verror.NewScriptError(
                verror.ErrIDInvalidOperation,
                [3]string{fmt.Sprintf("Unknown refinement: --%s", refName), "", ""},
            )
        }
        
        if refProvided[refName] {
            return position, verror.NewScriptError(
                verror.ErrIDInvalidOperation,
                [3]string{fmt.Sprintf("Duplicate refinement: --%s", refName), "", ""},
            )
        }
        
        if spec.TakesValue {
            if position+1 >= len(block) {
                return position, verror.NewScriptError(
                    verror.ErrIDInvalidOperation,
                    [3]string{fmt.Sprintf("Refinement --%s requires a value", refName), "", ""},
                )
            }
            var arg core.Value
            position, arg, err = e.evaluateExpression(block, position+1, value.NoneVal())
            if err != nil {
                return position, err
            }
            refValues[refName] = arg
        } else {
            refValues[refName] = value.LogicVal(true)
            position++
        }
        
        refProvided[refName] = true
    }
    
    return position, nil
}
```

### Phase 2: Maintain Public API Compatibility

The public API (DoNext, DoBlock, EvalExpressionFromTokens) must remain for backward compatibility with native functions.

#### 2.1 Update DoNext as Wrapper

```go
func (e *Evaluator) DoNext(val core.Value) (core.Value, error) {
    _, result, err := e.evaluateExpression([]core.Value{val}, 0, value.NoneVal())
    return result, err
}
```

#### 2.2 Update DoBlock as Wrapper

```go
func (e *Evaluator) DoBlock(vals []core.Value) (core.Value, error) {
    return e.evaluateBlock(vals)
}
```

#### 2.3 Update EvalExpressionFromTokens

```go
func (e *Evaluator) EvalExpressionFromTokens(tokens []core.Value, startPos int) (core.Value, int, error) {
    newPos, result, err := e.evaluateExpression(tokens, startPos, value.NoneVal())
    return result, newPos, err
}
```

### Phase 3: Remove Obsolete Code

After Phase 1 & 2 are working:

1. Delete `evalDispatch` map (lines 138-161)
2. Delete dispatch functions (lines 163-234):
   - evalLiteral
   - evalBlock  
   - evalFunction
   - evalParenDispatch
   - evalWordDispatch
   - evalSetWordDispatch
   - evalGetWordDispatch
   - evalLitWordDispatch
   - evalPathDispatch
3. Delete `evaluateWithFunctionCall` (lines 549-570)
4. Delete `evalExpressionFromTokens` internal (keep public wrapper)
5. Rename internal helpers:
   - `evalWord` → `evalWordValue` (just lookup logic)
   - `evalGetWord` → `evalGetWordValue`
   - `evalPath` → `evalPathValue`
   - `evalSetPath` → `evalSetPathExpression`

## Implementation Steps

### Phase 1: Core Refactoring

1. **Create new methods (keep old ones)**
   - Add `evaluateExpression(block, position, lastResult) (newPos, result, error)`
   - Add `evaluateBlock(block) (result, error)`
   - Add `invokeFunctionExpression(block, position, fn, lastResult) (newPos, result, error)`
   - Add `collectFunctionArgs(fn, block, startPos, lastResult) (args, refs, newPos, error)`
   - Update `readRefinements` signature to `(block, pos, ...) (newPos, error)`

2. **Test new implementation in parallel**
   - Create test file `internal/eval/evaluator_v2_test.go`
   - Port key tests to use new methods directly
   - Verify position tracking is correct

3. **Update public API methods to use new internals**
   - Modify `DoNext` to call `evaluateExpression`
   - Modify `DoBlock` to call `evaluateBlock`
   - Modify `EvalExpressionFromTokens` to call `evaluateExpression`
   - Run full test suite: `make test`

4. **Remove old implementation**
   - Delete dispatch map and init()
   - Delete individual dispatch functions
   - Delete `evaluateWithFunctionCall`
   - Delete old `evalExpressionFromTokens`
   - Delete old `collectFunctionArgsWithInfix`
   - Run full test suite: `make test`

### Phase 2: Path Expression Updates

5. **Update path evaluation to position-based**
   - Modify `evalSetPath` to `evalSetPathExpression(block, position, pathStr) (newPos, result, error)`
   - Ensure path operations return position

### Phase 3: Validation

6. **Full test suite**
   ```bash
   make grammar
   make build
   make test
   go test -coverprofile=coverage.out ./...
   ```

7. **Integration tests**
   ```bash
   go test -v ./test/integration/...
   go test -v ./test/contract/...
   ```

8. **Verify correct execution model**
   - Position tracking is consistent
   - Function calls consume correct number of values
   - Infix operations work correctly
   - Set-words consume next expression
   - Refinements work in all positions

## Expected Outcomes

### Architectural Improvements

✅ **Follows execution-model.md specification**
- evaluateExpression returns (position, value, error)
- evaluateBlock uses position tracking loop
- No index mutation via pointers
- Clear sequential evaluation model

### Code Reduction

| Component | Before | After | Reduction |
|-----------|--------|-------|-----------|
| Dispatch infrastructure | ~100 lines | 0 lines | -100 |
| Core evaluation | ~420 lines | ~250 lines | -170 |
| Position tracking clarity | Low | High | N/A |
| **Total** | **1,114 lines** | **~850 lines** | **~260 lines** |

### Maintainability

- Single evaluation path (evaluateExpression)
- Clear position tracking (no hidden mutations)
- Matches documented execution model
- Easier to understand control flow
- Simpler debugging (position is explicit)

## Risks and Mitigations

### Risk 1: Breaking position tracking
- **Mitigation:** Test position returns at each step
- **Test:** Verify `1 + 2 + 3` returns position 3 (consumed all)
- **Test:** Verify `x: add 5 3` returns correct position

### Risk 2: Performance regression
- **Mitigation:** Benchmark before/after
- **Acceptable:** <5% for cleaner architecture

### Risk 3: Native function compatibility
- **Mitigation:** Public API (DoNext, DoBlock, EvalExpressionFromTokens) unchanged
- **Test:** All native function tests pass

## Testing Strategy

### Position Tracking Tests

```go
// Test: evaluateExpression returns correct position
block := []core.Value{value.IntVal(42)}
newPos, result, err := e.evaluateExpression(block, 0, value.NoneVal())
assert.Equal(t, 1, newPos)  // Consumed 1 value
assert.Equal(t, int64(42), result.GetPayload().(int64))

// Test: set-word consumes expression
block = []core.Value{value.SetWordVal("x"), value.IntVal(10)}
newPos, result, err = e.evaluateExpression(block, 0, value.NoneVal())
assert.Equal(t, 2, newPos)  // Consumed set-word + value
assert.Equal(t, int64(10), result.GetPayload().(int64))

// Test: function call consumes arguments
// add 5 3 → should consume 3 positions
block = []core.Value{value.WordVal("add"), value.IntVal(5), value.IntVal(3)}
// (assuming 'add' is bound to a 2-arg function)
newPos, result, err = e.evaluateExpression(block, 0, value.NoneVal())
assert.Equal(t, 3, newPos)  // Consumed function + 2 args
```

### Infix Tests

```go
// Test: infix uses lastResult
// Block: [1, +, 2]
// Step 1: eval 1 → (pos=1, result=1)
// Step 2: eval + with lastResult=1 → consumes 2, returns (pos=3, result=3)
```

## Success Criteria

- [ ] All tests pass
- [ ] evaluateExpression returns (position, value, error)
- [ ] evaluateBlock uses position-based loop
- [ ] No pointer-based index mutation
- [ ] Code reduced by ~250 lines
- [ ] Execution model matches docs/execution-model.md
- [ ] No performance regression >5%

## Timeline Estimate

- Phase 1 (Core refactoring): 6 hours
- Phase 2 (Path updates): 2 hours
- Phase 3 (Validation): 2 hours

**Total:** ~10 hours

## References

- **Execution model:** `docs/execution-model.md`
- Current implementation: `internal/eval/evaluator.go`
- Lua reference: `https://github.com/marad/viro-lang/blob/main/experiments/test.lua`
- Test suite: `test/contract/*_test.go`, `test/integration/*_test.go`
