# Implementation Plan: Natives Within Frame

**Branch**: `003-natives-within-frame` | **Date**: 2025-10-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-natives-within-frame/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature eliminates the special-case native function registry and moves native functions into the root frame, enabling standard lexical scoping for all word lookups. Currently, the evaluator checks `native.Registry` before frame lookups, preventing users from shadowing natives and causing confusing name collisions (e.g., `--debug` refinement conflicts with the native `debug` function). Since native functions are already represented as `FunctionValue` instances, they can be stored directly in the root frame like user-defined functions, unifying the resolution strategy and removing special-case logic from the evaluator.

## Technical Context

**Language/Version**: Go 1.21+ (uses generics)
**Primary Dependencies**:
- `github.com/marcin-radoszewski/viro/internal/value` (FunctionValue, Value types)
- `github.com/marcin-radoszewski/viro/internal/frame` (Frame, scoping)
- `github.com/marcin-radoszewski/viro/internal/eval` (Evaluator, NewEvaluator)
- `github.com/marcin-radoszewski/viro/internal/native` (Registry, native implementations)

**Storage**: In-memory frame bindings (word → value mappings)
**Testing**: Go test framework (`go test ./...`), table-driven tests in `test/contract/`, integration tests
**Target Platform**: Cross-platform (Linux, macOS, Windows)
**Project Type**: Single project (language interpreter)
**Performance Goals**: No performance degradation - word lookup must remain O(1) for frame bindings, frame chain traversal O(depth)
**Constraints**:
- Zero-allocation optimization for frame lookups where possible
- Evaluator construction must remain fast (<1ms for 70+ native registrations)

**Scale/Scope**:
- ~70 native functions across 13 categories
- 1 evaluator modification (NewEvaluator, word lookup paths)
- 5-10 affected code locations (evaluator, native package, tests)
- ~150 existing test cases must continue passing

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Note**: Viro project does not have a formal constitution file yet (template placeholders found). The project follows principles documented in CLAUDE.md:

### Principle I: Test-First Development (TDD)
✅ **PASS** - Feature includes comprehensive contract tests before implementation
- Contract tests defined in spec (User Stories 1-3 with acceptance scenarios)
- Test-first workflow: write tests → verify failure → implement → verify pass
- Existing test suites must continue passing (SC-002)

### Principle II: Single Project Structure
✅ **PASS** - Fits existing single-project interpreter architecture
- Changes confined to `internal/eval/`, `internal/native/`, `test/contract/`
- No new libraries or services introduced
- Standard Go package structure maintained

### Principle III: Backward Compatibility
✅ **PASS** - Explicit requirement FR-007, SC-002
- Existing code that doesn't shadow natives continues working identically
- No breaking changes to Viro language semantics
- Migration strategy preserves all current functionality

### Principle IV: Simplicity & YAGNI
✅ **PASS** - Removes complexity rather than adding it
- Eliminates special-case registry lookup logic
- Unifies word resolution to single code path
- Reduces total lines of code (removes `native.Lookup()` callsites)

### Principle V: Performance Awareness
✅ **PASS** - Performance constraints explicitly stated
- No degradation in word lookup performance
- Frame-based lookup already optimized (O(1) per frame)
- Initialization performance tracked (<1ms for all natives)

**Constitution Check Result**: ✅ All gates passed. No violations to justify.

## Project Structure

### Documentation (this feature)

```
specs/003-natives-within-frame/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── evaluator.md     # NewEvaluator construction contract
│   ├── frame.md         # Root frame initialization contract
│   └── lookup.md        # Word lookup resolution contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Viro is a single-project interpreter with standard Go structure
internal/
├── eval/
│   └── evaluator.go     # MODIFIED: NewEvaluator() to initialize root frame with natives
│                        # MODIFIED: word lookup to remove native.Lookup() calls
├── native/
│   ├── registry.go      # MODIFIED: Split into category-specific registration files
│   ├── register_math.go      # NEW: Math operations registration (Groups 1-4)
│   ├── register_series.go    # NEW: Series operations registration (Group 5)
│   ├── register_data.go      # NEW: Data operations registration (Groups 6-7)
│   ├── register_io.go        # NEW: I/O and Port operations registration (Groups 8-9)
│   ├── register_control.go   # NEW: Control flow registration (Groups 10-11)
│   ├── register_help.go      # NEW: Help and reflection registration (Groups 12-13)
│   ├── math.go          # UNCHANGED: Math function implementations
│   ├── series.go        # UNCHANGED: Series function implementations
│   ├── control.go       # UNCHANGED: Control flow implementations
│   └── [other native impl files] # UNCHANGED
├── frame/
│   └── frame.go         # UNCHANGED: Frame struct and methods
└── value/
    └── value.go         # UNCHANGED: Value and FunctionValue types

test/
├── contract/
│   ├── native_scoping_test.go   # NEW: Tests for native shadowing behavior
│   ├── word_lookup_test.go      # MODIFIED: Update tests for unified lookup
│   └── [existing tests]         # MUST PASS: All existing contract tests
└── integration/
    └── [existing tests]          # MUST PASS: All existing integration tests
```

**Structure Decision**: Single project structure maintained. Native registration is split into 6 category-based files (`register_*.go`) to improve maintainability and code organization, as suggested by the user. This keeps each registration file focused (10-15 functions each) while preserving all existing implementation files unchanged. The evaluator is modified to call native registration functions during `NewEvaluator()` construction and to remove the special-case `native.Lookup()` calls from word resolution paths.

## Complexity Tracking

*No Constitution Check violations - this section is empty per template instructions.*

## Phase 0: Research & Unknowns

**Status**: Ready to execute

**Research Topics**:
1. Current native registry initialization order and dependencies
2. Frame binding performance characteristics (lookup, insertion, growth)
3. Existing native function metadata preservation requirements
4. Evaluator construction sequence and failure modes
5. Testing strategy for closure capture semantics with shadowed natives

**Unknowns to Resolve**:
- Are there any initialization-order dependencies between natives? (e.g., does native A call native B during registration?)
- What is the current performance baseline for word lookups? (frame vs. registry)
- How are native functions currently tested for name collisions or refinement conflicts?
- Are there any natives that modify the evaluator state during registration?

**Output**: `research.md` documenting findings and decisions

## Phase 1: Design Artifacts

**Status**: Pending Phase 0 completion

**Artifacts to Generate**:
1. **data-model.md**: Entity definitions
   - Root Frame initialization model
   - Native registration lifecycle
   - Word lookup resolution flow

2. **contracts/**: API contracts for modified components
   - `evaluator.md`: NewEvaluator() construction contract
   - `frame.md`: Root frame initialization contract
   - `lookup.md`: Unified word lookup resolution contract

3. **quickstart.md**: Developer guide for testing native shadowing behavior

4. **Agent context update**: Run `.specify/scripts/bash/update-agent-context.sh claude` to update AI context with new registration patterns

**Output**: Complete design documentation suite in `specs/003-natives-within-frame/`

## Phase 2: Implementation Tasks

**Status**: Not started (generated by `/speckit.tasks` command after Phase 1)

**High-Level Task Categories** (detailed breakdown in tasks.md):
1. **Refactor Native Registration** (~6 tasks)
   - Split registry.go into category-based register_*.go files
   - Create registration helper functions
   - Test each category independently

2. **Modify Evaluator Construction** (~4 tasks)
   - Add native registration call to NewEvaluator()
   - Initialize root frame with all natives
   - Add panic-on-failure error handling
   - Test evaluator construction

3. **Remove Native Registry Lookups** (~5 tasks)
   - Remove `native.Lookup()` calls from evaluator.go
   - Update word resolution to use frame chain only
   - Remove `native.Registry` export after confirming all usage removed
   - Update documentation

4. **Test Native Shadowing Behavior** (~8 tasks)
   - Write contract tests for shadowing scenarios
   - Test refinement name collision resolution
   - Test closure capture with shadowed natives
   - Test multi-library shadowing behavior
   - Verify all existing tests still pass

5. **Performance Validation** (~3 tasks)
   - Benchmark word lookup performance
   - Benchmark evaluator construction time
   - Verify no regressions

**Estimated Effort**: 5-7 days (26 total tasks estimated)

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Performance regression in word lookups | High | Benchmark before/after, optimize frame lookup if needed |
| Breaking existing code that relies on native registry | High | Comprehensive test suite coverage, manual review of all `native.Registry` usage |
| Initialization-order dependencies between natives | Medium | Research Phase 0 identifies dependencies, document and preserve order |
| Evaluator construction failures difficult to debug | Medium | Fail-fast with clear panic messages, test each registration category |
| Closure capture semantics unexpected behavior | Low | Comprehensive contract tests written before implementation, manual verification |

## Dependencies

**Internal**:
- `internal/value` - FunctionValue type (no changes needed)
- `internal/frame` - Frame binding methods (no changes needed)
- `internal/eval` - Evaluator construction and word lookup (modified)
- `internal/native` - Native implementations (registration refactored)

**External**:
- None (internal refactoring only)

**Blocking Issues**:
- None identified

## Success Metrics

**Completion Criteria** (from spec.md):
- ✅ SC-001: Functions with `--debug` refinement parameters work without collision errors
- ✅ SC-002: All existing Viro code and test suites pass without modification
- ✅ SC-003: Word lookups follow single resolution strategy (no `native.Lookup()` calls remain)
- ✅ SC-004: Native functions remain accessible unless explicitly shadowed
- ✅ SC-005: Lexical scoping consistent across all word types

**Additional Validation**:
- Code coverage maintained at ≥95% for modified files
- No performance regressions (±2% tolerance on benchmarks)
- Zero failing tests in existing suite
- Clean code review with no major issues flagged

## Next Steps

1. ✅ **Complete Phase 0**: Execute research tasks, document findings in `research.md`
2. ⏳ **Complete Phase 1**: Generate data-model.md, contracts/, quickstart.md
3. ⏳ **Run `/speckit.tasks`**: Generate detailed task breakdown in tasks.md
4. ⏳ **Run `/speckit.implement`**: Execute implementation workflow

---

**Plan Status**: Phase 0 Ready | Generated: 2025-10-12 | Author: Claude Code + User Input
