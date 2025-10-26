# Plan 009: Alternative Look-Ahead Fix for Reduce Function Infix Bug

STATUS: Rejected (Alternative to Plan 008 - too complex and fragile)

## Overview

Alternative implementation of the `reduce` function fix using a look-ahead approach instead of modifying the core evaluator interface. This approach inspects the next element after evaluation to determine if the current result will be consumed by an infix operation.

## Root Cause (Same as Plan 008)

The `reduce` function incorrectly includes values that are consumed by infix operations in the result block. For `reduce [1 + 2 3 * 4]`, it produces `[1 3 3 12]` instead of `[3 12]`.

## Look-Ahead Solution

**Core Concept:** After evaluating an expression, peek at the next element in the block. If it's an infix function, the current result will be consumed and should not be added to `reducedElements`.

**Implementation in Reduce function:**

```go
for position < len(vals) {
    newPos, result, err := eval.EvaluateExpression(vals, position, lastResult)
    if err != nil {
        return value.NoneVal(), err
    }
    
    // Look ahead: check if this result will be consumed by next infix operation
    shouldAdd := true
    if newPos < len(vals) {
        nextElement := vals[newPos]
        if nextElement.GetType() == value.TypeWord {
            if word, ok := value.AsWord(nextElement); ok {
                if resolved, found := eval.Lookup(word); found {
                    if fn, ok := value.AsFunction(resolved); ok && fn.Infix {
                        shouldAdd = false  // Next element is infix - this result will be consumed
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

## Detailed Logic

### Step-by-Step Evaluation for `[1 + 2 3 * 4]`:

1. **Position 0:** `1` → result = `1`
   - Look ahead at position 1: `+` (word)
   - Resolve `+` → infix function found
   - `shouldAdd = false` → **don't add `1`**
   - `reducedElements = []`

2. **Position 1:** `+` (infix) consumes `lastResult=1` and `2` → result = `3`
   - Look ahead at position 3: `3` (integer, not word)
   - `shouldAdd = true` → **add `3`**
   - `reducedElements = [3]`

3. **Position 3:** `3` → result = `3`
   - Look ahead at position 4: `*` (word)
   - Resolve `*` → infix function found
   - `shouldAdd = false` → **don't add `3`**
   - `reducedElements = [3]`

4. **Position 4:** `*` (infix) consumes `lastResult=3` and `4` → result = `12`
   - At end of block
   - `shouldAdd = true` → **add `12`**
   - `reducedElements = [3, 12]` ✅

## Edge Cases and Considerations

### Complex Expressions
- `reduce [1 + (2 * 3)]` → `[7]`
  - `1` consumed by `+`
  - `(2 * 3)` evaluates to `6`, consumed by `+`
  - Result `7` added

- `reduce [1 2 + 3]` → `[1, 5]`
  - `1` not consumed (no infix after)
  - `2` consumed by `+`
  - Result `5` added

### Non-Word Infix Triggers
- Parentheses: `reduce [1 + (2)]` → `[3]`
  - `1` consumed by `+`
  - `(2)` evaluates to `2`, consumed by `+`

- Literals: `reduce [1 + 2]` → `[3]`
  - `1` consumed by `+`

### Error Conditions
- Undefined words: `reduce [1 + undefined]` → error (lookup fails)
- Invalid infix usage: `reduce [+ 2]` → error (no left operand)

## Advantages

- **No core interface changes** - evaluator remains unchanged
- **Localized fix** - only affects `Reduce` function
- **Minimal risk** - doesn't touch evaluation engine
- **Easier rollback** - can be reverted without broader impact

## Disadvantages

- **Complex logic** - look-ahead creates coupling between evaluation and result collection
- **Fragile assumptions** - relies on evaluation order and position tracking
- **Potential edge cases** - may miss complex scenarios with nested blocks or paths
- **Maintenance burden** - logic must be kept in sync with evaluation model

## Implementation Steps

1. **Analyze current Reduce function** - understand position tracking
2. **Implement look-ahead logic** - add helper function to check next element
3. **Add comprehensive tests** - cover all infix scenarios
4. **Test edge cases** - parentheses, nested blocks, error conditions
5. **Performance validation** - ensure no significant overhead
6. **Regression testing** - verify no impact on non-infix cases

## Testing Requirements

**Core scenarios:**
- `reduce [1 + 2 3 * 4]` → `[3 12]`
- `reduce [1 2 + 3]` → `[1 5]`
- `reduce [1 + 2 * 3]` → `[9]`
- `reduce [1 + (2 * 3)]` → `[7]`

**Edge cases:**
- Empty blocks, single elements
- Mixed infix/prefix functions
- Nested blocks and parentheses
- Error conditions (undefined words, type mismatches)

## Risk Assessment

**Medium Risk:** Look-ahead logic is complex but contained to one function.

**Mitigation:**
- Extensive unit testing before/after
- Clear documentation of look-ahead rules
- Fallback to current behavior if look-ahead fails
- Performance benchmarking to ensure no degradation

## Comparison to Plan 008

| Aspect | Plan 008 (Interface Change) | Plan 009 (Look-Ahead) |
|--------|-----------------------------|----------------------|
| **Complexity** | High (core interface change) | Medium (local logic) |
| **Risk** | High (affects all evaluation) | Medium (contained to reduce) |
| **Accuracy** | High (direct consumption tracking) | Medium (heuristic-based) |
| **Maintenance** | Low (clean separation) | Medium (complex logic) |
| **Testing** | Extensive (interface change) | Focused (reduce function) |
| **Performance** | Minimal impact | Small look-ahead overhead |

**Recommendation:** Prefer Plan 008 for accuracy and maintainability, use Plan 009 as fallback if interface changes prove too risky.
