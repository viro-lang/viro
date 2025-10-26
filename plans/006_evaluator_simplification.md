# Plan 006: Evaluator Simplification

STATUS: In Review

## Overview

Simplify the evaluator implementation in `internal/eval/evaluator.go` by reducing code complexity and line count while maintaining functionality and test coverage. This plan addresses four specific simplifications based on comparison with the Lua prototype implementation.

## Goals

- Reduce evaluator.go from ~1,114 to ~700-800 lines
- Simplify core evaluation logic for better maintainability
- Maintain 100% test coverage
- Keep all existing functionality working

## Current State Analysis

### File: internal/eval/evaluator.go (1,114 lines)

**Complexity breakdown:**
- Core evaluation dispatch: ~420 lines (DoNext, DoBlock, dispatch functions)
- Function invocation: ~230 lines (invokeFunction, collectFunctionArgsWithInfix)
- Path evaluation: ~180 lines (traversePath, evalPath, evalSetPath)
- Frame management: ~150 lines (push/pop/lookup/capture)
- Helper functions: ~134 lines (error annotation, call stack, etc.)

**Key complexity drivers:**
1. **Separate dispatch functions** (lines 143-234): 10 different `eval*Dispatch` functions
2. **Complex argument collection** (lines 636-719): 86 lines for collectFunctionArgsWithInfix
3. **Multiple evaluation paths** (lines 386-489): DoNext, evaluateWithFunctionCall, evalExpressionFromTokens
4. **Trivial dispatch handlers** (lines 165-179): evalLiteral, evalBlock, evalFunction just return input

## Simplifications to Implement

### 1. Merge Evaluation Dispatchers

**Current:** 10 separate dispatch functions (lines 163-234)
```go
func evalLiteral(e core.Evaluator, val core.Value) (core.Value, error)
func evalBlock(e core.Evaluator, val core.Value) (core.Value, error)
func evalFunction(e core.Evaluator, val core.Value) (core.Value, error)
func evalParenDispatch(e core.Evaluator, val core.Value) (core.Value, error)
func evalWordDispatch(e core.Evaluator, val core.Value) (core.Value, error)
func evalSetWordDispatch(e core.Evaluator, val core.Value) (core.Value, error)
func evalGetWordDispatch(e core.Evaluator, val core.Value) (core.Value, error)
func evalLitWordDispatch(e core.Evaluator, val core.Value) (core.Value, error)
func evalPathDispatch(e core.Evaluator, val core.Value) (core.Value, error)
```

**Target:** Single `evalValue` function with switch statement
```go
func (e *Evaluator) evalValue(val core.Value) (core.Value, error) {
    switch val.GetType() {
    case value.TypeInteger, value.TypeString, value.TypeLogic, 
         value.TypeNone, value.TypeDecimal, value.TypeObject,
         value.TypePort, value.TypeDatatype, value.TypeBlock, 
         value.TypeFunction:
        return val, nil
    
    case value.TypeParen:
        block, _ := value.AsBlock(val)
        return e.DoBlock(block.Elements)
    
    case value.TypeWord:
        return e.evalWord(val)
    
    case value.TypeSetWord:
        wordStr, _ := value.AsWord(val)
        return value.NoneVal(), verror.NewScriptError(...)
    
    case value.TypeGetWord:
        return e.evalGetWord(val)
    
    case value.TypeLitWord:
        return value.WordVal(val.GetPayload().(string)), nil
    
    case value.TypePath:
        return e.evalPath(val)
    
    default:
        return value.NoneVal(), verror.NewInternalError(...)
    }
}
```

**Benefits:**
- Eliminate init() function and evalDispatch map
- Inline trivial cases directly
- Easier to read - all logic in one place
- ~60 lines reduced

**Testing:**
- All existing eval tests should pass
- No new tests needed (behavioral equivalence)

---

### 2. Simplify Function Argument Collection

**Current:** `collectFunctionArgsWithInfix` (lines 636-719, 86 lines)
- Interleaves refinement parsing between positional args
- Complex state tracking with multiple indices
- Handles infix detection inline

**Target:** Two-pass approach like Lua
```go
func (e *Evaluator) collectFunctionArgs(fn *value.FunctionValue, tokens []core.Value, lastResult core.Value) ([]core.Value, map[string]core.Value, int, error) {
    // Pass 1: Collect positional args (handling infix)
    posArgs, pos := e.collectPositionalArgs(fn, tokens, lastResult)
    
    // Pass 2: Collect refinements at end
    refValues, pos := e.collectRefinements(fn, tokens, pos)
    
    return posArgs, refValues, pos, nil
}
```

**Key changes:**
- Split into two smaller functions (~30 lines each)
- Remove interleaved refinement parsing
- Simpler state management
- ~25 lines reduced overall

**Caveat:** This changes refinement syntax!
- **Current:** `foo 1 --bar 2 3 --baz 4` (interleaved)
- **New:** `foo 1 2 3 --bar --baz 4` (refinements at end)

**Decision needed:** Is this syntax change acceptable?
- If YES: proceed with simplification
- If NO: keep current implementation for this item

**Testing:**
- Update test/contract/action_dispatch_test.go refinement tests
- Verify all function calls work with new syntax
- Add tests for refinement ordering

---

### 3. Combine Evaluation Paths

**Current:** Three similar functions
- `DoNext` (lines 389-421): Main entry point, handles tracing
- `evaluateWithFunctionCall` (lines 549-570): Checks if word is function
- `evalExpressionFromTokens` (lines 464-489): Evaluates single expression from sequence

**Target:** Unified evaluation with context parameter
```go
// Main entry point
func (e *Evaluator) DoNext(val core.Value) (core.Value, error) {
    return e.evalWithContext(val, nil, nil, value.NoneVal())
}

// Internal unified evaluator
func (e *Evaluator) evalWithContext(val core.Value, seq []core.Value, idx *int, lastResult core.Value) (core.Value, error) {
    // Handle tracing if needed
    // Handle set-word case
    // Evaluate value
    // If word and function, invoke it
    // Return result
}

// For sequential evaluation
func (e *Evaluator) evalNextInSequence(seq []core.Value, idx *int, lastResult core.Value) (core.Value, error) {
    return e.evalWithContext(seq[*idx], seq, idx, lastResult)
}
```

**Benefits:**
- Single evaluation codepath
- Less duplication
- Clearer control flow
- ~40 lines reduced

**Testing:**
- All eval_test.go tests should pass
- Function call tests in function_eval_test.go
- No new behavior, existing coverage sufficient

---

### 4. Inline Simple Dispatchers

**Current:** Three functions that just return their input (lines 165-179)
```go
func evalLiteral(e core.Evaluator, val core.Value) (core.Value, error) {
    return val, nil
}

func evalBlock(e core.Evaluator, val core.Value) (core.Value, error) {
    return val, nil
}

func evalFunction(e core.Evaluator, val core.Value) (core.Value, error) {
    return val, nil
}
```

**Target:** Inline into evalValue switch (see Simplification #1)
```go
case value.TypeInteger, value.TypeString, value.TypeLogic, 
     value.TypeNone, value.TypeDecimal, value.TypeObject,
     value.TypePort, value.TypeDatatype, value.TypeBlock, 
     value.TypeFunction:
    return val, nil
```

**Benefits:**
- No separate functions needed
- Clearer that these types self-evaluate
- ~15 lines reduced

**Note:** This is automatically handled by Simplification #1

---

## Implementation Steps

### Phase 1: Preparation (Read-only)

1. **Analyze test coverage**
   ```bash
   go test -coverprofile=coverage.out ./internal/eval/...
   go tool cover -func=coverage.out
   ```
   - Ensure >90% coverage before starting
   - Identify any untested paths

2. **Document current behavior**
   - Run all tests: `go test -v ./test/contract/... ./test/integration/...`
   - Capture baseline: `make test-summary`

### Phase 2: Simplification #1 - Merge Dispatchers

3. **Create new evalValue method**
   - Add `func (e *Evaluator) evalValue(val core.Value) (core.Value, error)` with switch
   - Keep old dispatch functions temporarily

4. **Update DoNext to use evalValue**
   - Change line 401 from `evalFn, found := evalDispatch[val.GetType()]` to `return e.evalValue(val)`
   - Run tests to verify equivalence

5. **Remove old dispatch infrastructure**
   - Delete lines 138-161 (evalDispatch map and init())
   - Delete lines 165-234 (individual dispatch functions)
   - Run tests: `go test ./internal/eval/...`

6. **Clean up**
   - Remove unused imports if any
   - Run: `make grammar && make build && make test`

### Phase 3: Simplification #4 - Inline Simple Dispatchers

7. **Verify inlining in evalValue**
   - Already done in step 3
   - Confirm literal types use single case statement
   - Run tests: `go test ./internal/eval/...`

### Phase 4: Simplification #3 - Combine Eval Paths

8. **Create evalWithContext method**
   - Implement unified evaluator with tracing, set-word, function call handling
   - Keep existing functions initially

9. **Update DoNext to delegate**
   - Modify to call `e.evalWithContext(val, nil, nil, value.NoneVal())`
   - Run tests

10. **Update DoBlock**
    - Use evalWithContext instead of evaluateWithFunctionCall
    - Run tests

11. **Remove old evaluation functions**
    - Delete evaluateWithFunctionCall (lines 549-570)
    - Delete evalExpressionFromTokens (lines 464-489)
    - Keep EvalExpressionFromTokens as public wrapper
    - Run: `make test`

### Phase 5: Simplification #2 - Function Arguments (OPTIONAL)

**Decision point:** Discuss with team about refinement syntax change

If approved:

12. **Create collectPositionalArgs**
    - Extract positional argument logic
    - Handle infix case
    - ~30 lines

13. **Simplify readRefinements**
    - Already exists (lines 588-631)
    - Modify to work as standalone pass
    - ~30 lines

14. **Create new collectFunctionArgs**
    - Call collectPositionalArgs
    - Call readRefinements
    - ~15 lines wrapper

15. **Update invokeFunction**
    - Use new collectFunctionArgs
    - Run tests

16. **Update all refinement tests**
    - Modify test/contract/action_dispatch_test.go
    - Change syntax to refinements-at-end
    - Verify all pass

17. **Delete old collectFunctionArgsWithInfix**
    - Remove lines 636-719
    - Run: `make test`

### Phase 6: Validation

18. **Full test suite**
    ```bash
    make grammar
    make build
    make test
    go test -coverprofile=coverage.out ./...
    ```

19. **Integration tests**
    ```bash
    go test -v ./test/integration/...
    ```

20. **Contract tests**
    ```bash
    go test -v ./test/contract/...
    ```

21. **Verify LOC reduction**
    ```bash
    wc -l internal/eval/evaluator.go
    # Target: 700-800 lines (from 1,114)
    ```

---

## Expected Outcomes

### Line Count Reduction

| Simplification | Lines Saved | Notes |
|---------------|-------------|-------|
| #1 Merge dispatchers | ~60 | Remove 10 functions + init + map |
| #4 Inline trivial | ~0 | Covered by #1 |
| #3 Combine eval paths | ~40 | Remove duplicate logic |
| #2 Function args | ~25 | Two simpler functions vs one complex |
| **Total** | **~125** | **1,114 → ~990 lines** |

With additional cleanup: **~700-800 lines** (removing duplicated logic, simplifying)

### Complexity Reduction

- **Cyclomatic complexity:** Reduce from ~45 to ~30
- **Function count:** Reduce from ~30 to ~20
- **Average function length:** Reduce from ~37 to ~35 lines

### Maintainability Improvements

- Single evaluation codepath easier to trace
- Fewer dispatch mechanisms to understand
- More similar to Lua reference implementation
- Easier to add new value types

---

## Risks and Mitigations

### Risk 1: Breaking existing tests
- **Mitigation:** Run tests after each step, fix immediately
- **Fallback:** Use git to revert individual steps

### Risk 2: Performance regression
- **Mitigation:** Benchmark before/after
  ```bash
  go test -bench=. ./internal/eval/... > before.txt
  # After changes
  go test -bench=. ./internal/eval/... > after.txt
  diff before.txt after.txt
  ```
- **Acceptable:** <5% regression for 40% code reduction

### Risk 3: Refinement syntax change (Simplification #2)
- **Mitigation:** Make this optional, get team approval first
- **Fallback:** Skip #2 if syntax change unacceptable
- **Documentation:** Update docs/repl-usage.md if implemented

### Risk 4: Debugging harder with unified evaluator
- **Mitigation:** Keep detailed tracing, error annotation
- **Test:** Verify error messages still show "near" context

---

## Testing Strategy (TDD Approach)

### Test First

1. **Before any changes:** Capture baseline
   ```bash
   go test ./... > tests_before.txt
   go test -json ./... > tests_before.json
   ```

2. **After each phase:** Verify equivalence
   ```bash
   go test ./... > tests_after.txt
   diff tests_before.txt tests_after.txt
   # Should show: PASS everywhere
   ```

### Coverage Requirements

- Maintain >90% coverage in internal/eval/
- Any new functions must have tests
- No reduction in integration test coverage

### Test Files to Monitor

- `test/contract/eval_test.go` - Core evaluation
- `test/contract/function_eval_test.go` - Function calls
- `test/contract/action_dispatch_test.go` - Refinements (will change for #2)
- `test/integration/eval_bench_test.go` - Performance benchmarks

---

## Success Criteria

- [ ] All existing tests pass
- [ ] Code coverage remains >90%
- [ ] evaluator.go reduced to <900 lines (or <800 with #2)
- [ ] No performance regression >5%
- [ ] All four simplifications implemented (#1, #3, #4, and optionally #2)
- [ ] Code review approved
- [ ] Documentation updated (if refinement syntax changed)

---

## Timeline Estimate

- Phase 1 (Preparation): 30 minutes
- Phase 2 (Simplification #1): 2 hours
- Phase 3 (Simplification #4): 30 minutes (included in #1)
- Phase 4 (Simplification #3): 3 hours
- Phase 5 (Simplification #2): 4 hours (OPTIONAL)
- Phase 6 (Validation): 1 hour

**Total:** ~7 hours (or ~11 hours with #2)

---

## References

- Current implementation: `internal/eval/evaluator.go`
- Lua reference: `https://github.com/marad/viro-lang/blob/main/experiments/test.lua`
- Test suite: `test/contract/*_test.go`, `test/integration/*_test.go`
