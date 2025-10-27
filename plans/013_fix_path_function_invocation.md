# Plan 013: Fix Path Function Invocation and Unify Path Implementations

## Overview

Fix get-path expressions to invoke functions when they resolve to functions, while ensuring get-words with dots return functions without invocation. Unify the two path implementation approaches for better maintainability.

## Problem Statement

Current path evaluation has two issues:

1. **Missing Function Invocation**: When evaluating a get-path expression like `foo.bar.baz` where `baz` is a function, the function is returned without being invoked. It should be invoked.

2. **Dual Path Implementations**: There are two different path systems:
   - **Proper paths**: `PathExpression` objects created by PEG parser for expressions like `user.address.city`
   - **Words with dots**: String-based parsing for set-words like `user.address.city:` and get-words like `:user.address.city`

3. **Inconsistent Get-Word Behavior**: Get-words with dots (`:foo.bar.baz`) should return functions without invocation, but this isn't clearly distinguished from get-path expressions.

## Goals

1. **Fix Function Invocation**: Get-path expressions that resolve to functions should invoke them (like word lookups)
2. **Preserve Get-Word Semantics**: `:foo.bar.baz` should return functions without invocation
3. **Method Calls**: `obj.method arg1 arg2` should work directly (paths behave like words for function invocation)
4. **Unify Path Systems**: Migrate words-with-dots to use the proper `PathExpression` system
5. **Maintain Backward Compatibility**: All existing code should continue to work
6. **Improve Maintainability**: Single, consistent path evaluation system

## Current Architecture Analysis

### Path Expression Types

1. **Get-Path Expressions** (e.g., `foo.bar.baz`):
   - Parsed as `Path` by PEG grammar
   - Creates `PathExpression` object
   - Evaluated by `evalPathValue()` → `traversePath()`
   - **Should invoke functions**

2. **Set-Path Expressions** (e.g., `foo.bar.baz:`):
   - Parsed as set-word with dots
   - Special-cased in evaluator: `strings.Contains(wordStr, ".")` → `evalSetPathExpression()`
   - Uses `parsePathString()` to create `PathExpression`
   - Evaluated by `assignToPathTarget()`

3. **Get-Word with Dots** (e.g., `:foo.bar.baz`):
   - Parsed as get-word
   - No special dot handling - just does `e.Lookup(wordStr)`
   - **Should NOT invoke functions**

### Evaluation Flow

```
Get-Path:     foo.bar.baz  → PathExpression → evalPathValue → traversePath → RETURN result
Get-Word:     :foo.bar.baz → Lookup "foo.bar.baz" → RETURN result  
Set-Path:     foo.bar.baz: → parsePathString → PathExpression → assignToPathTarget
```

## Recommended Approach: Unified Path System

### Phase 1: Extend PathExpression for Get-Words

1. **Modify PEG Parser**: Update `GetWord` rule to detect and parse dotted get-words as paths
2. **Create GetPathExpression**: New type that wraps `PathExpression` but marks it as non-invoking
3. **Update Evaluator**: Handle `GetPathExpression` differently from regular `PathExpression`

### Phase 2: Fix Function Invocation

1. **Modify `evaluateElement`**: For paths that resolve to functions, invoke them like word lookups
2. **Preserve Get-Word Behavior**: `GetPathExpression` returns functions without invocation

### Phase 3: Migrate Set-Paths

1. **Update PEG Parser**: Handle set-paths with dots in grammar instead of evaluator hack
2. **Remove String Parsing**: Eliminate `parsePathString()` and `strings.Contains()` checks

## Implementation Details

### New Types

```go
// GetPathExpression marks a path as non-invoking (like get-words)
type GetPathExpression struct {
    *PathExpression
}

// Constructor
func NewGetPath(segments []PathSegment, base core.Value) *GetPathExpression {
    return &GetPathExpression{
        PathExpression: NewPath(segments, base),
    }
}
```

### Parser Changes

**Current GetWord rule:**
```
GetWord <- ':' word:WordChars
```

**New GetWord rule:**
```
GetWord <- ':' path:(Path | word:WordChars) {
    if path != nil {
        return value.NewGetPath(path.(*value.PathExpression)), nil
    }
    return value.NewGetWordVal(word.(string)), nil
}
```

### Evaluator Changes

**evalPathValue (modified):**
```go
func (e *Evaluator) evalPathValue(path *value.PathExpression) (core.Value, error) {
    tr, err := traversePath(e, path, false)
    if err != nil {
        return value.NewNoneVal(), err
    }
    result := tr.values[len(tr.values)-1]
    
    // NEW: Invoke functions for get-path expressions
    if result.GetType() == value.TypeFunction {
        return e.invokeFunction(result, []core.Value{}, map[string]core.Value{})
    }
    
    return result, nil
}
```

**evaluateElement (modified):**
```go
case value.TypeGetPath:
    getPath, _ := value.AsGetPath(element)
    result, err := e.evalGetPathValue(getPath)
    return position + 1, result, err
```

**evalGetPathValue (new):**
```go
func (e *Evaluator) evalGetPathValue(getPath *value.GetPathExpression) (core.Value, error) {
    tr, err := traversePath(e, getPath.PathExpression, false)
    if err != nil {
        return value.NewNoneVal(), err
    }
    // Get-paths NEVER invoke functions - just return the result
    return tr.values[len(tr.values)-1], nil
}
```

## Migration Strategy

### Phase 1: Add GetPath Support (Safe)

1. Add `GetPathExpression` type and methods
2. Update parser to create `GetPathExpression` for dotted get-words
3. Add `evalGetPathValue` method
4. Update `evaluateElement` to handle `TypeGetPath`
5. **Tests pass**: Get-words with dots now use proper path evaluation

### Phase 2: Add Function Invocation (Breaking Change)

1. Modify `evalPathValue` to invoke functions
2. **Tests may break**: Existing code expecting functions from paths will now get function results
3. Update affected tests to expect invoked results

### Phase 3: Migrate Set-Paths (Cleanup)

1. Update PEG parser to handle set-paths with dots
2. Remove `strings.Contains()` check in evaluator
3. Remove `parsePathString()` function
4. **Cleanup**: Single path evaluation system

## Risk Assessment

### Medium Risk: Function Invocation Change

**Impact**: Path expressions now invoke functions like word lookups, enabling direct method calls.

**Mitigation**:
- No-arg functions work as before (invoke and return result)
- Functions with args can be called directly: `obj.method arg1 arg2`
- Get-path still returns functions: `:obj.method`

### Medium Risk: Parser Changes

**Impact**: Changes to PEG grammar could affect parsing of edge cases.

**Mitigation**:
- Comprehensive test coverage
- Incremental changes with validation
- Fallback to old behavior if issues

## Success Criteria

1. **Function Invocation**: `foo.bar.baz` invokes function when `baz` is a function (like word lookup)
2. **Get-Word Preservation**: `:foo.bar.baz` returns function without invocation
3. **Method Calls**: `obj.method arg1 arg2` works directly for calling methods with arguments
4. **No-Arg Functions**: `obj.getter` invokes and returns result
5. **Arg Functions Alone**: `obj.method` fails with arity error when called without args
6. **Backward Compatibility**: Existing code continues to work
7. **Unified System**: Single `PathExpression` system for all path operations
8. **Test Coverage**: All path evaluation scenarios tested
9. **Performance**: No regression in path evaluation performance

## Timeline Estimate

- Phase 1 (GetPath support): 4-6 hours
- Phase 2 (Function invocation): 2-3 hours  
- Phase 3 (Set-path migration): 3-4 hours
- Testing & validation: 4-6 hours

**Total**: 13-19 hours

## Open Questions

1. **Should get-words always use path evaluation?** Yes, for consistency.
2. **What if path resolves to function that needs arguments?** For now, assume no-arg functions or handle as error.
3. **Should this affect set-paths?** No, set-paths are for assignment, not invocation.
4. **Migration path for breaking changes?** Document clearly and provide get-word alternative.

## References

- Current path evaluation: `internal/eval/evaluator.go:644` (`evalPathValue`)
- Path parsing: `grammar/viro.peg:118` (`Path` rule)  
- Path types: `internal/value/path.go`
- Tests: `test/contract/objects_test.go` (T082, T083)

## Addressing User Concern: Dotted Word Confusion

**Question**: Why introduce `GetPathExpression` instead of allowing dotted words?

**Answer**: The grammar currently doesn't allow dots in word names (`WordChars` excludes `.`), so `foo.bar` cannot be a word. The current `:foo.bar` does a nonsensical lookup of `"foo.bar"` as a word name (impossible since dots aren't allowed).

`GetPathExpression` fixes this by making `:foo.bar` properly traverse paths, which is the correct behavior.

### Semantic Consistency

The distinction becomes clear and logical:
- `foo.bar` → PathExpression → traverse path + invoke functions
- `:foo.bar` → GetPathExpression → traverse path + return functions  
- `foo` → Word → lookup word
- `:foo` → GetWord → lookup word

This maintains the colon's "get without evaluating" semantics while preserving path traversal for dotted syntax.

### Alternative Considered: Modify GetWord Rule

Instead of `GetPathExpression`, modify the `GetWord` rule:
```
GetWord <- ':' word:WordChars {
    if strings.Contains(word.(string), ".") {
        // Parse as path and mark as non-invoking
        return parseGetPath(word.(string)), nil
    }
    return value.NewGetWordVal(word.(string)), nil
}
```

**Rejected because**: This keeps string-based parsing in the grammar, making it harder to maintain and extend. `GetPathExpression` provides type safety and consistency.

## Implementation Status

### Phase 1: GetPath Support - COMPLETED ✅

**Already Implemented:**
- `GetPathExpression` type in `internal/value/path.go`
- PEG grammar `GetPath` rule for parsing `:word.path` syntax
- `evalGetPathValue` method for non-invoking path evaluation
- `evaluateElement` handles `TypeGetPath` case
- Comprehensive tests in `TestGetPathEvaluation`

**Results:**
- ✅ `:obj.field` returns field values without function invocation
- ✅ `:obj.method` returns function values without invocation
- ✅ Proper path traversal for dotted get-words

### Phase 2: Function Invocation - COMPLETED ✅

**Changes Made:**
1. **Modified `evaluateElement`**: Added function invocation for path expressions that resolve to functions
2. **Fixed evaluator set-word handling**: Changed from `Set()` to `Bind()` to allow new variable creation
3. **Fixed set-path assignment**: Removed field existence check to allow dynamic field creation
4. **Updated tests**: Fixed `TestPathWriteMutation/update_object_field_via_path` to use proper object creation

**Key Changes:**
- `internal/eval/evaluator.go`: Modified path evaluation to invoke functions, fixed set-word binding, allowed dynamic field assignment
- `test/contract/objects_test.go`: Corrected test setup for dynamic object field assignment

**Results:**
- ✅ Path expressions invoke functions: `obj.method` works for no-arg functions
- ✅ Method calls work: `obj.method arg1 arg2` syntax supported
- ✅ Get-paths return functions: `:obj.method` preserves function values
- ✅ Dynamic field assignment: `obj.newField: value` creates new object fields
- ✅ All tests pass

### Phase 3: Set-Path Unification - COMPLETED ✅

**Changes Made:**
1. **Added `SetPath` rule to PEG grammar**: Parses `word.path:` syntax into `SetPathExpression`
2. **Updated `Value` rule**: Added `SetPath` before `SetWord` for proper precedence
3. **Modified `Path` rule**: Removed set-path handling (now only handles get-paths)
4. **Updated evaluator**: Added `TypeSetPath` case, removed string-based dot detection
5. **Removed `parsePathString`**: Eliminated string parsing hack

**Key Changes:**
- `grammar/viro.peg`: New `SetPath` rule, updated parsing precedence
- `internal/eval/evaluator.go`: Added `TypeSetPath` handling, removed `parsePathString`
- Regenerated parser with `make grammar`

**Results:**
- ✅ Set-paths use proper `SetPathExpression` instead of string parsing
- ✅ Unified path system: `PathExpression`, `GetPathExpression`, `SetPathExpression`
- ✅ Cleaner evaluator code without string manipulation
- ✅ All existing set-path tests pass

### Summary

All three phases of path unification are now complete:

1. **GetPath Support**: `:word.path` uses proper path evaluation
2. **Function Invocation**: `word.path` invokes functions like word lookups
3. **Set-Path Unification**: `word.path:` uses proper `SetPathExpression`

The path evaluation system is now fully unified with consistent behavior across all path types.

