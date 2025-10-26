# Parameter Collection Method

This document describes a simple, uniform parameter collection method for both normal and infix functions.

## Core Principle

**Parameter collection must respect operator precedence while maintaining uniform behavior across function types.**

## The Problem

Consider these examples:
1. `foo 2 + 3 * 4 10` (where `foo` takes 2 arguments)
2. `bar 2 < 4` (where `bar` takes 1 argument)

Naively calling `EvaluateExpression` to get the next argument would return just `2` in both cases,
ignoring the fact that `2 + 3 * 4` forms a complete expression.

## The Solution: Expression-Based Collection

### Algorithm

```
collectParameter(block, position, paramSpec):
    if paramSpec.Eval:
        return EvaluateExpression(block, position)
    else:
        return (position + 1, block[position])
```

### Key Insight

**`EvaluateExpression` already handles infix operators correctly via lookahead.**

From `evaluator.go:488-508`:
```go
func (e *Evaluator) EvaluateExpression(block []core.Value, position int) (int, core.Value, error) {
    newPos, result, err := e.evaluateElement(block, position)
    if err != nil {
        return position, value.NoneVal(), err
    }

    // Lookahead for infix operators
    for newPos < len(block) {
        if !e.isNextInfixOperator(block, newPos) {
            break
        }

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

### How It Works

**Example 1**: `foo 2 + 3 * 4 10` (foo takes 2 args)

1. Collect arg 1:
   - `EvaluateExpression([2, +, 3, *, 4, 10], 0)`
   - Evaluates `2`, sees `+` is infix, continues
   - Evaluates `2 + 3 = 5`, sees `*` is infix, continues  
   - Evaluates `5 * 4 = 20`, sees `10` is NOT infix, stops
   - Returns `(position=5, value=20)`

2. Collect arg 2:
   - `EvaluateExpression([10], 5)`
   - Returns `(position=6, value=10)`

3. Result: `foo 20 10`

**Example 2**: `bar 2 < 4` (bar takes 1 arg)

1. Collect arg 1:
   - `EvaluateExpression([2, <, 4], 0)`
   - Evaluates `2`, sees `<` is infix, continues
   - Evaluates `2 < 4 = true`, no more elements
   - Returns `(position=3, value=true)`

2. Result: `bar true`

### Handling Refinements

Refinements are collected using the same principle:

```
readRefinements(block, position):
    while position < len(block) and isRefinement(block[position]):
        refinement = block[position]
        if refinement.TakesValue:
            position, value = evaluateElement(block, position + 1)
            refValues[refinement.Name] = value
        else:
            refValues[refinement.Name] = true
            position++
```

**Key difference**: Refinement values use `evaluateElement` (single value, no infix lookahead)
to prevent capturing expressions meant for the next parameter.

### Infix vs Normal Functions

**Normal function**:
```
collectFunctionArgs(fn, block, startPos, startParam=0, useElementEval=false)
```

**Infix function**:
```
collectFunctionArgs(fn, block, startPos, startParam=1, useElementEval=true)
```

- `startParam=1`: First parameter already provided (left operand)
- `useElementEval=true`: Remaining parameters use `evaluateElement` instead of `EvaluateExpression`

### Why Different Evaluation for Infix?

Consider: `2 + 3 * 4`

1. Left operand `2` is evaluated normally
2. Infix operator `+` is encountered
3. For right operand, we want `3`, not `3 * 4`
4. Using `evaluateElement` gives us `3`, allowing `*` to continue the chain

This maintains left-to-right evaluation: `(2 + 3) * 4 = 20`

## Implementation Requirements

### Current Implementation

The current `collectFunctionArgs` (lines 566-633) already implements this approach:

- Uses `EvaluateExpression` for normal functions
- Uses `evaluateElement` for infix functions
- Handles refinements before each parameter
- Validates parameter count

### Unified Behavior

Both normal and infix functions use the same collection logic, differing only in:
1. Starting parameter index (0 vs 1)
2. Evaluation method (`EvaluateExpression` vs `evaluateElement`)

This ensures:
- Consistent refinement handling
- Proper expression evaluation
- Correct infix chaining
- Simple, maintainable code

## Examples

### Complex Expression
```viro
add: fn [a b] [a + b]
mul: fn [a b] [a * b]

; Normal call with expressions
add 2 + 3 mul 4 5
; → add (2 + 3) (mul 4 5)
; → add 5 20
; → 25

; Infix chain
2 + 3 * 4 - 1
; → ((2 + 3) * 4) - 1
; → (5 * 4) - 1
; → 20 - 1
; → 19
```

### Refinements
```viro
process: fn [data --verbose --limit []] [...]

; Refinements anywhere
process 10 --verbose --limit 5
process --limit 5 10 --verbose
process --verbose 10 --limit 5

; All equivalent to:
; data=10, verbose=true, limit=5
```

## Conclusion

The parameter collection is **already solved** by leveraging `EvaluateExpression`'s infix lookahead.

The key principles:
1. **For normal functions**: Use `EvaluateExpression` (full expression with infix)
2. **For infix functions**: Use `evaluateElement` (single value, no infix lookahead)  
3. **For refinements**: Use `evaluateElement` (prevent over-capturing)
4. **Unified logic**: Same collection loop, different evaluation strategy
