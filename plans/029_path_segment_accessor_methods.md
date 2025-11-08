# Implementation Plan: PathSegment Accessor Methods

**Branch**: `029-path-segment-accessors` | **Date**: 2025-11-08 | **Issue**: [#55](https://github.com/marcin-radoszewski/viro/issues/55)

## Feature Summary

Add type-safe accessor methods to `PathSegment` to eliminate scattered unsafe type assertions throughout the codebase. This improves code safety, readability, and maintainability by providing a consistent API pattern (Is*/As*) for accessing PathSegment values.

## Research Findings

### Current State Analysis

**PathSegment Structure** (`internal/value/path.go`):
```go
type PathSegment struct {
    Type  PathSegmentType
    Value any
}

type PathSegmentType int

const (
    PathSegmentWord PathSegmentType = iota  // Value: string
    PathSegmentIndex                         // Value: int64
    PathSegmentEval                          // Value: *BlockValue
)
```

**Type Assertion Locations Identified**:

1. **internal/value/path.go** (3 instances):
   - Line 55: `seg.Value.(string)` - UNSAFE, no ok check
   - Line 57: `seg.Value.(int64)` - UNSAFE, no ok check
   - Line 59: `block, ok := seg.Value.(*BlockValue)` - safe, but should use accessor

2. **internal/eval/evaluator.go** (6 instances):
   - Line 912: `block, ok := seg.Value.(*value.BlockValue)` - safe assertion in `materializeSegment`
   - Line 953: `wordStr, ok := firstSeg.Value.(string)` - safe assertion in `resolvePathBase`
   - Line 964: `num, ok := firstSeg.Value.(int64)` - safe assertion in `resolvePathBase`
   - Line 991: `fieldName, ok := seg.Value.(string)` - safe assertion in `traverseWordSegment`
   - Line 1006: `index, ok := seg.Value.(int64)` - safe assertion in `traverseIndexSegment`
   - Line 1194: `index, ok := finalSeg.Value.(int64)` - safe assertion in `assignToIndexTarget`
   - Line 1224: `fieldName, ok := finalSeg.Value.(string)` - safe assertion in `assignToWordTarget`

3. **internal/parse/path_test.go** (2 instances):
   - Line 108: `block, ok := seg.Value.(*value.BlockValue)` - test code, safe assertion
   - Line 139: `block, ok := seg.Value.(*value.BlockValue)` - test code, safe assertion

**Total**: 11 type assertions, 2 of which are UNSAFE (panic risk)

### Existing Patterns in Viro

Viro uses consistent accessor patterns throughout the value system:
- `AsBlockValue(v core.Value) (*BlockValue, bool)` - global type assertion helper
- `AsObject(v core.Value) (*Object, bool)` - global type assertion helper
- `AsIntValue(v core.Value) (int64, bool)` - global type assertion helper

The proposed `PathSegment` accessors follow this pattern but as methods on the segment itself.

### Architecture Context

**Path Evaluation Pipeline**:
1. Parser creates PathSegment instances during tokenization
2. Evaluator materializes eval segments (converts PathSegmentEval → Word/Index)
3. Evaluator traverses segments to resolve path values
4. Assignment operations modify containers through path segments

**Type Invariants** (currently enforced by convention, not compile-time):
- `PathSegmentWord` → Value MUST be `string`
- `PathSegmentIndex` → Value MUST be `int64`
- `PathSegmentEval` → Value MUST be `*BlockValue`

These invariants are violated by unsafe assertions in `renderPathSegments`.

## Architecture Overview

### Design Approach

Add six accessor methods to `PathSegment` following Viro's Is*/As* pattern:

```go
// Type checkers
func (seg PathSegment) IsWord() bool
func (seg PathSegment) IsIndex() bool
func (seg PathSegment) IsEval() bool

// Value extractors (return zero value + false if wrong type)
func (seg PathSegment) AsWord() (string, bool)
func (seg PathSegment) AsIndex() (int64, bool)
func (seg PathSegment) AsEvalBlock() (*BlockValue, bool)
```

**Design Rationale**:
- **Method receivers**: Use value receivers (not pointers) to match PathSegment usage patterns
- **Dual API**: Is* for type checking, As* for extraction (reduces assertion boilerplate)
- **Consistent naming**: AsEvalBlock (not AsEval) clarifies return type
- **Safe by default**: All As* methods return (value, bool) tuple

### Integration Strategy

Replace all type assertions following this migration path:

**Pattern 1: Safe assertion with ok check**
```go
// Before
fieldName, ok := seg.Value.(string)
if !ok {
    return error
}

// After
fieldName, ok := seg.AsWord()
if !ok {
    return error
}
```

**Pattern 2: Unsafe assertion (panic risk)**
```go
// Before
result += seg.Value.(string)

// After
if word, ok := seg.AsWord(); ok {
    result += word
}
// Note: Error handling required for previously unsafe code
```

**Pattern 3: Type switch replacement**
```go
// Before
switch seg.Type {
case PathSegmentWord:
    name := seg.Value.(string)  // Trusts Type field
    
// After
switch seg.Type {
case PathSegmentWord:
    name, _ := seg.AsWord()  // Double verification
```

## Implementation Roadmap

### Phase 1: Add Accessor Methods

**File**: `internal/value/path.go`

**Task**: Implement six accessor methods on PathSegment

**Guidance for Implementation**:
1. Add methods immediately after the `PathSegmentType.String()` method (line 38)
2. Use value receivers: `func (seg PathSegment) MethodName()`
3. Each Is* method checks `seg.Type == PathSegment*` constant
4. Each As* method:
   - First checks type using corresponding Is* method
   - Then performs type assertion with ok check
   - Returns zero value + false on mismatch
5. Follow exact API signature from proposed interface (issue #55)

**Implementation Pattern**:
```go
func (seg PathSegment) IsWord() bool {
    return seg.Type == PathSegmentWord
}

func (seg PathSegment) AsWord() (string, bool) {
    if !seg.IsWord() {
        return "", false
    }
    str, ok := seg.Value.(string)
    return str, ok
}
```

**Validation Checkpoint**: 
- All six methods compile without errors
- Methods are exported (capitalized names)
- Return types match proposed API exactly

### Phase 2: Update renderPathSegments Function

**File**: `internal/value/path.go`

**Function**: `renderPathSegments` (lines 47-68)

**Task**: Replace unsafe type assertions with accessor methods

**Current Code Issues**:
- Lines 55, 57: Unsafe assertions (no ok check) - WILL PANIC on type mismatch
- Line 59: Safe assertion but should use new API

**Guidance for Implementation**:
1. For PathSegmentWord case (line 55):
   - Replace `seg.Value.(string)` with `seg.AsWord()`
   - Add error handling: if !ok, append error marker or type name
   
2. For PathSegmentIndex case (line 57):
   - Replace `seg.Value.(int64)` with `seg.AsIndex()`
   - Add error handling: if !ok, append error marker or type name
   
3. For PathSegmentEval case (line 59):
   - Replace `block, ok := seg.Value.(*BlockValue)` with `block, ok := seg.AsEvalBlock()`
   - Keep existing ok check logic

**Decision Point**: How to handle accessor failures in rendering?
- **Option A**: Append placeholder like `<invalid-word>` or `<invalid-index>`
- **Option B**: Log error and skip segment (dangerous - breaks path representation)
- **Recommended**: Option A - maintain string output integrity

**Validation Checkpoint**:
- `renderPathSegments` compiles without errors
- All existing path tests pass: `go test ./internal/value -run TestPath`
- No unsafe type assertions remain in function

### Phase 3: Update evaluator.go - materializeSegment

**File**: `internal/eval/evaluator.go`

**Function**: `materializeSegment` (lines 907-948)

**Task**: Replace type assertion on line 912

**Current Code**:
```go
block, ok := seg.Value.(*value.BlockValue)
if !ok {
    return value.PathSegment{}, verror.NewInternalError(...)
}
```

**Guidance for Implementation**:
1. Replace with `block, ok := seg.AsEvalBlock()`
2. Keep existing error handling unchanged
3. Verify return type matches (no changes needed)

**Validation Checkpoint**:
- Function compiles
- Path evaluation tests pass: `go test ./test/contract -run TestPath`

### Phase 4: Update evaluator.go - resolvePathBase

**File**: `internal/eval/evaluator.go`

**Function**: `resolvePathBase` (lines 950-972)

**Task**: Replace two type assertions (lines 953, 964)

**Current Code**:
```go
case value.PathSegmentWord:
    wordStr, ok := firstSeg.Value.(string)
    
case value.PathSegmentIndex:
    num, ok := firstSeg.Value.(int64)
```

**Guidance for Implementation**:
1. Line 953: Replace with `wordStr, ok := firstSeg.AsWord()`
2. Line 964: Replace with `num, ok := firstSeg.AsIndex()`
3. Keep all error handling unchanged

**Validation Checkpoint**:
- Function compiles
- Path base resolution tests pass

### Phase 5: Update evaluator.go - traverseWordSegment

**File**: `internal/eval/evaluator.go`

**Function**: `traverseWordSegment` (lines 974-1003)

**Task**: Replace type assertion on line 991

**Current Code**:
```go
fieldName, ok := seg.Value.(string)
if !ok {
    return verror.NewInternalError("word segment does not contain string", [3]string{})
}
```

**Guidance for Implementation**:
1. Replace with `fieldName, ok := seg.AsWord()`
2. Keep error handling unchanged

**Validation Checkpoint**:
- Function compiles
- Object field access tests pass: `go test ./test/contract -run TestObject`

### Phase 6: Update evaluator.go - traverseIndexSegment

**File**: `internal/eval/evaluator.go`

**Function**: `traverseIndexSegment` (lines 1005-1055)

**Task**: Replace type assertion on line 1006

**Current Code**:
```go
index, ok := seg.Value.(int64)
if !ok {
    return verror.NewInternalError("index segment does not contain int64", [3]string{})
}
```

**Guidance for Implementation**:
1. Replace with `index, ok := seg.AsIndex()`
2. Keep error handling unchanged

**Validation Checkpoint**:
- Function compiles
- Block/string/binary indexing tests pass

### Phase 7: Update evaluator.go - assignToIndexTarget

**File**: `internal/eval/evaluator.go`

**Function**: `assignToIndexTarget` (lines 1193-1221)

**Task**: Replace type assertion on line 1194

**Guidance for Implementation**:
1. Replace `index, ok := finalSeg.Value.(int64)` with `index, ok := finalSeg.AsIndex()`
2. Keep error handling unchanged

**Validation Checkpoint**:
- Function compiles
- Set-path tests pass: `go test ./test/contract -run TestSetPath`

### Phase 8: Update evaluator.go - assignToWordTarget

**File**: `internal/eval/evaluator.go`

**Function**: `assignToWordTarget` (lines 1223-1247)

**Task**: Replace type assertion on line 1224

**Guidance for Implementation**:
1. Replace `fieldName, ok := finalSeg.Value.(string)` with `fieldName, ok := seg.AsWord()`
2. Note: Variable name is `finalSeg` but function parameter is `seg` - use correct variable

**Validation Checkpoint**:
- Function compiles
- Object field assignment tests pass

### Phase 9: Update Test Files (Optional Enhancement)

**File**: `internal/parse/path_test.go`

**Lines**: 108, 139

**Task**: Replace safe assertions in test code

**Guidance for Implementation**:
- This is test code, so change is optional but recommended for consistency
- Replace `block, ok := seg.Value.(*value.BlockValue)` with `block, ok := seg.AsEvalBlock()`
- Maintain existing test logic

**Validation Checkpoint**:
- All parse tests pass: `go test ./internal/parse -run TestPath`

### Phase 10: Final Verification

**Tasks**:
1. Run full test suite: `go test ./...`
2. Verify no remaining unsafe assertions: `rg "seg\.Value\.\(" --type go` (should return 0 results)
3. Build CLI: `make build`
4. Manual smoke test with path expressions:
   ```bash
   ./viro -c "obj: make object! [name: \"test\"]  obj.name"
   ./viro -c "data: [10 20 30]  data.2"
   ./viro -c "data: [10 20 30]  data.2: 99  data"
   ```

**Success Criteria**:
- Zero unsafe type assertions on PathSegment.Value
- All tests pass (no regressions)
- CLI builds and executes path expressions correctly
- Code is more maintainable and type-safe

## Integration Points

### With Parser (`internal/parse/semantic_parser.go`)
- Parser creates PathSegment instances during path tokenization
- Accessors don't change parser behavior (creation side unchanged)
- Parser tests validate correct Type/Value assignment

### With Evaluator (`internal/eval/evaluator.go`)
- Evaluator is primary consumer of PathSegment
- All path traversal and assignment logic uses accessors
- Error messages remain unchanged (internal errors still trigger)

### With Value System (`internal/value/`)
- PathSegment follows existing Value accessor patterns
- Consistent with AsBlockValue, AsObject, AsIntValue helpers
- No changes to PathExpression, GetPathExpression, SetPathExpression types

## Testing Strategy

### Existing Test Coverage

**Contract Tests** (`test/contract/path_eval_test.go`):
- Path evaluation with word segments
- Path evaluation with index segments
- Path evaluation with eval segments
- Set-path assignment operations

**Integration Tests** (`test/integration/`):
- End-to-end path expression execution
- Error handling for invalid paths

**Unit Tests** (`internal/value/path_test.go`):
- PathSegment type string representation
- Path molding and formatting
- Empty path edge cases

**Parser Tests** (`internal/parse/path_test.go`):
- Path tokenization with all segment types
- Nested eval segments
- Invalid path syntax rejection

### No New Tests Required

**Rationale**: This is a refactoring change that improves internal API safety without altering external behavior. Existing comprehensive test suite validates correctness.

**Regression Detection**: If any type assertion was previously masking a bug (wrong Type/Value pairing), accessor methods will expose it as test failures.

### Manual Testing Checklist

After implementation, verify these scenarios work:

1. **Word segments**: `obj.field.nested`
2. **Index segments**: `block.1.2`
3. **Eval segments**: `data.(idx).value`
4. **Set-path**: `obj.field: newval`
5. **Mixed segments**: `data.1.(key).3`
6. **Error cases**: Invalid index, missing field, none traversal

## Potential Challenges

### Challenge 1: Unsafe Assertion Discovery

**Issue**: Lines 55 and 57 in `renderPathSegments` use unsafe assertions. If a PathSegment has mismatched Type/Value, these will panic in production.

**Mitigation**:
- Replace with accessor methods + error handling
- Add defensive placeholder rendering for invalid segments
- Existing tests may not cover this edge case (no Type/Value mismatch tests)

**Decision Point**: Should we add validation tests for Type/Value invariants?
- **Recommended**: Add test in Phase 10 that attempts to create invalid PathSegment
- **Location**: `internal/value/path_test.go`
- **Test**: Verify accessor returns false for mismatched Type/Value

### Challenge 2: Error Message Changes

**Issue**: Changing from panic (unsafe assertion) to graceful error handling might alter error messages.

**Mitigation**:
- Keep existing error handling logic unchanged
- Only replace assertion syntax, not error paths
- Internal errors remain internal errors

**Impact**: None - all safe assertions already use same error handling pattern

### Challenge 3: Performance Impact

**Issue**: Adding method calls instead of direct field access might impact performance.

**Analysis**:
- PathSegment accessors are trivial (type check + assertion)
- Go compiler likely inlines these methods
- Path evaluation is not performance-critical (dominated by I/O, computation)

**Mitigation**:
- No optimization needed
- If performance regression detected, run benchmarks: `go test -bench=BenchmarkPath`

**Validation**: Existing benchmarks will detect any significant regression

## Viro Guidelines Reference

### Code Style Compliance

✅ **No Comments Rule**: Accessor methods are self-documenting through names
✅ **Constructor Pattern**: Not applicable (methods, not constructors)
✅ **Index-based Refs**: Not applicable (no stack/frame usage)
✅ **Table-driven Tests**: Existing tests already use this pattern
✅ **Error Handling**: Uses verror.NewInternalError for type mismatches
✅ **Import Organization**: No new imports required
✅ **Naming Convention**: Is*/As* pattern matches Viro style (AsBlockValue, AsObject, etc.)

### Architectural Alignment

✅ **Type-based Dispatch**: Accessors strengthen type system safety
✅ **Value Constructor Usage**: PathSegment uses direct struct literal (allowed for internal types)
✅ **Error Categories**: Maintains existing internal error usage
✅ **Frame System**: Not applicable
✅ **No Pointers in Stack**: Not applicable

## Open Questions & Assumptions

### Assumptions Made

1. **Type/Value Invariant**: PathSegment instances are always created with matching Type and Value fields
   - **Risk**: If violated, accessors return false (safe failure)
   - **Validation**: Existing tests + accessor implementation will expose violations

2. **Error Handling Preservation**: All existing error messages and codes remain unchanged
   - **Risk**: None - only assertion syntax changes, not error paths
   - **Validation**: Grep for verror usage confirms same error IDs used

3. **No Public API Impact**: PathSegment is internal package type
   - **Risk**: None - changes confined to internal/ packages
   - **Validation**: No external users of PathSegment exist

### Clarifications Needed

❓ **Should renderPathSegments panic or degrade gracefully on invalid segments?**
- **Current Behavior**: Panics on type mismatch (unsafe assertions)
- **Proposed Behavior**: Append placeholder like `<invalid>`
- **Recommendation**: Graceful degradation for better debugging

❓ **Should we add constructor functions to enforce invariants?**
- **Example**: `NewWordSegment(s string) PathSegment`
- **Benefit**: Compile-time safety for segment creation
- **Cost**: Larger refactoring scope
- **Recommendation**: Out of scope for this issue; file follow-up issue

## Implementation Checklist

- [ ] Phase 1: Add six accessor methods to PathSegment
- [ ] Phase 2: Update renderPathSegments (fix unsafe assertions!)
- [ ] Phase 3: Update materializeSegment
- [ ] Phase 4: Update resolvePathBase
- [ ] Phase 5: Update traverseWordSegment
- [ ] Phase 6: Update traverseIndexSegment
- [ ] Phase 7: Update assignToIndexTarget
- [ ] Phase 8: Update assignToWordTarget
- [ ] Phase 9: Update test files (optional)
- [ ] Phase 10: Final verification and testing
- [ ] Verify zero unsafe assertions remain: `rg "seg\.Value\.\("`
- [ ] Run full test suite: `go test ./...`
- [ ] Build CLI: `make build`
- [ ] Manual smoke tests with path expressions
- [ ] Update issue #55 with results

## Success Metrics

**Code Quality**:
- ✅ Zero unsafe type assertions on `seg.Value`
- ✅ Consistent accessor API pattern
- ✅ No code comments (self-documenting)

**Correctness**:
- ✅ All existing tests pass (no regressions)
- ✅ CLI builds successfully
- ✅ Path expressions execute correctly

**Maintainability**:
- ✅ Single source of truth for type assertions
- ✅ Future code uses accessors (enforced by code review)
- ✅ Type safety improvements (panic → graceful error)

## Estimated Effort

- **Phase 1**: 15 minutes (straightforward method implementation)
- **Phase 2**: 20 minutes (requires error handling decisions)
- **Phases 3-8**: 30 minutes (mechanical replacement)
- **Phase 9**: 10 minutes (optional test updates)
- **Phase 10**: 20 minutes (verification and smoke testing)

**Total**: ~95 minutes for complete implementation and validation

**Complexity**: LOW - Pure refactoring with comprehensive test coverage
