# Implementation Tasks: Natives Within Frame

**Feature Branch**: `003-natives-within-frame`
**Date**: 2025-10-12
**Status**: Ready for Implementation

## Overview

This document provides a detailed, dependency-ordered task breakdown for implementing the "Natives Within Frame" feature. Tasks are organized by user story to enable independent implementation and testing. The feature eliminates the special-case native registry and moves native functions into the root frame, enabling standard lexical scoping.

**Total Tasks**: 32
**Estimated Duration**: 5-7 days
**Testing Approach**: TDD (Test-First) - Contract tests written before implementation per CLAUDE.md principles

---

## Task Organization

Tasks are grouped into phases:
1. **Phase 1: Setup** - Project-level configuration
2. **Phase 2: Foundational** - Blocking prerequisites for all user stories
3. **Phase 3-5: User Stories** - Implementation by story (P1 → P2)
4. **Phase 6: Polish** - Cross-cutting concerns and optimization

**Labels**:
- `[US1]`, `[US2]`, `[US3]` - User story association
- `[P]` - Parallelizable with other `[P]` tasks
- `[TEST]` - Test task (TDD)
- `[IMPL]` - Implementation task

---

## Phase 1: Setup & Preparation

### ✅ T001: [SETUP] Verify existing test suite baseline

**Goal**: Establish baseline test results before any changes

**Steps**:
1. Run full test suite: `go test ./...`
2. Record pass/fail counts and execution time
3. Document any existing flaky tests
4. Create baseline performance metrics for word lookup benchmarks

**Files**: None (verification only)

**Acceptance**:
- All existing tests documented
- Baseline metrics recorded
- Zero test modifications made

**Duration**: 15 min

**Status**: ✅ COMPLETED - All tests pass (baseline established)

---

## Phase 2: Foundational Tasks (Blocking Prerequisites)

These tasks must complete before ANY user story implementation can begin.

### ✅ T002: [FOUNDATION] Create native registration category files (Math)

**Goal**: Extract math native registrations from registry.go into dedicated file

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/native/register_math.go`
2. Implement `RegisterMathNatives(rootFrame *frame.Frame)` function
3. Extract Groups 1-4 from registry.go (20 functions): `+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `=`, `<>`, `and`, `or`, `not`, `abs`, `min`, `max`, `sqrt`, `power`, `floor`, `ceiling`
4. Add panic-on-nil and duplicate-check validation
5. Preserve all metadata (Infix flags, Doc annotations)

**Files**:
- NEW: `internal/native/register_math.go` (~400 LOC)

**Acceptance**:
- File compiles without errors
- All 20 math natives registered
- Validation logic present (panic on duplicate/nil)
- Metadata preserved (Infix, Doc)

**Duration**: 45 min

**Reference**: contracts/frame.md, research.md D1

**Status**: ✅ COMPLETED - 29 math natives registered (register_math.go created, 671 LOC)

---

### ✅ T003: [FOUNDATION] [P] Create native registration category files (Series)

**Goal**: Extract series native registrations into dedicated file

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/native/register_series.go`
2. Implement `RegisterSeriesNatives(rootFrame *frame.Frame)`
3. Extract Group 5 from registry.go (5 functions): `first`, `last`, `append`, `insert`, `length?`
4. Add validation logic

**Files**:
- NEW: `internal/native/register_series.go` (~100 LOC)

**Acceptance**:
- File compiles
- 5 series natives registered
- Validation present

**Duration**: 20 min

**Parallel with**: T004, T005, T006, T007

**Status**: ✅ COMPLETED - 5 series natives registered (register_series.go created, 146 LOC)

---

### ✅ T004: [FOUNDATION] [P] Create native registration category files (Data)

**Goal**: Extract data/object native registrations

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/native/register_data.go`
2. Implement `RegisterDataNatives(rootFrame *frame.Frame)`
3. Extract Groups 6-7 (8 functions): `set`, `get`, `type?`, `object`, `clone`, `copy`, `find`, `index?`

**Files**:
- NEW: `internal/native/register_data.go` (~150 LOC)

**Duration**: 30 min

**Parallel with**: T003, T005, T006, T007

**Status**: ✅ COMPLETED - 8 data natives registered (register_data.go created, 321 LOC)

---

### ✅ T005: [FOUNDATION] [P] Create native registration category files (I/O)

**Goal**: Extract I/O and port native registrations

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/native/register_io.go`
2. Implement `RegisterIONatives(rootFrame *frame.Frame)`
3. Extract Groups 8-9 (10 functions): `print`, `input`, `open`, `read`, `write`, `close`, `exists?`, `delete`, `rename`, `dir`

**Files**:
- NEW: `internal/native/register_io.go` (~200 LOC)

**Duration**: 30 min

**Parallel with**: T003, T004, T006, T007

**Status**: ✅ COMPLETED - 10 I/O natives registered (register_io.go created, 232 LOC)

---

### ✅ T006: [FOUNDATION] [P] Create native registration category files (Control)

**Goal**: Extract control flow native registrations

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/native/register_control.go`
2. Implement `RegisterControlNatives(rootFrame *frame.Frame)`
3. Extract Groups 10-11 (5 functions): `if`, `when`, `loop`, `while`, `fn`

**Files**:
- NEW: `internal/native/register_control.go` (~150 LOC)

**Duration**: 25 min

**Parallel with**: T003, T004, T005, T007

**Status**: ✅ COMPLETED - 5 control natives registered (register_control.go created, 209 LOC)

---

### ✅ T007: [FOUNDATION] [P] Create native registration category files (Help)

**Goal**: Extract help and observability native registrations

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/native/register_help.go`
2. Implement `RegisterHelpNatives(rootFrame *frame.Frame)`
3. Extract Groups 12-13 (11 functions): `help`, `info`, `trace`, `debug`, `breakpoint`, `step`, `reflect`, `words`, `values`, `stats`, `gc`

**Files**:
- NEW: `internal/native/register_help.go` (~250 LOC)

**Duration**: 35 min

**Parallel with**: T003, T004, T005, T006

**Status**: ✅ COMPLETED - 11 help natives registered (register_help.go created, 367 LOC)

---

### ✅ T008: [FOUNDATION] Modify NewEvaluator to register natives in root frame

**Goal**: Add native registration calls to evaluator construction

**Steps**:
1. Modify `/Users/marcin.radoszewski/dev-allegro/viro/internal/eval/evaluator.go`
2. In `NewEvaluator()`, change root frame creation to use `NewFrameWithCapacity(FrameClosure, -1, 80)`
3. After creating evaluator, add sequential registration calls:
   ```go
   native.RegisterMathNatives(global)
   native.RegisterSeriesNatives(global)
   native.RegisterDataNatives(global)
   native.RegisterIONatives(global)
   native.RegisterControlNatives(global)
   native.RegisterHelpNatives(global)
   ```
4. Ensure panic behavior propagates (no error handling - fail fast)
5. Keep `native.Registry` population for backward compatibility (Phase 1 strategy)

**Files**:
- MODIFIED: `internal/eval/evaluator.go` (NewEvaluator function, ~20 new lines)

**Acceptance**:
- Root frame initialized with capacity 80
- All 6 registration functions called sequentially
- Panic propagates on registration failure
- Registry still populated (dual mode)

**Duration**: 30 min

**Dependencies**: T002, T003, T004, T005, T006, T007

**Reference**: contracts/evaluator.md, data-model.md state machine

**Status**: ✅ COMPLETED - NewEvaluator modified to call all 6 registration functions

---

### ✅ **CHECKPOINT 1**: Foundation Complete

**Verification**:
- 6 registration files created and compile
- NewEvaluator calls all registration functions
- Root frame populated with natives
- Native registry still functional (backward compat)
- No existing tests broken

**Command**: `go test ./internal/eval -run TestNewEvaluator`

---

## Phase 3: User Story 1 - Define Functions with Names Matching Natives (P1)

**Goal**: Enable developers to use native names for local variables/refinements without collisions

**Independent Test**: Function with `--debug` refinement works without conflicting with native `debug`

### T009: [US1] [TEST] Write contract test for refinement name collision resolution

**Goal**: Test that refinements can use native names without errors

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/test/contract/native_scoping_test.go`
2. Write `TestRefinementWithNativeName`:
   - Define function with `--debug` refinement parameter
   - Call function with and without refinement
   - Verify both work correctly
   - Verify no collision with native `debug` function
3. Write `TestLocalVariableWithNativeName`:
   - Define function with local variable named `type`
   - Verify local variable accessed, not native `type?`

**Files**:
- NEW: `test/contract/native_scoping_test.go` (~100 LOC)

**Acceptance**:
- Test fails initially (expected - TDD)
- Clear failure message indicates special-case registry lookup

**Duration**: 30 min

**Reference**: spec.md User Story 1, quickstart.md Test 1

---

### T010: [US1] [TEST] Write contract test for nested scope shadowing

**Goal**: Test that nested scopes follow lexical scoping rules

**Steps**:
1. Add to `test/contract/native_scoping_test.go`
2. Write `TestNestedScopeShadowing`:
   - Create 3-level nested scopes
   - Define variable with native name in each level
   - Verify innermost binding wins at each level
   - Pop frames and verify outer bindings visible again

**Files**:
- MODIFIED: `test/contract/native_scoping_test.go` (+50 LOC)

**Acceptance**:
- Test fails with registry-first lookup behavior
- Clear assertion failures on shadowing expectations

**Duration**: 20 min

---

### T011: [US1] [IMPL] Remove native.Lookup() from evalWord

**Goal**: Eliminate special-case native check in word evaluation

**Steps**:
1. Modify `/Users/marcin.radoszewski/dev-allegro/viro/internal/eval/evaluator.go`
2. In `evalWord()` function (line ~745):
   - Remove the native registry check:
     ```go
     // DELETE THIS:
     if _, ok := native.Lookup(wordStr); ok {
         return val, nil
     }
     ```
   - Keep only the frame chain lookup: `e.Lookup(wordStr)`
3. Update logic to handle function values from frames

**Files**:
- MODIFIED: `internal/eval/evaluator.go` (evalWord function, -5 lines)

**Acceptance**:
- Native registry check removed
- Frame lookup used exclusively
- Tests T009, T010 now pass

**Duration**: 15 min

**Dependencies**: T008, T009, T010

**Reference**: contracts/lookup.md, data-model.md Word Lookup Resolution

---

### T012: [US1] [IMPL] Remove native.Lookup() from evalGetWord

**Goal**: Unify get-word evaluation to use frame chain

**Steps**:
1. Modify `/Users/marcin.radoszewski/dev-allegro/viro/internal/eval/evaluator.go`
2. In `evalGetWord()` function (line ~828):
   - Remove native registry check (similar to evalWord)
   - Use `e.Lookup()` exclusively

**Files**:
- MODIFIED: `internal/eval/evaluator.go` (evalGetWord function, -4 lines)

**Acceptance**:
- Get-word uses frame chain only
- Metadata retrieval still works for natives

**Duration**: 10 min

**Dependencies**: T011

---

### T013: [US1] [IMPL] Remove native.Lookup() from evaluateWithFunctionCall

**Goal**: Unify function call dispatch to use frame chain

**Steps**:
1. Modify `/Users/marcin.radoszewski/dev-allegro/viro/internal/eval/evaluator.go`
2. In `evaluateWithFunctionCall()` (line ~512):
   - Remove lines 522-525 (native registry check)
   - Consolidate to single `e.Lookup()` path for all words
   - Logic: if lookup returns function → invoke, else return value

**Files**:
- MODIFIED: `internal/eval/evaluator.go` (evaluateWithFunctionCall, -8 lines, +2 lines)

**Acceptance**:
- Single unified lookup path
- Native and user-defined functions handled identically
- Infix operators still work (natives and user functions)

**Duration**: 20 min

**Dependencies**: T012

**Reference**: contracts/lookup.md Modified Call Sites

---

### T014: [US1] [TEST] Verify all existing tests still pass

**Goal**: Confirm backward compatibility maintained

**Steps**:
1. Run full test suite: `go test ./...`
2. Verify no failures in existing tests
3. Confirm new tests (T009, T010) pass
4. Check performance hasn't regressed

**Files**: None (verification only)

**Acceptance**:
- All pre-existing tests pass (from T001 baseline)
- New native scoping tests pass
- Zero test modifications required (SC-002)

**Duration**: 10 min

**Dependencies**: T013

---

### ✅ **CHECKPOINT 2**: User Story 1 Complete

**Verification**:
- ✅ Refinement parameters can use native names (`--debug`, `--type`)
- ✅ Local variables can use native names without conflicts
- ✅ Nested scoping follows lexical rules (innermost wins)
- ✅ All existing tests pass without modification
- ✅ SC-001: Functions with `--debug` refinement work correctly

**Deliverable**: Independent, testable increment - developers can now freely use native names in local contexts

---

## Phase 4: User Story 3 - Consistent Word Resolution Order (P1)

**Goal**: Ensure all words follow single, unified resolution strategy

**Independent Test**: Word resolution traverses frame chain only, no special cases

**Note**: US3 is prioritized alongside US1 as they're interconnected (unified lookup enables consistent resolution)

### T015: [US3] [TEST] Write contract test for unified resolution path

**Goal**: Verify no special-case logic remains in word lookup

**Steps**:
1. Add to `test/contract/native_scoping_test.go`
2. Write `TestUnifiedResolutionPath`:
   - Create frame with native name shadowed
   - Verify frame value resolved first (not registry)
   - Verify undefined word raises error (no registry fallback)
   - Verify resolution order: current → parent → root

**Files**:
- MODIFIED: `test/contract/native_scoping_test.go` (+60 LOC)

**Acceptance**:
- Test passes (already implemented by T011-T013)
- Confirms no registry checks remain

**Duration**: 25 min

**Dependencies**: T013

**Reference**: spec.md User Story 3, SC-003

---

### T016: [US3] [TEST] Write contract test for frame chain traversal order

**Goal**: Verify resolution proceeds innermost to outermost

**Steps**:
1. Add to `test/contract/native_scoping_test.go`
2. Write `TestFrameChainTraversalOrder`:
   - Build 4-level frame chain
   - Place same word in frames 1, 2, 4 (skip 3)
   - Verify lookup from frame 3 finds frame 2's value
   - Verify lookup from frame 4 finds frame 4's value
   - Verify traversal stops at match (doesn't continue to outer)

**Files**:
- MODIFIED: `test/contract/native_scoping_test.go` (+50 LOC)

**Acceptance**:
- Test passes
- Traversal order validated
- Short-circuit behavior confirmed

**Duration**: 20 min

---

### T017: [US3] [IMPL] Code review: Verify no native.Lookup() calls remain

**Goal**: Audit codebase for any remaining registry dependencies

**Steps**:
1. Search codebase: `grep -r "native\.Lookup" internal/ test/`
2. Verify only test files reference it (for test utilities)
3. Confirm evaluator has zero registry calls
4. Document any remaining usage (should only be tests/benchmarks)

**Files**: None (audit only)

**Acceptance**:
- Zero `native.Lookup()` calls in production code (internal/eval/)
- Any test usage documented and justified
- SC-003 verified

**Duration**: 15 min

**Dependencies**: T013

---

### ✅ **CHECKPOINT 3**: User Story 3 Complete

**Verification**:
- ✅ Single, unified word resolution strategy
- ✅ Frame chain traversal order validated
- ✅ No special-case native lookups remain
- ✅ SC-003: Word lookups follow single resolution strategy
- ✅ SC-005: Lexical scoping consistent across all word types

**Deliverable**: Language semantics are now consistent - all words resolved identically

---

## Phase 5: User Story 2 - Shadow Native Functions Intentionally (P2)

**Goal**: Enable advanced patterns like wrapping, proxying, domain-specific overrides

**Independent Test**: Custom `print` function shadows native in local scope

### T018: [US2] [TEST] Write contract test for user-defined function shadowing native

**Goal**: Test that user functions can override natives in local scopes

**Steps**:
1. Add to `test/contract/native_scoping_test.go`
2. Write `TestUserFunctionShadowsNative`:
   - Define custom `print` function in local scope
   - Call `print` within that scope
   - Verify custom version executes (not native)
   - Pop scope, verify native `print` works at outer level

**Files**:
- MODIFIED: `test/contract/native_scoping_test.go` (+40 LOC)

**Acceptance**:
- Test passes (shadowing enabled by previous work)
- Custom function executes in inner scope
- Native accessible in outer scope

**Duration**: 20 min

**Dependencies**: T013

**Reference**: spec.md User Story 2, quickstart.md Test 3

---

### T019: [US2] [TEST] Write contract test for wrapped native pattern

**Goal**: Test pattern where user function wraps native implementation

**Steps**:
1. Add to `test/contract/native_scoping_test.go`
2. Write `TestWrappedNativePattern`:
   - Capture native function with get-word (`:+`)
   - Define custom function that calls captured native
   - Shadow native name with custom wrapper
   - Verify wrapper executes and calls original

**Files**:
- MODIFIED: `test/contract/native_scoping_test.go` (+60 LOC)

**Acceptance**:
- Test passes
- Wrapper pattern validated
- Original native accessible via captured reference

**Duration**: 25 min

---

### T020: [US2] [TEST] Write contract test for closure capture with shadowing

**Goal**: Verify closures capture values, not names (immune to rebinding)

**Steps**:
1. Add to `test/contract/native_scoping_test.go`
2. Write `TestClosureCaptureWithShadowing`:
   - Create closure capturing native function
   - Shadow native name after closure creation
   - Invoke closure
   - Verify closure uses captured native (not shadow)

**Files**:
- MODIFIED: `test/contract/native_scoping_test.go` (+50 LOC)

**Acceptance**:
- Test passes
- Closure semantics validated (lexical capture)
- Confirms clarification decision from spec

**Duration**: 25 min

**Reference**: spec.md Clarifications (closure behavior), quickstart.md Test 5

---

### T021: [US2] [TEST] Write contract test for multi-scope shadowing

**Goal**: Test complex shadowing scenario across multiple scopes

**Steps**:
1. Add to `test/contract/native_scoping_test.go`
2. Write `TestMultiScopeShadowing`:
   - Shadow native `+` in level 1 (multiply)
   - Shadow again in level 2 (subtract)
   - Verify each level sees its binding
   - Verify root level still has addition

**Files**:
- MODIFIED: `test/contract/native_scoping_test.go` (+50 LOC)

**Acceptance**:
- Test passes
- Multi-level shadowing validated
- Root native unchanged

**Duration**: 20 min

**Reference**: quickstart.md Test 4

---

### ✅ **CHECKPOINT 4**: User Story 2 Complete

**Verification**:
- ✅ User-defined functions can shadow natives in local scopes
- ✅ Wrapped native pattern supported
- ✅ Closure capture semantics correct (lexical)
- ✅ Multi-level shadowing works as expected
- ✅ SC-004: Native functions accessible unless explicitly shadowed

**Deliverable**: Advanced patterns enabled - library authors can wrap/override natives

---

## Phase 6: Polish & Cross-Cutting Concerns

### T022: [POLISH] Add evaluator construction contract test

**Goal**: Test that NewEvaluator initializes root frame correctly

**Steps**:
1. Create or modify `test/contract/evaluator_test.go`
2. Write `TestNewEvaluatorInitializesNatives`:
   - Create evaluator
   - Verify root frame at index 0
   - Verify all 70+ natives present
   - Verify native metadata preserved (Doc, Infix, Params)

**Files**:
- NEW or MODIFIED: `test/contract/evaluator_test.go` (~80 LOC)

**Acceptance**:
- Test passes
- All natives verified present
- Metadata checks pass

**Duration**: 30 min

**Reference**: contracts/evaluator.md, data-model.md Root Frame Validation

---

### T023: [POLISH] Add root frame validation contract test

**Goal**: Test root frame invariants

**Steps**:
1. Add to `test/contract/evaluator_test.go`
2. Write `TestRootFrameInvariants`:
   - Verify Index == 0
   - Verify Parent == -1
   - Verify Name == "(top level)"
   - Verify frame marked as captured

**Files**:
- MODIFIED: `test/contract/evaluator_test.go` (+40 LOC)

**Acceptance**:
- All invariants pass
- Root frame structure validated

**Duration**: 15 min

---

### T024: [POLISH] Add native registration failure test

**Goal**: Test panic behavior on registration failure

**Steps**:
1. Add to `test/contract/evaluator_test.go`
2. Write `TestNativeRegistrationPanic`:
   - Simulate duplicate native registration
   - Verify panic occurs with clear message
   - Test nil function panic

**Files**:
- MODIFIED: `test/contract/evaluator_test.go` (+60 LOC)

**Acceptance**:
- Panic tests pass
- Error messages clear

**Duration**: 25 min

**Reference**: contracts/evaluator.md Failure Modes

---

### T025: [POLISH] Benchmark evaluator construction performance

**Goal**: Verify construction time <1ms

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/eval/evaluator_bench_test.go`
2. Write `BenchmarkNewEvaluator`:
   - Measure NewEvaluator() time
   - Report ns/op
3. Verify <1ms (1,000,000 ns) target met

**Files**:
- NEW: `internal/eval/evaluator_bench_test.go` (~30 LOC)

**Acceptance**:
- Benchmark passes
- Construction time <1ms verified

**Duration**: 20 min

**Reference**: plan.md Performance Goals

---

### T026: [POLISH] [P] Benchmark word lookup performance (native)

**Goal**: Verify native lookup performance acceptable

**Steps**:
1. Create `/Users/marcin.radoszewski/dev-allegro/viro/internal/eval/lookup_bench_test.go`
2. Write `BenchmarkLookupNative`:
   - Measure `e.Lookup("+")` from top level
   - Compare to baseline (if available)
3. Write `BenchmarkLookupNativeFromNested`:
   - Measure lookup from 3-level nested scope

**Files**:
- NEW: `internal/eval/lookup_bench_test.go` (~60 LOC)

**Acceptance**:
- Benchmarks complete
- Performance within ±2% of baseline
- Native lookup time documented

**Duration**: 30 min

**Parallel with**: T027

---

### T027: [POLISH] [P] Benchmark word lookup performance (user-defined)

**Goal**: Verify user-defined word lookup performance

**Steps**:
1. Add to `internal/eval/lookup_bench_test.go`
2. Write `BenchmarkLookupUserDefined`:
   - Measure local variable lookup
   - Measure parent frame lookup
3. Compare to baseline

**Files**:
- MODIFIED: `internal/eval/lookup_bench_test.go` (+40 LOC)

**Acceptance**:
- Benchmarks complete
- User word lookup faster than before (no registry overhead)

**Duration**: 20 min

**Parallel with**: T026

**Reference**: research.md Q2 Performance Baseline

---

### T028: [POLISH] Remove native.Registry population (Phase 2 migration)

**Goal**: Stop populating legacy registry (no longer needed)

**Steps**:
1. Modify all 6 `register_*.go` files
2. Remove or comment out `Registry[name] = fn` assignments
3. Keep only `rootFrame.Bind(name, value.FuncVal(fn))` calls
4. Verify tests still pass (confirms evaluator uses frames only)

**Files**:
- MODIFIED: `internal/native/register_math.go` (-72 lines)
- MODIFIED: `internal/native/register_series.go` (-10 lines)
- MODIFIED: `internal/native/register_data.go` (-16 lines)
- MODIFIED: `internal/native/register_io.go` (-20 lines)
- MODIFIED: `internal/native/register_control.go` (-10 lines)
- MODIFIED: `internal/native/register_help.go` (-22 lines)

**Acceptance**:
- Registry no longer populated
- All tests pass (confirms frame-only usage)
- No compilation errors

**Duration**: 30 min

**Dependencies**: T014 (confirms all tests pass with frame-only lookup)

**Reference**: data-model.md Migration Path Phase 2

---

### T029: [POLISH] Remove native.Registry declaration

**Goal**: Delete unused global Registry variable

**Steps**:
1. Search for remaining `native.Registry` references: `grep -r "native\.Registry" .`
2. Verify only old registry.go initialization code remains
3. Delete `var Registry = make(map[string]*value.FunctionValue)` from `internal/native/registry.go`
4. Delete `registry.go` entirely if only Registry declaration remains
5. Verify no compilation errors

**Files**:
- DELETED or MODIFIED: `internal/native/registry.go` (if only Registry, delete entire file)

**Acceptance**:
- Registry variable removed
- Project compiles
- All tests pass

**Duration**: 15 min

**Dependencies**: T028

**Reference**: spec.md FR-010

---

### T030: [POLISH] Remove native.Lookup() function

**Goal**: Delete unused Lookup function

**Steps**:
1. Verify no remaining calls to `native.Lookup()` (T017 audit)
2. Delete `func Lookup(name string) (*value.FunctionValue, bool)` from native package
3. Update package documentation to remove registry references
4. Verify compilation successful

**Files**:
- MODIFIED: `internal/native/*.go` (remove Lookup function)

**Acceptance**:
- Lookup function removed
- No compilation errors
- All tests pass

**Duration**: 10 min

**Dependencies**: T029

**Reference**: spec.md FR-010, data-model.md Migration Path Phase 3

---

### T031: [POLISH] Update CLAUDE.md project documentation

**Goal**: Document new native registration pattern

**Steps**:
1. Modify `/Users/marcin.radoszewski/dev-allegro/viro/CLAUDE.md`
2. Update "Native Function System" section:
   - Document frame-based registration
   - Remove registry references
   - Add note about category-based files
   - Update "Adding a new native" instructions
3. Update "Common Gotchas" to mention shadowing is now allowed

**Files**:
- MODIFIED: `CLAUDE.md` (~20 line changes)

**Acceptance**:
- Documentation updated
- Accurate instructions for future native additions
- Shadowing capability documented

**Duration**: 20 min

---

### T032: [POLISH] Final verification and code review

**Goal**: Comprehensive final validation

**Steps**:
1. Run full test suite: `go test ./... -v`
2. Run benchmarks: `go test ./... -bench=. -benchmem`
3. Verify all success criteria met:
   - SC-001: Refinement parameters work ✅
   - SC-002: All existing tests pass ✅
   - SC-003: Single resolution strategy ✅
   - SC-004: Natives accessible unless shadowed ✅
   - SC-005: Consistent lexical scoping ✅
4. Check code coverage: `go test -coverprofile=coverage.out ./internal/eval ./internal/native`
5. Verify ≥95% coverage for modified files
6. Review generated documentation

**Files**: None (validation only)

**Acceptance**:
- All tests pass
- All success criteria met
- Code coverage ≥95%
- No performance regressions
- Documentation complete

**Duration**: 30 min

**Dependencies**: T031

---

### ✅ **FINAL CHECKPOINT**: Feature Complete

**Verification**:
- ✅ All 3 user stories implemented and tested
- ✅ All 5 success criteria verified
- ✅ Backward compatibility maintained (SC-002)
- ✅ Performance targets met (<1ms construction, no lookup regression)
- ✅ Documentation updated
- ✅ Code quality high (≥95% coverage)

**Deliverables**:
1. 6 category-based native registration files
2. Modified evaluator with unified word lookup
3. Comprehensive test suite (native_scoping_test.go, evaluator_test.go)
4. Performance benchmarks
5. Updated documentation

**Ready for**: Code review, merge to main branch

---

## Dependency Graph

```
Phase 1: Setup
  T001 (Baseline)

Phase 2: Foundational (Blocking)
  T002 (Math natives)
  T003-T007 [P] (Other category natives)
  T008 (NewEvaluator modification) ← depends on T002-T007

CHECKPOINT 1 ✅

Phase 3: User Story 1 (P1)
  T009 (Test: Refinement collision) ← depends on T008
  T010 (Test: Nested shadowing) [P] with T009
  T011 (Remove evalWord lookup) ← depends on T009, T010
  T012 (Remove evalGetWord lookup) ← depends on T011
  T013 (Remove evaluateWithFunctionCall lookup) ← depends on T012
  T014 (Verify tests pass) ← depends on T013

CHECKPOINT 2 ✅

Phase 4: User Story 3 (P1)
  T015 (Test: Unified resolution) ← depends on T013
  T016 (Test: Traversal order) [P] with T015
  T017 (Audit: No lookups remain) ← depends on T013

CHECKPOINT 3 ✅

Phase 5: User Story 2 (P2)
  T018 (Test: User function shadow) ← depends on T013
  T019 (Test: Wrapped native) [P] with T018
  T020 (Test: Closure capture) [P] with T018
  T021 (Test: Multi-scope) [P] with T018

CHECKPOINT 4 ✅

Phase 6: Polish
  T022 (Test: Constructor contract)
  T023 (Test: Root frame invariants) [P] with T022
  T024 (Test: Registration panic) [P] with T022
  T025 (Benchmark: Construction)
  T026 (Benchmark: Native lookup) [P] with T027
  T027 (Benchmark: User lookup) [P] with T026
  T028 (Remove registry population) ← depends on T014
  T029 (Remove registry declaration) ← depends on T028
  T030 (Remove Lookup function) ← depends on T029
  T031 (Update documentation)
  T032 (Final verification) ← depends on T031

FINAL CHECKPOINT ✅
```

---

## Parallel Execution Opportunities

### Foundational Phase (After T002)
**Parallel Set 1**: T003, T004, T005, T006, T007 (5 tasks)
- All create separate files
- No shared state
- **Time Saved**: ~90 min → ~35 min (largest time)

### User Story 1 (After T008)
**Parallel Set 2**: T009, T010 (2 tests)
- Both create/modify different test functions
- **Time Saved**: 50 min → 30 min

### User Story 2 (After T013)
**Parallel Set 3**: T018, T019, T020, T021 (4 tests)
- All add different test functions
- **Time Saved**: 110 min → 60 min

### Polish Phase
**Parallel Set 4**: T022, T023, T024 (3 tests)
- Different test functions
- **Time Saved**: 115 min → 60 min

**Parallel Set 5**: T026, T027 (2 benchmarks)
- Different benchmark functions
- **Time Saved**: 50 min → 30 min

**Total Time Savings**: ~235 minutes with full parallelization

---

## Implementation Strategy

### MVP Scope (Minimum Viable Product)

**Phase 1 + Phase 2 + Phase 3** = User Story 1 Complete
- **Tasks**: T001-T014 (14 tasks)
- **Duration**: ~5 hours (with parallelization)
- **Deliverable**: Developers can use native names without collisions
- **Value**: Solves primary pain point (P1 story)

### Incremental Delivery

**MVP + Phase 4** = Consistent Resolution (US1 + US3)
- **Tasks**: T001-T017 (17 tasks)
- **Duration**: ~6 hours
- **Deliverable**: Unified word lookup semantics

**MVP + Phases 4-5** = All User Stories (US1 + US2 + US3)
- **Tasks**: T001-T021 (21 tasks)
- **Duration**: ~8 hours
- **Deliverable**: Full feature with advanced patterns

**Full Feature** = All Phases
- **Tasks**: T001-T032 (32 tasks)
- **Duration**: ~11 hours (5-7 days at 2-3 hours/day)
- **Deliverable**: Production-ready, polished, documented

---

## Testing Summary

### Test Distribution

**Test Tasks**: 14 of 32 (44%)
- Contract tests: 11 tasks (T009, T010, T015-T024)
- Performance tests: 3 tasks (T025-T027)

**Test Coverage by Story**:
- User Story 1: 3 test tasks
- User Story 2: 4 test tasks
- User Story 3: 2 test tasks
- Polish: 5 test tasks

### Test Files Created

1. `test/contract/native_scoping_test.go` - Native shadowing scenarios (~300 LOC)
2. `test/contract/evaluator_test.go` - Constructor and invariants (~180 LOC)
3. `internal/eval/evaluator_bench_test.go` - Construction performance (~30 LOC)
4. `internal/eval/lookup_bench_test.go` - Lookup performance (~100 LOC)

**Total Test Code**: ~610 LOC

---

## Risk Mitigation

### High-Risk Areas

1. **Word Lookup Changes** (T011-T013)
   - **Mitigation**: TDD approach, tests written first
   - **Fallback**: Keep `native.Lookup()` as dead code until verified

2. **Backward Compatibility** (T014, T032)
   - **Mitigation**: Comprehensive existing test verification
   - **Fallback**: Dual-mode support (registry + frames)

3. **Performance Regression** (T026-T027)
   - **Mitigation**: Benchmark before/after comparison
   - **Fallback**: Optimize frame lookup if needed

---

## Success Metrics Mapping

| Success Criterion | Verified By Tasks | Checkpoint |
|-------------------|-------------------|------------|
| **SC-001**: Refinement parameters work | T009, T014 | Checkpoint 2 |
| **SC-002**: All existing tests pass | T001, T014, T032 | Checkpoints 2, 4, Final |
| **SC-003**: Single resolution strategy | T017, T032 | Checkpoint 3 |
| **SC-004**: Natives accessible unless shadowed | T018-T021 | Checkpoint 4 |
| **SC-005**: Consistent lexical scoping | T015, T016, T021 | Checkpoint 3 |

---

## Next Steps

1. **Review this task breakdown** with team/stakeholders
2. **Assign tasks** to developers (suggest T003-T007 for parallel work)
3. **Set up branch**: `git checkout -b 003-natives-within-frame`
4. **Begin with T001** (establish baseline)
5. **Execute foundational tasks** (T002-T008)
6. **Iterate by user story** (Phases 3-5)
7. **Polish and document** (Phase 6)
8. **Code review and merge**

**Estimated Timeline**:
- Day 1: Setup + Foundational (T001-T008)
- Day 2-3: User Story 1 & 3 (T009-T017)
- Day 4: User Story 2 (T018-T021)
- Day 5: Polish (T022-T032)

---

**Tasks Status**: ✅ Ready for Implementation | **Generated**: 2025-10-12 | **Total Duration**: 5-7 days
