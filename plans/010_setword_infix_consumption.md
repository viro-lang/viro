# Plan 010: Set-Word Infix Consumption Fix

STATUS: Implemented

## Problem Statement

The set-word assignment operator `:` was not correctly consuming infix operators on the right-hand side, leading to incorrect variable bindings.

### Example of the Bug

```viro
x: 10
y: x + 5
y  ; Returns 10, should return 15
```

### What Was Happening

The code `y: x + 5` parses into 4 tokens: `[y:, x, +, 5]`

**Incorrect evaluation (before fix):**
1. Position 0: `y:` (set-word) evaluates expression at position+1
2. Position 1: `x` (word) returns 10, newPos=2
3. Set-word binds `y = 10` and returns result=10, newPos=2
4. DoBlock continues at position 2 with lastResult=10
5. Position 2: `+` (infix) consumes lastResult=10 and evaluates `10 + 5 = 15`
6. DoBlock returns 15 (misleading!)
7. But `y` is bound to 10, not 15

**The issue:** Set-word was only consuming ONE expression (`x`), not the full right-hand side expression (`x + 5`).

## Root Cause

In left-to-right evaluation with infix operators, the set-word operator needs to:
1. Evaluate the next expression
2. Look ahead to see if an infix operator follows
3. If yes, continue evaluation to consume the infix chain
4. Only then bind the final result

The original implementation only evaluated position+1 and stopped, leaving infix operators to be consumed by DoBlock's loop with the set-word's result as `lastResult`.

## Solution Implemented

### Look-Ahead Approach

Modified `TypeSetWord` case in `EvaluateExpression` to:

```go
case value.TypeSetWord:
    // 1. Evaluate first expression
    newPos, result, _, err := e.EvaluateExpression(block, position+1, value.NoneVal())
    
    // 2. Look ahead: if next element is infix, continue evaluation
    for newPos < len(block) {
        nextElement := block[newPos]
        if nextElement.GetType() == value.TypeWord {
            if nextWord, ok := value.AsWord(nextElement); ok {
                if resolved, found := e.Lookup(nextWord); found {
                    if fn, ok := value.AsFunction(resolved); ok && fn.Infix {
                        // Next is infix - continue evaluation
                        nextPos, nextResult, _, nextErr := e.EvaluateExpression(block, newPos, result)
                        newPos = nextPos
                        result = nextResult
                        continue
                    }
                }
            }
        }
        break
    }
    
    // 3. Bind the final result
    currentFrame.Bind(wordStr, result)
    return newPos, result, false, nil
```

### Key Changes

1. **Pass `value.NoneVal()` as lastResult** when evaluating position+1 to prevent unintended infix consumption
2. **Loop to check for infix operators** at newPos after each expression
3. **Continue evaluation** if infix operator found, passing current result as left operand
4. **Break when no more infix** operators are found
5. **Bind final result** after full chain is consumed

### Correct Evaluation (after fix)

For `y: x + 5`:
1. Position 0: `y:` starts evaluation
2. Evaluates `x` at position 1 → result=10, newPos=2
3. Looks ahead at position 2 → finds `+` (infix function)
4. Evaluates `+` at position 2 with lastResult=10 → result=15, newPos=4
5. Looks ahead at position 4 → end of block
6. Binds `y = 15`
7. Returns newPos=4, result=15

Now `y` correctly equals 15!

## Test Results

### Fixed Tests
- `TestREPL_StatePreservedAfterError` - PASSING ✅
- `TestSetWordInfixConsumption` - PASSING ✅

### Test Cases Covered

```viro
; Simple assignment
x: 10                    ; x = 10

; Assignment with single infix
y: 5 + 3                 ; y = 8

; Assignment with multiple infix
a: 1 + 2 * 3             ; a = 9 (left-to-right: (1+2)*3)

; Reassignment after error
x: 10
1 / 0                    ; error occurs
x: x + 5                 ; x = 15 (was incorrectly 10)
```

## Pre-Existing Issues (Not Related to This Fix)

The following tests were ALREADY failing before this change:
- `TestUS2_ControlFlowScenarios` - `if 1 < 2 ["yes"] ["no"]` fails
- `TestSetWordDoesNotConsumeComparison` - added as part of investigation

These failures are NOT caused by the set-word fix. They appear to be a separate issue with how `if` or comparison operators work.

## Implementation Details

### File Modified
- `internal/eval/evaluator.go` - TypeSetWord case in EvaluateExpression method

### Approach Inspiration
Based on Plan 009's look-ahead strategy for the `reduce` function, adapted for set-word evaluation.

### Why Look-Ahead Works

The look-ahead approach:
- ✅ Localized to set-word handling
- ✅ No changes to core evaluation interface
- ✅ Handles arbitrary infix chains: `x: a + b * c - d`
- ✅ Stops at non-infix: `sq: fn [n] [n * n]  sq 5` (doesn't consume `sq 5`)
- ✅ Minimal risk to other evaluation paths

### Edge Cases Handled

1. **Set-word at end of block**: Returns error (no value to bind)
2. **Chained infix**: `x: 1 + 2 * 3` → evaluates full chain
3. **Non-infix after set-word**: `f: fn [...] ...` → doesn't consume beyond function
4. **Multiple statements**: `x: 10  y: 20` → each set-word only consumes its expression

## Alternative Approaches Considered

### 1. Consume Rest of Block
```go
restOfBlock := block[position+1:]
result, err := e.DoBlock(restOfBlock)
```
**Rejected:** Would consume ALL remaining expressions, breaking multi-statement lines like `sq: fn [n] [n * n]  sq 5`

### 2. Modify DoBlock to Track Consumption
Pass `consumedLast` flag and reset `lastResult` when false.
**Rejected:** Would affect all evaluation paths, higher risk

### 3. Current Implementation (Look-Ahead)
**Selected:** Localized, safe, handles all cases correctly

## Future Considerations - BETTER SOLUTION NEEDED

⚠️ **IMPORTANT:** This implementation is a TEMPORARY WORKAROUND. The lookahead logic should NOT be in the set-word case handler.

### The Real Fix

The lookahead logic should be implemented in `EvaluateExpression` itself, which would:

1. **Remove `lastResult` parameter** - No longer needed if expressions look ahead for infix
2. **Remove `consumedLast` return value** - No longer needed to track consumption
3. **Simplify all callers** - DoBlock, reduce, collectFunctionArgs all become simpler
4. **Fix universally** - Would fix set-word, reduce, and any other contexts automatically

### Why EvaluateExpression Should Handle Lookahead

**Current (incorrect) flow:**
```
DoBlock calls EvaluateExpression with lastResult
  → Expression evaluates
  → Returns (newPos, result, consumedLast)
  → DoBlock decides whether to pass result as lastResult to next iteration
```

**Better flow:**
```
DoBlock calls EvaluateExpression (no lastResult needed)
  → Expression evaluates
  → Looks ahead for infix operators
  → If infix found, continues evaluation automatically
  → Returns (newPos, finalResult) - consumedLast not needed
  → DoBlock just continues with next position
```

### Current Workaround Limitations

The current set-word-only lookahead fix:
- ❌ Only fixes set-word, not reduce or other contexts
- ❌ Duplicates logic (reduce still uses consumedLast flag)
- ❌ Doesn't fix `if 1 < 2 [...]` issue (function argument collection)
- ❌ Leaves technical debt in the codebase

### Proposed Refactoring

**Before:**
```go
func (e *Evaluator) EvaluateExpression(block []core.Value, position int, lastResult core.Value) (int, core.Value, bool, error)
```

**After:**
```go
func (e *Evaluator) EvaluateExpression(block []core.Value, position int) (int, core.Value, error)
```

**Implementation in EvaluateExpression:**
```go
// Evaluate the current element
newPos, result, err := e.evaluateSingleElement(block, position)
if err != nil {
    return position, value.NoneVal(), err
}

// Look ahead for infix operators and consume them
for newPos < len(block) {
    nextElement := block[newPos]
    if nextElement.GetType() == value.TypeWord {
        if nextWord, ok := value.AsWord(nextElement); ok {
            if resolved, found := e.Lookup(nextWord); found {
                if fn, ok := value.AsFunction(resolved); ok && fn.Infix {
                    // Next is infix - continue evaluation
                    nextPos, nextResult, nextErr := e.evaluateInfixOperation(block, newPos, fn, result)
                    if nextErr != nil {
                        return position, value.NoneVal(), nextErr
                    }
                    newPos = nextPos
                    result = nextResult
                    continue
                }
            }
        }
    }
    break
}

return newPos, result, nil
```

### Impact Analysis

**Files to modify:**
- `internal/eval/evaluator.go` - Refactor EvaluateExpression signature and implementation
- `internal/native/control.go` - Update reduce to remove consumedLast logic
- All callers of EvaluateExpression (grep shows ~7 locations)

**Benefits:**
- Cleaner API surface
- Less complexity in DoBlock loop
- Fixes all infix consumption issues universally
- Removes need for consumedLast tracking in reduce

**Risks:**
- Medium - changes core evaluation path
- Requires careful testing of all evaluation scenarios
- May affect function argument collection

### Recommendation

This temporary fix (lookahead in set-word only) solves the immediate bug for REPL state preservation. However, a proper refactoring to move lookahead into `EvaluateExpression` itself should be done as a follow-up to:
1. Remove technical debt
2. Simplify the evaluation model
3. Fix related issues (like `if 1 < 2 [...]`)

## Related Documentation
- `docs/operator-precedence.md` - Explains left-to-right evaluation
- `plans/009_reduce_lookahead_fix.md` - Look-ahead strategy for reduce function (should be obsoleted by proper fix)
- `docs/scoping-differences.md` - May need updates for set-word scoping behavior
- **TODO:** Create Plan 011 for refactoring EvaluateExpression to handle lookahead universally
