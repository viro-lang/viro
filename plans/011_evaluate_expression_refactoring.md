# Plan 011: Refactor EvaluateExpression to Handle Infix Lookahead

STATUS: Proposed (Follow-up to Plan 010)

## Overview

Refactor `EvaluateExpression` to handle infix operator lookahead internally, eliminating the need for `lastResult` parameter and `consumedLast` return value. This will simplify the evaluation model and fix infix consumption issues universally.

## Current Problem

The current evaluation model splits infix handling responsibility between:
1. **EvaluateExpression** - Checks if current function is infix and uses `lastResult`
2. **DoBlock** - Passes `lastResult` between expressions
3. **Set-word** - Has custom lookahead logic (Plan 010 workaround)
4. **Reduce** - Uses `consumedLast` flag to track infix consumption

This creates complexity, duplication, and edge cases.

## Current Signature

```go
func (e *Evaluator) EvaluateExpression(
    block []core.Value, 
    position int, 
    lastResult core.Value,
) (newPos int, result core.Value, consumedLast bool, error)
```

**Problems:**
- `lastResult` couples evaluation to previous result
- `consumedLast` creates backward dependency (current expression signals about past)
- Callers must track and manage infix consumption state
- Duplicate lookahead logic in set-word handler

## Proposed Signature

```go
func (e *Evaluator) EvaluateExpression(
    block []core.Value, 
    position int,
) (newPos int, result core.Value, error)
```

**Benefits:**
- Simpler signature (3 params → 2 params)
- Self-contained evaluation (no external state needed)
- Lookahead handled internally and consistently
- No backward signaling needed

## Implementation Strategy

### Step 1: Core Lookahead Logic

```go
func (e *Evaluator) EvaluateExpression(block []core.Value, position int) (int, core.Value, error) {
    if position >= len(block) {
        return position, value.NoneVal(), verror.NewScriptError(...)
    }

    // 1. Evaluate current element (literal, word, set-word, paren, etc.)
    newPos, result, err := e.evaluateElement(block, position)
    if err != nil {
        return position, value.NoneVal(), err
    }

    // 2. Look ahead for infix operators and consume them
    for newPos < len(block) {
        if !e.isNextInfixOperator(block, newPos) {
            break
        }
        
        // Consume infix operator with current result as left operand
        nextPos, nextResult, err := e.consumeInfixOperator(block, newPos, result)
        if err != nil {
            return position, value.NoneVal(), err
        }
        
        newPos = nextPos
        result = nextResult
    }

    return newPos, result, nil
}
```

### Step 2: Helper Functions

```go
// evaluateElement evaluates a single element without lookahead
func (e *Evaluator) evaluateElement(block []core.Value, position int) (int, core.Value, error) {
    element := block[position]
    
    switch element.GetType() {
    case value.TypeInteger, value.TypeString, ...:
        return position + 1, element, nil
        
    case value.TypeSetWord:
        return e.evaluateSetWord(block, position)
        
    case value.TypeWord:
        return e.evaluateWord(block, position)
        
    // ... other cases
    }
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
    
    word, ok := value.AsWord(nextElement)
    if !ok {
        return false
    }
    
    resolved, found := e.Lookup(word)
    if !found {
        return false
    }
    
    fn, ok := value.AsFunction(resolved)
    return ok && fn.Infix
}

// consumeInfixOperator evaluates an infix operator with given left operand
func (e *Evaluator) consumeInfixOperator(block []core.Value, position int, leftOperand core.Value) (int, core.Value, error) {
    wordElement := block[position]
    word, _ := value.AsWord(wordElement)
    resolved, _ := e.Lookup(word)
    fn, _ := value.AsFunction(resolved)
    
    // Invoke function with leftOperand as lastResult (for infix consumption)
    return e.invokeFunctionExpression(block, position, fn, leftOperand)
}
```

### Step 3: Update Set-Word Handler

Remove custom lookahead logic from set-word since it's now handled by EvaluateExpression:

```go
case value.TypeSetWord:
    wordStr, _ := value.AsWord(element)
    
    if strings.Contains(wordStr, ".") {
        return e.evalSetPathExpression(block, position, wordStr)
    }
    
    if position+1 >= len(block) {
        return position, value.NoneVal(), verror.NewScriptError(...)
    }
    
    // Just evaluate next expression - lookahead handled by EvaluateExpression
    newPos, result, err := e.EvaluateExpression(block, position+1)
    if err != nil {
        return position, value.NoneVal(), err
    }
    
    // Bind and return
    currentFrame := e.currentFrame()
    if currentFrame != nil {
        currentFrame.Bind(wordStr, result)
    }
    return newPos, result, nil
```

### Step 4: Update DoBlock

Simplify DoBlock loop (no lastResult tracking needed):

```go
func (e *Evaluator) DoBlock(vals []core.Value) (core.Value, error) {
    if len(vals) == 0 {
        return value.NoneVal(), nil
    }

    position := 0
    lastResult := value.NoneVal()

    for position < len(vals) {
        newPos, result, err := e.EvaluateExpression(vals, position)
        if err != nil {
            return value.NoneVal(), e.annotateError(err, vals, position)
        }
        position = newPos
        lastResult = result
    }

    return lastResult, nil
}
```

Note: `lastResult` is kept only for returning the final result, not passed to EvaluateExpression.

### Step 5: Update Reduce Function

Simplify reduce (no consumedLast tracking needed):

```go
func Reduce(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    // ... validation ...
    
    vals := block.Elements
    reducedElements := make([]core.Value, 0)

    position := 0
    for position < len(vals) {
        newPos, result, err := eval.EvaluateExpression(vals, position)
        if err != nil {
            return value.NoneVal(), err
        }
        
        // Simply add each result - lookahead already consumed infix chains
        reducedElements = append(reducedElements, result)
        position = newPos
    }

    return value.BlockVal(reducedElements), nil
}
```

### Step 6: Update Function Invocation

The tricky part - `invokeFunctionExpression` currently returns `consumedLast`. This needs refactoring:

**Current:**
```go
func (e *Evaluator) invokeFunctionExpression(block []core.Value, position int, fn *value.FunctionValue, lastResult core.Value) (int, core.Value, bool, error)
```

**After:**
```go
func (e *Evaluator) invokeFunctionExpression(block []core.Value, position int, fn *value.FunctionValue, leftOperand core.Value) (int, core.Value, error)
```

The `leftOperand` parameter is only used when this function is called from `consumeInfixOperator`. Regular calls pass `value.NoneVal()`.

## Migration Plan

### Phase 1: Add New Method (Non-Breaking)
1. Create `EvaluateExpressionV2(block, position) (int, Value, error)`
2. Implement lookahead logic
3. Test extensively with existing test suite

### Phase 2: Migrate Callers
1. Update DoBlock to use V2
2. Update set-word handler to use V2
3. Update reduce to use V2
4. Update all other callers

### Phase 3: Remove Old Method
1. Delete `EvaluateExpression` (old signature)
2. Rename `EvaluateExpressionV2` → `EvaluateExpression`
3. Clean up `consumedLast` tracking code

## Test Cases

### Basic Infix Consumption
```viro
1 + 2          → 3          (lookahead consumes +)
1 + 2 * 3      → 9          (lookahead consumes + and *)
x: 1 + 2       → x = 3      (set-word gets full result)
```

### Reduce Function
```viro
reduce [1 + 2 3 * 4]  → [3 12]   (no more consumedLast tracking)
reduce [1 2 + 3]      → [1 5]    (correct behavior)
```

### Function Arguments
```viro
if 1 < 2 ["yes"] ["no"]  → "yes"  (should fix this currently broken case)
```

### Set-Word Chains
```viro
x: 10
y: x + 5        → y = 15
y: y * 2        → y = 30
```

### No Unintended Consumption
```viro
sq: fn [n] [n * n]  sq 5  → sq defined, then called with 5
x: 10  y: 20              → both assigned correctly
```

## Risks

### High Risk Areas
1. **Function argument collection** - May need careful adjustment
2. **Infix operator invocation** - Core to the change
3. **Nested expressions** - Parentheses, blocks, etc.

### Mitigation
1. Comprehensive unit tests before refactoring
2. Phased migration (V2 approach)
3. Test each phase independently
4. Keep old implementation until V2 fully validated

## Success Criteria

1. ✅ All existing tests pass
2. ✅ `TestREPL_StatePreservedAfterError` passes (currently fixed by Plan 010)
3. ✅ `if 1 < 2 [...]` works correctly (currently broken)
4. ✅ Reduce function works without consumedLast
5. ✅ Set-word simplified (no custom lookahead)
6. ✅ Cleaner codebase (less complexity)

## Related Plans
- Plan 008: DoNext method refactoring (similar complexity)
- Plan 009: Reduce lookahead fix (would be obsoleted by this)
- Plan 010: Set-word infix consumption (temporary workaround to be replaced)

## Estimated Effort

- **Investigation**: 2-4 hours (understand all EvaluateExpression call sites)
- **Implementation**: 4-8 hours (core logic + helper functions)
- **Testing**: 4-6 hours (comprehensive test coverage)
- **Migration**: 2-4 hours (update all callers)
- **Validation**: 2-4 hours (regression testing)

**Total**: 14-26 hours over 2-3 days

## Next Steps

1. Review this plan with team/maintainer
2. Get approval for API-breaking change
3. Create feature branch
4. Implement Phase 1 (V2 method)
5. Validate with existing tests
6. Proceed with Phases 2-3 if successful
