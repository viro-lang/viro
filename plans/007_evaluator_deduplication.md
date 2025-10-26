# Plan 007: Evaluator Code Deduplication

## Overview

Analyze and eliminate code redundancies in `internal/eval/evaluator.go` identified after Phase 3 of the evaluator simplification (Plan 006). The file is currently 930 lines and contains several opportunities for deduplication without sacrificing clarity.

## Analysis Summary

Current file size: 930 lines
Potential reduction: 30-40 lines with low-risk refactorings
Overall code quality: Good - most duplication is intentional for clarity

## Identified Redundancies

### 1. Frame Management Duplication (HIGH PRIORITY)

**Issue:** `pushFrame()` and `PushFrameContext()` have 100% identical implementations.

**Locations:**
- `pushFrame()` (lines 112-121) - private method
- `PushFrameContext()` (lines 249-258) - public method

**Current Code:**
```go
// pushFrame (lines 112-121)
func (e *Evaluator) pushFrame(f core.Frame) int {
	idx := f.GetIndex()
	if idx < 0 {
		idx = len(e.frameStore)
		e.frameStore = append(e.frameStore, f)
		f.SetIndex(idx)
	}
	e.Frames = append(e.Frames, f)
	return idx
}

// PushFrameContext (lines 249-258) - IDENTICAL CODE
func (e *Evaluator) PushFrameContext(f core.Frame) int {
	idx := f.GetIndex()
	if idx < 0 {
		idx = len(e.frameStore)
		e.frameStore = append(e.frameStore, f)
		f.SetIndex(idx)
	}
	e.Frames = append(e.Frames, f)
	return idx
}
```

**Impact:** 10 lines of duplicate code

**Recommendation:**
Remove `pushFrame()` entirely and replace all internal calls to `pushFrame()` with `PushFrameContext()`. The public method serves both internal and external use cases.

**Risk:** Low - simple find and replace
**Savings:** 10 lines (entire pushFrame function removed)

---

### 2. Index Bounds Checking (MEDIUM PRIORITY)

**Issue:** Repeated bounds checking pattern for block/string/binary indexing in path traversal.

**Locations:**
- Lines 806-809 (block indexing)
- Lines 814-817 (string indexing)
- Lines 821-824 (binary indexing)
- Lines 898-900 (block assignment)

**Current Code:**
```go
// Block indexing
if index < 1 || index > int64(len(block.Elements)) {
	return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, 
		[3]string{fmt.Sprintf("index %d out of range for block of length %d", index, len(block.Elements)), "", ""})
}

// String indexing
if index < 1 || index > int64(len(runes)) {
	return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, 
		[3]string{fmt.Sprintf("index %d out of range for string of length %d", index, len(runes)), "", ""})
}

// Binary indexing
if index < 1 || index > int64(bin.Length()) {
	return nil, verror.NewScriptError(verror.ErrIDIndexOutOfRange, 
		[3]string{fmt.Sprintf("index %d out of range for binary of length %d", index, bin.Length()), "", ""})
}
```

**Impact:** ~30 lines of similar code

**Recommendation:**
```go
// checkIndexBounds validates that an index is within bounds (1-based indexing).
// Returns nil if valid, error if out of bounds.
func checkIndexBounds(index, length int64, typeName string) error {
	if index < 1 || index > length {
		return verror.NewScriptError(verror.ErrIDIndexOutOfRange, 
			[3]string{fmt.Sprintf("index %d out of range for %s of length %d", index, typeName, length), "", ""})
	}
	return nil
}

// Usage:
if err := checkIndexBounds(index, int64(len(block.Elements)), "block"); err != nil {
	return nil, err
}

if err := checkIndexBounds(index, int64(len(runes)), "string"); err != nil {
	return nil, err
}

if err := checkIndexBounds(index, int64(bin.Length()), "binary"); err != nil {
	return nil, err
}
```

**Risk:** Low - simple extraction
**Savings:** ~20 lines

---

### 3. Refinement Error Creation (LOW PRIORITY)

**Issue:** Similar error creation patterns in `readRefinements()`.

**Locations:**
- Lines 630-633 (unknown refinement)
- Lines 637-640 (duplicate refinement)
- Lines 647-649 (missing value)

**Current Code:**
```go
// Unknown refinement
return pos, verror.NewScriptError(
	verror.ErrIDInvalidOperation,
	[3]string{fmt.Sprintf("Unknown refinement: --%s", refName), "", ""},
)

// Duplicate refinement
return pos, verror.NewScriptError(
	verror.ErrIDInvalidOperation,
	[3]string{fmt.Sprintf("Duplicate refinement: --%s", refName), "", ""},
)

// Missing value
return pos, verror.NewScriptError(
	verror.ErrIDInvalidOperation,
	[3]string{fmt.Sprintf("Refinement --%s requires a value", refName), "", ""},
)
```

**Impact:** Moderate - error creation verbosity

**Recommendation:**
```go
// refinementError creates a script error for refinement-related issues.
func refinementError(kind, refName string) error {
	var msg string
	switch kind {
	case "unknown":
		msg = fmt.Sprintf("Unknown refinement: --%s", refName)
	case "duplicate":
		msg = fmt.Sprintf("Duplicate refinement: --%s", refName)
	case "missing-value":
		msg = fmt.Sprintf("Refinement --%s requires a value", refName)
	}
	return verror.NewScriptError(verror.ErrIDInvalidOperation, [3]string{msg, "", ""})
}

// Usage:
return pos, refinementError("unknown", refName)
return pos, refinementError("duplicate", refName)
return pos, refinementError("missing-value", refName)
```

**Risk:** Low
**Savings:** ~10 lines
**Note:** This might reduce clarity for error handling. Consider if the abstraction is worth it.

---

## Not Recommended for Deduplication

### I/O Writer/Reader Setters

**Locations:** `SetOutputWriter()`, `SetErrorWriter()`, `SetInputReader()`

**Reason:** Different types (io.Writer vs io.Reader), different defaults (Stdout vs Stderr vs Stdin)
**Decision:** Keep as is - clarity > DRY for this case

### Type Assertions

**Locations:** Multiple locations for `.(string)` and `.(int64)` assertions

**Reason:** Would require complex generics, reduces readability
**Decision:** Keep as is - Go's type system doesn't lend itself to abstraction here

### Position Advancement

**Locations:** Various `position + 1` and `position++` patterns

**Reason:** Core execution model, different contexts need different patterns
**Decision:** Keep as is - this is not redundant, just similar

---

## Implementation Plan

### Phase 1: Frame Management Consolidation

**Steps:**
1. Find all calls to `pushFrame()` in `evaluator.go`
2. Replace all calls to `pushFrame()` with `PushFrameContext()`
3. Delete the `pushFrame()` function entirely
4. Run tests to verify no behavioral changes

**Expected Impact:**
- Lines reduced: 10
- Risk: Very low
- Test coverage: Existing tests should pass

### Phase 2: Bounds Check Extraction

**Steps:**
1. Create `checkIndexBounds()` helper function
2. Replace 4 instances of bounds checking with helper calls
3. Ensure error messages remain consistent
4. Run tests

**Expected Impact:**
- Lines reduced: ~20
- Risk: Low
- Test coverage: Existing path tests should cover this

### Phase 3: Refinement Error Helper (Optional)

**Steps:**
1. Create `refinementError()` helper
2. Replace 3 error creation sites
3. Run tests

**Expected Impact:**
- Lines reduced: ~10
- Risk: Low
- Test coverage: Existing refinement tests should cover this

---

## Success Criteria

- [ ] All tests pass (especially contract tests and integration tests)
- [ ] File reduced by 30-40 lines
- [ ] No behavioral changes
- [ ] Code remains readable and maintainable
- [ ] No performance regression

---

## Expected Outcomes

### Before
- File size: 930 lines
- Duplicate frame management: 10 lines
- Repeated bounds checking: ~30 lines

### After
- File size: ~890-900 lines
- Single frame management implementation
- Centralized bounds checking
- Optional: Centralized refinement errors

### Code Quality Improvements
- Single source of truth for frame registration
- Consistent bounds checking logic
- More maintainable error handling
- Reduced cognitive load when reading code

---

## Risks and Mitigations

### Risk 1: Test Failures
**Mitigation:** Run full test suite after each phase
**Rollback:** Git allows easy revert if needed

### Risk 2: Performance Impact
**Mitigation:** These are helper functions on non-critical paths
**Verification:** Bounds checking happens during path traversal, not tight loops

### Risk 3: Reduced Error Message Clarity
**Mitigation:** Ensure error messages remain identical after refactoring
**Verification:** Check error message tests

---

## Timeline Estimate

- Phase 1 (Frame consolidation): 30 minutes
- Phase 2 (Bounds checking): 1 hour
- Phase 3 (Refinement errors): 30 minutes (optional)

**Total:** 1.5-2 hours

---

## References

- Source file: `internal/eval/evaluator.go` (930 lines)
- Related plan: `plans/006_evaluator_simplification.md` (Phase 3 completed)
- Test files: `test/contract/*_test.go`, `test/integration/*_test.go`
