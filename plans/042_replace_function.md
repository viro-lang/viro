# Implementation Plan: Native `replace` Function

## Feature Summary

Implement the native `replace` function in Viro for replacing occurrences of a pattern in a series (string or block) with a replacement value. The function supports in-place modification by default, with optional refinements for working on a copy and replacing all occurrences instead of just the first.

**Signature**:
```viro
replace series pattern replacement           ; modifies in-place, returns modified series
replace/copy series pattern replacement      ; returns new copy with replacements
replace/all series pattern replacement       ; replaces all occurrences (default: only first)
replace/copy/all series pattern replacement  ; combines both refinements
```

## Research Findings

### Codebase Architecture

1. **Native implementation location**: `internal/native/`
   - Series operations: `series_string.go`, `series_block.go`, `series_polymorphic.go`
   - Registration: `register_series.go`

2. **Testing location**: `test/contract/`
   - Series tests: `series_test.go`
   - Table-driven test pattern with `[]struct{name, input, want, wantErr, errID}`

3. **Value constructors**:
   - `value.NewStrVal(string)` - create string value
   - `value.NewBlockVal([]core.Value)` - create block value
   - `value.AsStringValue(core.Value)` - extract string value
   - `value.AsBlockValue(core.Value)` - extract block value
   - `value.SeriesVal()` - series creation interface

4. **Error patterns**:
   - `verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"expected", "got", "context"})`
   - `verror.NewScriptError(verror.ErrIDInvalidOperation, [3]string{"message", "", ""})`

### Similar Existing Functions

Examined for patterns:
- **`find`**: Locates values in series (strings use substring matching, blocks use equality)
- **`change`**: Modifies element at current position
- **`trim`**: Removes values with `/all` refinement support
- **`copy`**: Creates copies with `/part` refinement
- **`split`**: String manipulation returning blocks

### Key Observations

1. **String operations** use Go's `strings` package for pattern matching
2. **Block operations** use `value.Equals()` for element comparison
3. **Refinements** are checked via `refValues map[string]core.Value`
4. **In-place modification** is the default for series operations
5. **Type-specific implementations** are registered per series type (block, string, binary)

## Architecture Overview

### Implementation Approach

The `replace` function will be implemented as:
1. **Type-specific actions**: Separate implementations for strings and blocks
2. **Registration pattern**: Register as series action for TypeString and TypeBlock/TypeParen
3. **Refinement support**: Handle `/copy` and `/all` refinements independently or combined

### Design Decisions

**Decision 1: In-place vs Copy**
- **Default**: In-place modification (consistent with `change`, `trim`, `remove`)
- **With `/copy`**: Work on copy, leave original unchanged

**Decision 2: First vs All**
- **Default**: Replace only first occurrence (consistent with `find` behavior)
- **With `/all`**: Replace all occurrences

**Decision 3: Type Handling**
- **Strings**: Pattern must be string, use substring matching
- **Blocks**: Pattern can be any value, use `value.Equals()` for comparison
- **Type mismatch**: Return appropriate error

**Decision 4: Empty Pattern**
- **Behavior**: Return script error with `ErrIDInvalidOperation`
- **Rationale**: Empty pattern is ambiguous and likely a programming error

## Implementation Roadmap

### Phase 1: Test-Driven Development (TDD)

#### Step 1.1: Create tests in `test/contract/replace_test.go`

**Test file structure**:
```go
func TestSeries_Replace(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    core.Value
        wantErr bool
        errID   string
    }{
        // String replacement tests
        // Block replacement tests
        // Copy mode tests
        // All mode tests
        // Edge case tests
        // Error case tests
    }
}
```

**Test categories**:

1. **String replacement - first occurrence only**:
   - Basic replacement: `replace "hello world" "world" "Viro"` → `"hello Viro"`
   - Pattern not found: `replace "hello" "xyz" "abc"` → `"hello"`
   - Multiple occurrences: `replace "abc abc abc" "abc" "x"` → `"x abc abc"`
   - Pattern at start: `replace "hello world" "hello" "hi"` → `"hi world"`
   - Pattern at end: `replace "hello world" "world" "Viro"` → `"hello Viro"`

2. **String replacement - all occurrences**:
   - All occurrences: `replace/all "abc abc abc" "abc" "x"` → `"x x x"`
   - Overlapping patterns: `replace/all "aaa" "aa" "b"` → `"ba"` (left-to-right, no re-scanning)
   - Pattern not found: `replace/all "hello" "xyz" "abc"` → `"hello"`

3. **Block replacement - first occurrence only**:
   - Basic replacement: `replace [1 2 3 2 1] 2 99` → `[1 99 3 2 1]`
   - Pattern not found: `replace [1 2 3] 4 99` → `[1 2 3]`
   - Multiple occurrences: `replace [a a a] 'a 'b` → `[b a a]`

4. **Block replacement - all occurrences**:
   - All occurrences: `replace/all [1 2 3 2 1] 2 99` → `[1 99 3 99 1]`
   - Pattern not found: `replace/all [1 2 3] 4 99` → `[1 2 3]`

5. **Copy mode**:
   - String copy mode: Verify original unchanged
   - Block copy mode: Verify original unchanged
   - Copy with all: `replace/copy/all`

6. **Edge cases**:
   - Empty series: `replace "" "x" "y"` → `""`
   - Empty series: `replace [] 1 2` → `[]`
   - Empty pattern error: `replace "hello" "" "x"` → error
   - Empty replacement: `replace "hello" "l" ""` → `"helo"`

7. **Type mixing**:
   - Integer in block: `replace [1 2 3] 2 "two"` → `[1 "two" 3]`
   - String in block: `replace ["a" "b"] "a" 99` → `[99 "b"]`

8. **Error cases**:
   - Empty pattern: Error with `ErrIDInvalidOperation`
   - Non-series input: Error with `ErrIDActionNoImpl`
   - String with non-string pattern: Error with `ErrIDTypeMismatch`
   - Wrong arity: Verify behavior

### Phase 2: Implementation

#### Step 2.1: Implement `StringReplace` in `internal/native/series_string.go`

**Function signature**:
```go
func StringReplace(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error)
```

**Implementation approach**:
1. Extract and validate arguments (series, pattern, replacement - all must be strings)
2. Check refinements: `copy` (boolean) and `all` (boolean)
3. If `/copy`, clone the string series
4. Handle empty pattern → error
5. Use `strings.Replace(str, pattern, replacement, n)` where:
   - `n = 1` for first occurrence (default)
   - `n = -1` for all occurrences (`/all` refinement)
6. Update string runes in-place (or on copy)
7. Return modified series

**Key decisions**:
- Use Go's `strings.Replace()` for efficient implementation
- Handle multi-byte UTF-8 correctly via rune operations
- Preserve series index position

#### Step 2.2: Implement `BlockReplace` in `internal/native/series_block.go`

**Function signature**:
```go
func BlockReplace(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error)
```

**Implementation approach**:
1. Extract and validate arguments (series must be block, pattern and replacement can be any value)
2. Check refinements: `copy` (boolean) and `all` (boolean)
3. If `/copy`, clone the block
4. Iterate through block elements
5. Compare each element with pattern using `value.Equals()`
6. On match:
   - Replace element with replacement value
   - If not `/all`, break after first replacement
7. Return modified block

**Key decisions**:
- Use `value.Equals()` for comparison (handles all value types correctly)
- Allow heterogeneous replacements (any type can replace any type)
- Preserve block structure and index

### Phase 3: Registration

#### Step 3.1: Register `replace` for strings in `register_series.go`

**In `registerStringSeriesActions()`**:
```go
RegisterActionImpl(value.TypeString, "replace", value.NewNativeFunction("replace", []value.ParamSpec{
    value.NewParamSpec("series", true),
    value.NewParamSpec("pattern", true),
    value.NewParamSpec("replacement", true),
    value.NewRefinementSpec("copy", false),
    value.NewRefinementSpec("all", false),
}, StringReplace, false, nil))
```

#### Step 3.2: Register `replace` for blocks in `register_series.go`

**In `registerBlockSeriesActions()`**:
```go
RegisterActionImpl(typ, "replace", value.NewNativeFunction("replace", []value.ParamSpec{
    value.NewParamSpec("series", true),
    value.NewParamSpec("pattern", true),
    value.NewParamSpec("replacement", true),
    value.NewRefinementSpec("copy", false),
    value.NewRefinementSpec("all", false),
}, BlockReplace, false, nil))
```

#### Step 3.3: Register global `replace` action

**In `RegisterSeriesNatives()`**:
```go
registerAndBind("replace", CreateAction("replace", []value.ParamSpec{
    value.NewParamSpec("series", true),
    value.NewParamSpec("pattern", true),
    value.NewParamSpec("replacement", true),
    value.NewRefinementSpec("copy", false),
    value.NewRefinementSpec("all", false),
}, &NativeDoc{
    Category: "Series",
    Summary:  "Replaces pattern occurrences in a series with a replacement value",
    Description: `Replaces occurrences of a pattern in a series (string or block).

By default: replaces only the first occurrence and modifies in-place.
With /copy: works on a copy, leaving original unchanged.
With /all: replaces all occurrences instead of just the first.
Refinements can be combined: /copy/all creates a copy and replaces all.

For strings: uses substring matching.
For blocks: uses value equality (any type can match any type).`,
    Parameters: []ParamDoc{
        {Name: "series", Type: "string! block!", Description: "Series to search and modify"},
        {Name: "pattern", Type: "any!", Description: "Value/substring to find (must be string for string series)"},
        {Name: "replacement", Type: "any!", Description: "Value/substring to replace with (must be string for string series)"},
        {Name: "--copy", Type: "flag", Description: "Work on copy, don't modify original", Optional: true},
        {Name: "--all", Type: "flag", Description: "Replace all occurrences (default: first only)", Optional: true},
    },
    Returns: "string! block! The modified series (or copy if /copy used)",
    Examples: []string{
        `replace "hello world" "world" "Viro"  ; => "hello Viro"`,
        "replace [1 2 3 2 1] 2 99  ; => [1 99 3 2 1]",
        `replace/all "abc abc" "abc" "x"  ; => "x x"`,
        "replace/all [1 2 1 2] 2 99  ; => [1 99 1 99]",
        `s: "test" replace/copy s "t" "x"  ; s unchanged, returns "xest"`,
    },
    SeeAlso: []string{"find", "change", "trim", "remove"},
    Tags:    []string{"series", "modification", "search"},
}))
```

## Integration Points

### Series System Integration
- **Depends on**: String and block series implementations
- **Uses**: `value.Equals()` for block comparisons
- **Uses**: Go `strings` package for string operations
- **Integrates with**: Existing series action dispatch system

### Refinement System Integration
- **Pattern**: Check `refValues["copy"]` and `refValues["all"]`
- **Validation**: Boolean logic values expected
- **Combination**: Both refinements can be used together

## Testing Strategy

### Test Coverage Requirements

1. **String replacement tests** (12 tests):
   - Basic first occurrence replacement
   - Pattern not found
   - Multiple occurrences (first only)
   - All occurrences replacement
   - Pattern at start/end
   - Empty series
   - Empty replacement (deletion)

2. **Block replacement tests** (8 tests):
   - Basic first occurrence replacement
   - Pattern not found
   - Multiple occurrences (first only)
   - All occurrences replacement
   - Empty block
   - Type mixing

3. **Copy mode tests** (4 tests):
   - String copy mode (verify original unchanged)
   - Block copy mode (verify original unchanged)
   - Copy with all refinement
   - Copy alone vs copy/all comparison

4. **Edge case tests** (4 tests):
   - Empty pattern error
   - Empty series handling
   - Empty replacement
   - Large series performance

5. **Error tests** (3 tests):
   - Non-series input
   - String with non-string pattern
   - Empty pattern

**Total**: ~30 tests minimum

### Validation Criteria

**All tests must**:
- Follow table-driven test pattern
- Use `Evaluate()` helper for integration testing
- Check both success and error cases
- Validate error IDs match expected categories
- Use `Equals()` for value comparison
- Verify original unchanged when `/copy` is used

**Code quality**:
- NO COMMENTS in implementation code
- Use value constructors (`NewStrVal`, `NewBlockVal`)
- Use proper error categories (`ErrIDInvalidOperation`, `ErrIDTypeMismatch`, `ErrIDActionNoImpl`)
- Follow existing patterns from `find`, `change`, `trim`

## Potential Challenges

### Challenge 1: UTF-8 String Handling
**Issue**: Multi-byte Unicode characters in pattern/replacement

**Mitigation**: 
- Use rune-based operations throughout
- Test with multi-byte characters: `replace "żółć" "ó" "o"`
- Leverage existing string series rune handling

**Verification**: Add UTF-8 test cases

### Challenge 2: Overlapping Patterns
**Issue**: `replace/all "aaa" "aa" "b"` - should it produce `"ba"` or `"b"`?

**Decision**: Left-to-right, no re-scanning (standard `strings.Replace` behavior)
- First match replaces, index advances past replacement
- Result: `"ba"` (matches first "aa", leaves trailing "a")

**Rationale**: Consistent with Go standard library, predictable behavior

### Challenge 3: Copy Semantics
**Issue**: Ensuring deep copy for blocks with nested structures

**Solution**: Use existing series `Clone()` method
- Block clone already handles nested values correctly
- String clone creates independent rune array

**Verification**: Test with nested blocks

### Challenge 4: In-place Modification Safety
**Issue**: Mutating series during iteration

**Solution**: 
- For strings: Build new string, then update runes atomically
- For blocks: Direct element replacement (no structural changes)

**Verification**: Test that series index remains valid after replacement

## Viro Guidelines Reference

### Followed Guidelines

1. ✓ **TDD mandatory**: Write tests FIRST in `test/contract/replace_test.go`
2. ✓ **NO COMMENTS**: Documentation in package docs only
3. ✓ **Use constructors**: `value.NewStrVal()`, `value.NewBlockVal()`
4. ✓ **Error handling**: Use `verror.NewScriptError()` with appropriate categories
5. ✓ **Table-driven tests**: `[]struct{name, input, want, wantErr, errID}`
6. ✓ **Index-based refs**: Series use index-based positioning
7. ✓ **Naming conventions**: Use `replace` (lowercase), refinements as `/copy`, `/all`

### Specific Code Patterns

**Refinement checking**:
```go
hasRefinement := func(refName string) bool {
    val, ok := refValues[refName]
    return ok && val.GetType() == value.TypeLogic && val.Equals(value.NewLogicVal(true))
}

copyMode := hasRefinement("copy")
replaceAll := hasRefinement("all")
```

**Type checking**:
```go
str, ok := value.AsStringValue(args[0])
if !ok {
    return value.NewNoneVal(), verror.NewScriptError(
        verror.ErrIDTypeMismatch, 
        [3]string{"string", value.TypeToString(args[0].GetType()), ""},
    )
}
```

**Error reporting**:
```go
if patternStr == "" {
    return value.NewNoneVal(), verror.NewScriptError(
        verror.ErrIDInvalidOperation,
        [3]string{"empty pattern not allowed in replace", "", ""},
    )
}
```

## Implementation Checklist

### Pre-Implementation
- [ ] Review similar series operations (`find`, `change`, `trim`)
- [ ] Understand refinement handling patterns
- [ ] Identify test file location and patterns
- [ ] Verify value constructor usage

### Phase 1: Tests (TDD)
- [ ] Create `test/contract/replace_test.go`
- [ ] Write string replacement tests (first occurrence)
- [ ] Write string replacement tests (all occurrences)
- [ ] Write block replacement tests (first occurrence)
- [ ] Write block replacement tests (all occurrences)
- [ ] Write copy mode tests
- [ ] Write edge case tests
- [ ] Write error case tests
- [ ] Run tests, confirm all FAIL (not yet implemented)

### Phase 2: Implementation
- [ ] Create `StringReplace` in `internal/native/series_string.go`
- [ ] Create `BlockReplace` in `internal/native/series_block.go`
- [ ] Handle refinements (`/copy`, `/all`)
- [ ] Implement empty pattern validation
- [ ] Implement pattern matching (substring for strings, equals for blocks)
- [ ] Implement replacement logic
- [ ] Handle copy semantics correctly

### Phase 3: Registration
- [ ] Register `StringReplace` in `registerStringSeriesActions()`
- [ ] Register `BlockReplace` in `registerBlockSeriesActions()`
- [ ] Register global `replace` action with full NativeDoc
- [ ] Verify registration doesn't conflict with existing natives

### Phase 4: Validation
- [ ] Run tests: `go test ./test/contract -run TestSeries_Replace`
- [ ] Verify all tests pass
- [ ] Run full test suite: `go test ./...`
- [ ] Manual testing: `./viro -c 'replace "test" "t" "x"'`
- [ ] Manual testing: `./viro -c 'replace [1 2 3 2] 2 99'`
- [ ] Manual testing: `./viro -c 'replace/all "aaa" "a" "b"'`
- [ ] Manual testing: `./viro -c 'replace/copy [1 2 3] 2 99'`

### Final Verification
- [ ] Code review via viro-reviewer agent
- [ ] No comments in implementation code
- [ ] All tests passing
- [ ] Error messages clear and consistent
- [ ] Proper integration with series action system
- [ ] UTF-8 handling verified
- [ ] Copy semantics verified (original unchanged)

## Success Criteria

**Implementation complete when**:
1. All tests pass (100% of new tests, ~30 tests)
2. Manual testing succeeds for all documented examples
3. Code review approved by viro-reviewer agent
4. Full test suite passes: `go test ./...`
5. No regressions in existing series operations

**Quality metrics**:
- Test coverage: 100% of new functions (both branches)
- Error handling: All edge cases covered
- Documentation: Complete NativeDoc with examples
- Code style: Zero comments, follows Viro conventions
- Performance: No O(n²) operations for `/all` mode

## Post-Implementation Notes

_(This section will be filled during implementation to track any deviations, fixes, or insights)_

---

**Plan Version**: 1.0  
**Created**: 2025-12-15  
**Plan Number**: 042  
**Status**: Ready for implementation
