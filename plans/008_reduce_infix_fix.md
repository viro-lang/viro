# Plan 008: Fix Reduce Function Infix Evaluation Bug

STATUS: Implemented

## Overview

The `reduce` native function in `internal/native/control.go` incorrectly handles infix function evaluation. When evaluating expressions like `reduce [1 + 2 3 * 4]`, it produces `[1 3 3 12]` instead of the expected `[3 12]`. The issue is that values consumed by infix operations are still being added to the result block.

## Root Cause Analysis

The problem occurs in the `Reduce` function's evaluation loop:

```go
for position < len(vals) {
    newPos, result, err := eval.EvaluateExpression(vals, position, lastResult)
    if err != nil {
        return value.NoneVal(), err
    }
    reducedElements = append(reducedElements, result)  // ❌ Always adds result
    position = newPos
    lastResult = result
}
```

When evaluating `[1 + 2 3 * 4]`:
1. `1` evaluates to `1` → added to `reducedElements`: `[1]`
2. `+` (infix) consumes `lastResult=1` and `2` → produces `3` → added: `[1, 3]`  
3. `3` evaluates to `3` → added: `[1, 3, 3]`
4. `*` (infix) consumes `lastResult=3` and `4` → produces `12` → added: `[1, 3, 3, 12]`

But `1` and the second `3` should not be added since they were consumed by infix operations.

## Proposed Solutions

### Solution 1: Modify EvaluateExpression to Track Infix Consumption (SELECTED)

**Change the evaluator interface to return whether `lastResult` was consumed:**

```go
// Current signature
func (e *Evaluator) EvaluateExpression(block []core.Value, position int, lastResult core.Value) (int, core.Value, error)

// Proposed signature  
func (e *Evaluator) EvaluateExpression(block []core.Value, position int, lastResult core.Value) (int, core.Value, bool, error)
//                                                                                               ^^^^^
//                                                                                               true if lastResult was consumed as infix operand
```

**Update Reduce function:**
```go
for position < len(vals) {
    newPos, result, consumedLast, err := eval.EvaluateExpression(vals, position, lastResult)
    if err != nil {
        return value.NoneVal(), err
    }
    if !consumedLast {  // Only add result if lastResult wasn't consumed
        reducedElements = append(reducedElements, result)
    }
    position = newPos
    lastResult = result
}
```

**Pros:**
- Clean separation of concerns
- Minimal changes to Reduce logic
- Accurate tracking of infix consumption

**Cons:**
- Requires changing the core evaluator interface
- Affects all callers of EvaluateExpression

### Solution 2: Look-Ahead Detection (REJECTED - too complex and fragile)

**After evaluation, peek ahead to see if the result would be consumed by an infix function:**

```go
for position < len(vals) {
    newPos, result, err := eval.EvaluateExpression(vals, position, lastResult)
    if err != nil {
        return value.NoneVal(), err
    }
    
    // Check if this result would be consumed by next infix operation
    shouldAdd := true
    if newPos < len(vals) {
        nextElement := vals[newPos]
        if nextElement.GetType() == value.TypeWord {
            if word, ok := value.AsWord(nextElement); ok {
                if resolved, found := eval.Lookup(word); found {
                    if fn, ok := resolved.(*value.FunctionValue); ok && fn.Infix {
                        shouldAdd = false  // Next element is infix, so this result will be consumed
                    }
                }
            }
        }
    }
    
    if shouldAdd {
        reducedElements = append(reducedElements, result)
    }
    position = newPos
    lastResult = result
}
```

**Pros:**
- No changes to evaluator interface
- Localized to Reduce function

**Cons:**
- Complex look-ahead logic
- Fragile (assumes evaluation order)
- May not handle all edge cases

### Solution 3: Post-Processing Filter (REJECTED - too complex and error-prone)

**Evaluate all expressions, then filter out consumed values:**

This would require tracking which positions produced values that were later consumed, which is complex and error-prone.

**Selected: Solution 1** - Modify the evaluator interface to properly track infix consumption. This provides the most accurate and maintainable solution.

## Implementation Steps

1. **Modify EvaluateExpression signature** in `internal/core/core.go` and `internal/eval/evaluator.go`
2. **Update collectFunctionArgs** to return whether `lastResult` was consumed
3. **Update all EvaluateExpression callers** to handle the new return value
4. **Update Reduce function** to conditionally add results
5. **Add comprehensive tests** for infix scenarios in reduce
6. **Verify no regressions** in existing functionality

## Testing Requirements

- `reduce [1 + 2 3 * 4]` → `[3 12]`
- `reduce [1 2 + 3]` → `[1 5]` (infix consumes previous result)
- `reduce [1 + 2 * 3]` → `[9]` (chained infix)
- `reduce [1 + (2 * 3)]` → `[7]` (parentheses prevent infix consumption)
- Mixed infix and prefix scenarios
- Edge cases with none values and error conditions

## Risk Assessment

**High Risk:** Changes to core evaluator interface could break existing functionality.

**Mitigation:**
- Comprehensive test coverage before/after changes
- Incremental implementation with frequent testing
- Focus on backward compatibility for non-infix cases
