# Implementation Plan: Add `--deep` Refinement to `copy` Function

## Feature Summary
Add a `--deep` boolean refinement to the `copy` function that performs deep copying of nested mutable containers (blocks, objects) while sharing immutable values. The refinement defaults to `false` (shallow copy) for backward compatibility and can be combined with the existing `--part` refinement.

## Research Findings

### Current Implementation Analysis
1. **`copy` function location**: `internal/native/series_polymorphic.go:seriesCopy()`
2. **Refinement handling**: Uses `readPartCount()` and `validatePartCount()` helpers from `series_helpers.go`
3. **Series copying**: Delegates to `Series.CopyPart()` method implemented per series type:
   - `BlockValue.CopyPart()` (block.go) - shallow copy of element references
   - `StringValue.CopyPart()` (string.go) - creates new string (already immutable)
   - `BinaryValue.CopyPart()` (binary.go) - creates new binary (already immutable)
4. **Registration**: `register_series.go` defines the action with `--part` refinement only
5. **Refinement detection**: `hasRefinement()` helper checks for boolean logic! true values

### Key Design Decisions
- **Immutable values**: Strings, binaries, integers, decimals, logic, none, functions remain shared
- **Mutable containers**: Blocks (and parens) and objects require recursive copying
- **Cycle detection**: Use visited map with pointer equality to prevent infinite recursion
- **Depth limit**: Implement safety limit (1000) to prevent stack overflow
- **Prototype preservation**: Object copies maintain same prototype chain

## Architecture Overview

### Core Components
1. **Refinement registration**: Add `--deep` refinement to copy action signature
2. **Deep copy engine**: Recursive value copier with cycle detection
3. **Series integration**: Modify `seriesCopy()` to invoke deep copy when refinement active
4. **Object support**: Extend deep copy to handle object! type (non-series but nested)

### Data Flow
```
copy --deep series
    → seriesCopy() detects --deep refinement
    → series.CopyPart() creates shallow copy
    → deepCopySeries() processes shallow copy recursively
    → deepCopyValue() handles individual values
        → Blocks: create new block with deep copied elements
        → Objects: create new object with deep copied fields (same prototype)
        → Immutable types: return same reference
```

## Implementation Roadmap

### Phase 1: Refinement Registration & Detection
1. **Update `register_series.go`**:
   - Add `value.NewRefinementSpec("deep", false)` to copy action registration
   - Update documentation in NativeDoc to describe deep copy behavior
   - Apply to block, string, and binary series registrations

2. **Modify `seriesCopy()` function**:
   - Use `hasRefinement(refValues, "deep")` to detect deep mode
   - When deep=true, pass result to deep copy processor
   - Maintain compatibility with `--part` refinement (deep copy after part extraction)

### Phase 2: Deep Copy Engine
1. **Create `internal/native/deep_copy.go`**:
   ```go
   package native
   
   func deepCopySeries(series value.Series, visited map[core.Value]core.Value, depth int) (value.Series, error)
   func deepCopyValue(val core.Value, visited map[core.Value]core.Value, depth int) (core.Value, error)
   ```

2. **Algorithm**:
   - **Base case**: depth > 1000 → return error
   - **Cycle detection**: Check visited map before recursion
   - **Immutable types**: Return same value (int, string, binary, logic, none, function, etc.)
   - **Block/paren**: Create new block with deep copied elements
   - **Object**: Create new object with cloned frame, deep copy field values, preserve prototype
   - **Visited tracking**: Store original→copy mapping before recursing into children

3. **Helper functions**:
   - `isImmutableType(core.ValueType) bool` - identify types that never need copying
   - `createObjectCopy(*ObjectInstance, map[core.Value]core.Value, int) (*ObjectInstance, error)`

### Phase 3: Object Deep Copy Support
1. **Object copying strategy**:
   - Use `Frame.Clone()` for shallow frame copy
   - Iterate `Frame.GetAll()` bindings, deep copy values
   - Set copied values back to cloned frame
   - Preserve `ParentProto` reference (prototype sharing)

2. **Integration**:
   - Add deep copy support in `deepCopyValue()` for `TypeObject`
   - Handle potential cycles via visited map

### Phase 4: Testing & Validation
1. **Unit tests in `internal/value/series_test.go`**:
   - Test deep copy of nested blocks
   - Test deep copy with objects in blocks
   - Test cycle detection (self-referential structures)
   - Test combination with `--part` refinement
   - Test immutable value sharing

2. **Integration tests in `test/contract/series_test.go`**:
   - End-to-end Viro script tests for `copy/deep`
   - Edge cases: empty series, single element, mixed types
   - Performance regression validation

## Integration Points

### Modified Files
1. `internal/native/register_series.go` - Refinement registration
2. `internal/native/series_polymorphic.go` - `seriesCopy()` modification
3. `internal/native/series_helpers.go` - Optional: add deep copy helpers
4. `internal/native/deep_copy.go` - New deep copy implementation
5. `test/contract/series_test.go` - Integration tests

### Dependencies
1. **Series interface**: No changes required
2. **Frame interface**: Uses existing `Clone()` method
3. **Object type**: Uses existing `GetAllFieldsWithProto()` and frame access
4. **Error handling**: Uses existing `verror.NewScriptError()` patterns

## Testing Strategy

### Test Categories
1. **Basic functionality**:
   ```viro
   a: [1 [2 3] 4]
   b: copy/deep a
   poke b 2 99
   ; a[2] should still be [2 3], not [99 3]
   ```

2. **Object support**:
   ```viro
   obj: object [x: 1 y: [2 3]]
   obj2: copy/deep obj
   poke obj2/y 1 99
   ; obj/y should still be [2 3]
   ```

3. **Cycle detection**:
   ```viro
   a: []
   append a a  ; Self-reference
   b: copy/deep a  ; Should terminate
   ```

4. **Refinement combinations**:
   ```viro
   copy/deep/part 2 [1 [2 3] [4 5]]  ; Should deep copy first 2 elements
   ```

5. **Immutable sharing**:
   ```viro
   str: "hello"
   a: [str str]
   b: copy/deep a
   ; b[1] and b[2] should reference same string object
   ```

### Validation Criteria
- All existing tests continue to pass
- Deep copy creates independent mutable containers
- Immutable values remain shared (reference equality)
- Cycles are detected and handled gracefully
- Memory usage scales appropriately with nesting depth

## Potential Challenges

### Technical Risks
1. **Cycle detection complexity**: Requires careful visited map management
2. **Object prototype semantics**: Need to decide whether to copy prototype chain (decision: share prototypes)
3. **Performance overhead**: Deep copying large nested structures could be expensive
4. **Memory management**: Visited map must not create reference cycles preventing GC

### Mitigation Strategies
- **Depth limit**: Hard cap recursion depth to prevent stack overflow
- **Lazy copying**: Consider future optimization for large structures
- **Testing**: Comprehensive test coverage for edge cases
- **Documentation**: Clear behavior specification for object prototype handling

## Viro Guidelines Reference

### Code Style Compliance
- **No comments in code**: Documentation belongs in package/docs only
- **Constructor functions**: Use `value.NewBlockValue()`, `value.ObjectVal()` not direct struct creation
- **Error handling**: Use `verror.NewScriptError()` with appropriate category/ID
- **Table-driven tests**: Use struct-based test cases with name/args/want/wantErr pattern

### Development Workflow
- **TDD mandatory**: Write tests first in `test/contract/`, then implement
- **Agent usage**: All code changes must use viro-coder agent via Task tool
- **Code review**: viro-reviewer agent mandatory after implementation
- **Backward compatibility**: Default behavior (shallow copy) must remain unchanged

### Performance Considerations
- **Minimize allocations**: Reuse visited map across recursion
- **Avoid reflection**: Use type switches not reflect package
- **Immutable optimization**: Early return for immutable types without visited map check

## Success Metrics
1. ✅ `copy/deep` works on all series types (block, string, binary)
2. ✅ Nested mutable containers are independently copyable
3. ✅ Cycles are detected and handled without infinite recursion
4. ✅ Immutable values are shared (reference equality preserved)
5. ✅ `--part` refinement combines correctly with `--deep`
6. ✅ Object deep copy maintains prototype chain semantics
7. ✅ All existing tests pass without modification