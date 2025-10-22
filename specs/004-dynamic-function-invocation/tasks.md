---
description: "Task list for Dynamic Function Invocation (Action Types) feature"
---

# Tasks: Dynamic Function Invocation (Action Types)

**Input**: Design documents from `/specs/004-dynamic-function-invocation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included per TDD requirements in CLAUDE.md and plan.md (TDD Mandatory constitutional requirement)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure required for action system

- [ ] T001 Add TypeAction constant to internal/value/types.go (new value type constant)
- [ ] T002 [P] Create internal/value/action.go stub file (ActionValue struct definition)
- [ ] T003 [P] Create internal/eval/dispatch.go stub file (dispatch logic placeholder)
- [ ] T004 [P] Create internal/native/action.go stub file (action creation utilities)
- [ ] T005 [P] Create internal/frame/typeframe.go stub file (type frame initialization)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Define ActionValue struct in internal/value/action.go (Name string, ParamSpec []Parameter fields)
- [ ] T007 Add ActionVal() constructor and AsAction() assertion methods in internal/value/action.go
- [ ] T008 Implement action.String() and action.Equals() methods in internal/value/action.go
- [ ] T009 Create global TypeRegistry variable in internal/frame/typeframe.go (map[ValueType]*Frame)
- [ ] T010 Implement InitTypeFrames() function in internal/frame/typeframe.go (creates type frames for all native types)
- [ ] T011 Add GetTypeFrame(ValueType) function in internal/frame/typeframe.go (lookup type frame from registry)
- [ ] T012 Call InitTypeFrames() in cmd/viro/main.go after root frame initialization
- [ ] T013 Add TypeAction case to evaluator switch in internal/eval/evaluator.go (dispatch entry point)
- [ ] T014 Implement DispatchAction() function in internal/eval/dispatch.go (type lookup, function resolution, invocation)
- [ ] T015 Add action-no-impl error code (341) to internal/verror/errors.go
- [ ] T016 Add action-frame-corrupt error code (941) to internal/verror/errors.go
- [ ] T017 Implement helper function CreateAction(name, paramSpec) in internal/native/action.go
- [ ] T018 Implement helper function RegisterActionImpl(typeName, actionName, func) in internal/native/action.go

**Checkpoint**: Foundation ready - action dispatch infrastructure complete, user story implementation can now begin

---

## Phase 3: User Story 1 - Type-Safe Series Operations (Priority: P1) 🎯 MVP

**Goal**: Enable polymorphic dispatch for series operations (first, last, append, insert, length?) on blocks and strings

**Independent Test**: Define action `first` with block/string implementations, verify `first [1 2 3]` and `first "hello"` dispatch correctly

### Contract Tests for User Story 1 (TDD - Write First)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T019 [P] [US1] Contract test for action dispatch basics in test/contract/action_dispatch_test.go
- [ ] T020 [P] [US1] Contract test for series actions in test/contract/series_action_test.go
- [ ] T021 [P] [US1] Contract test for first action in test/contract/series_action_test.go (TestActionFirst)
- [ ] T022 [P] [US1] Contract test for last action in test/contract/series_action_test.go (TestActionLast)
- [ ] T023 [P] [US1] Contract test for append action in test/contract/series_action_test.go (TestActionAppend)
- [ ] T024 [P] [US1] Contract test for insert action in test/contract/series_action_test.go (TestActionInsert)
- [ ] T025 [P] [US1] Contract test for length? action in test/contract/series_action_test.go (TestActionLength)

### Implementation for User Story 1

- [ ] T026 [P] [US1] Create internal/native/series_block.go with block-specific first implementation
- [ ] T027 [P] [US1] Create internal/native/series_string.go with string-specific first implementation
- [ ] T028 [US1] Implement block-specific last function in internal/native/series_block.go
- [ ] T029 [US1] Implement string-specific last function in internal/native/series_string.go
- [ ] T030 [US1] Implement block-specific append function in internal/native/series_block.go
- [ ] T031 [US1] Implement string-specific append function in internal/native/series_string.go
- [ ] T032 [US1] Implement block-specific insert function in internal/native/series_block.go
- [ ] T033 [US1] Implement string-specific insert function in internal/native/series_string.go
- [ ] T034 [US1] Implement block-specific length? function in internal/native/series_block.go
- [ ] T035 [US1] Implement string-specific length? function in internal/native/series_string.go
- [ ] T036 [US1] Register block-specific implementations into block type frame in internal/native/series_block.go (init function)
- [ ] T037 [US1] Register string-specific implementations into string type frame in internal/native/series_string.go (init function)
- [ ] T038 [US1] Create action values for first, last, append, insert, length? in internal/native/register_series.go
- [ ] T039 [US1] Bind action values to root frame in internal/native/register_series.go (replacing old natives)
- [ ] T040 [US1] Remove old native implementations from internal/native/register_series.go (commented out or deleted)
- [ ] T041 [US1] Verify all existing series operation tests still pass (go test ./test/contract/... -v)

**Checkpoint**: At this point, all 5 series actions (first, last, append, insert, length?) work polymorphically for blocks and strings

---

## Phase 4: User Story 2 - Extensibility for Future User-Defined Types (Priority: P2)

**Goal**: Ensure action dispatch architecture supports future user-defined types without requiring core changes

**Independent Test**: Verify design doesn't hardcode native types, and dispatch logic would handle hypothetical custom types identically

### Tests for User Story 2 (Architecture Validation)

- [ ] T042 [US2] Architecture test: verify TypeRegistry uses map[ValueType]*Frame (not hardcoded types) in test/contract/action_dispatch_test.go (TestTypeRegistryExtensibility)
- [ ] T043 [US2] Architecture test: verify InitTypeFrames() is data-driven (can add types without code changes) in test/contract/action_dispatch_test.go (TestTypeFrameRegistration)
- [ ] T044 [US2] Documentation test: verify research.md and data-model.md explain how to register new types

### Implementation for User Story 2

- [ ] T045 [US2] Refactor InitTypeFrames() to use type metadata registry (if needed for extensibility)
- [ ] T046 [US2] Add RegisterTypeFrame(typeName, frame) function in internal/frame/typeframe.go (for future custom types)
- [ ] T047 [US2] Document type registration process in internal/frame/typeframe.go (godoc comments)
- [ ] T048 [US2] Create example stub for hypothetical custom type in test/contract/action_dispatch_test.go (TestCustomTypeStub)
- [ ] T049 [US2] Verify dispatch logic in internal/eval/dispatch.go treats all types uniformly (no special cases)

**Checkpoint**: Architecture validated for extensibility - no hardcoded type assumptions, ready for future custom types

---

## Phase 5: User Story 3 - Error Handling for Missing Type Implementations (Priority: P3)

**Goal**: Provide clear, actionable error messages when actions are called on unsupported types

**Independent Test**: Define action with partial type support, call on unsupported type, verify error message is clear

### Contract Tests for User Story 3 (TDD - Write First)

- [ ] T050 [P] [US3] Contract test for action-no-impl error in test/contract/action_errors_test.go (TestActionNoImpl)
- [ ] T051 [P] [US3] Contract test for wrong-arity error in test/contract/action_errors_test.go (TestActionWrongArity)
- [ ] T052 [P] [US3] Contract test for empty series error in test/contract/action_errors_test.go (TestActionEmptySeries)
- [ ] T053 [P] [US3] Contract test for type-mismatch error in test/contract/action_errors_test.go (TestActionTypeMismatch)

### Implementation for User Story 3

- [ ] T054 [US3] Implement error generation for unsupported type in internal/eval/dispatch.go (action-no-impl with action name + type)
- [ ] T055 [US3] Implement error generation for missing function in type frame in internal/eval/dispatch.go (action-frame-corrupt internal error)
- [ ] T056 [US3] Add Near context (action call expression + first arg) to dispatch errors in internal/eval/dispatch.go
- [ ] T057 [US3] Add Where context (call stack trace) to dispatch errors in internal/eval/dispatch.go
- [ ] T058 [US3] Verify error messages in REPL match contracts (manual test: ./viro then try `first 42`)
- [ ] T059 [US3] Update quickstart.md error handling section with actual error message format

**Checkpoint**: All error scenarios have clear, helpful error messages with proper context

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories, validation, and documentation

- [ ] T060 [P] Add benchmarks for action dispatch overhead in test/contract/action_benchmark_test.go (BenchmarkActionDispatch vs BenchmarkDirectCall)
- [ ] T061 [P] Run go test ./... to verify all tests pass (full test suite validation)
- [ ] T062 [P] Run gofmt -w internal/ to format all code
- [ ] T063 [P] Update CLAUDE.md with action system documentation (if needed)
- [ ] T064 Code review: verify index-based architecture maintained (type frames use Parent=0 index, not pointers)
- [ ] T065 Code review: verify local-by-default scoping respected (actions in root frame, can be shadowed)
- [ ] T066 Verify quickstart.md examples all work in REPL (manual validation)
- [ ] T067 Run benchmark and document dispatch overhead in research.md (Performance Considerations section)
- [ ] T068 Create migration checklist verification in contracts/series-actions.md (ensure all items checked)
- [ ] T069 Final validation: run all contract tests with -v flag and verify 100% pass rate

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User Story 1 (Phase 3) can start after Foundational
  - User Story 2 (Phase 4) can start after Foundational, but validates US1 design
  - User Story 3 (Phase 5) can start after Foundational, adds error handling to US1 dispatch
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Type-Safe Series Operations**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2) - Extensibility**: Can start after Foundational (Phase 2), validates US1 architecture but is independently testable
- **User Story 3 (P3) - Error Handling**: Can start after Foundational (Phase 2), enhances US1 dispatch with error reporting but is independently testable

### Within Each User Story

- Contract tests MUST be written and FAIL before implementation (TDD)
- Type-specific implementations (series_block.go, series_string.go) before registration
- Registration of type-specific functions into type frames before action creation
- Action value creation before binding to root frame
- Old native removal only after new action works

### Parallel Opportunities

- **Setup (Phase 1)**: T002-T005 all marked [P] can run in parallel (different files)
- **User Story 1 Contract Tests**: T019-T025 all marked [P] can run in parallel (different test functions)
- **User Story 1 Implementation**: T026-T027 can run in parallel (different files: series_block.go vs series_string.go)
- **User Story 3 Contract Tests**: T050-T053 all marked [P] can run in parallel (different test functions)
- **Polish**: T060-T063 all marked [P] can run in parallel (independent validation tasks)

---

## Parallel Example: User Story 1 Contract Tests

```bash
# Launch all contract tests for User Story 1 together (TDD - ensure they FAIL first):
Task: "Contract test for action dispatch basics in test/contract/action_dispatch_test.go"
Task: "Contract test for series actions in test/contract/series_action_test.go"
Task: "Contract test for first action (TestActionFirst)"
Task: "Contract test for last action (TestActionLast)"
Task: "Contract test for append action (TestActionAppend)"

# Then verify all tests fail (red):
go test ./test/contract/... -v

# Then implement type-specific functions in parallel:
Task: "Create internal/native/series_block.go with block-specific implementations"
Task: "Create internal/native/series_string.go with string-specific implementations"

# Then verify all tests pass (green):
go test ./test/contract/... -v
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup → T001-T005 (stub files created)
2. Complete Phase 2: Foundational → T006-T018 (CRITICAL - action dispatch infrastructure ready)
3. Complete Phase 3: User Story 1 → T019-T041 (series actions work polymorphically)
4. **STOP and VALIDATE**: Run `go test ./test/contract/... -v` and verify all series action tests pass
5. **STOP and VALIDATE**: Run `./viro` REPL and manually test `first [1 2 3]`, `first "hello"`, etc.
6. Merge to main if ready (User Story 1 delivers core polymorphic dispatch functionality)

### Incremental Delivery

1. Complete Setup + Foundational (Phases 1-2) → Foundation ready
2. Add User Story 1 (Phase 3) → Test independently → Merge (MVP! Core series actions work polymorphically)
3. Add User Story 2 (Phase 4) → Test independently → Merge (Architecture validated for extensibility)
4. Add User Story 3 (Phase 5) → Test independently → Merge (Error handling complete)
5. Complete Polish (Phase 6) → Final validation → Feature complete

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (Phases 1-2)
2. Once Foundational is done:
   - Developer A: User Story 1 (Phase 3) - Core series action implementation
   - Developer B: User Story 2 (Phase 4) - Architecture validation tests
   - Developer C: User Story 3 (Phase 5) - Error handling tests and implementation
3. Stories complete and integrate independently

### TDD Workflow (User Story 1 Example)

1. Write contract tests first (T019-T025) - ensure they compile but FAIL
2. Run `go test ./test/contract/... -v` → see RED failures
3. Implement type-specific functions (T026-T035) - keep tests failing initially
4. Register implementations and create actions (T036-T039) - tests should start passing
5. Run `go test ./test/contract/... -v` → see GREEN passes
6. Refactor if needed (T040-T041) - tests stay GREEN
7. Commit with message: "feat: Implement polymorphic series actions (US1)"

---

## Notes

- **[P] tasks** = different files, no dependencies, can run in parallel
- **[Story] label** maps task to specific user story for traceability (US1, US2, US3)
- **TDD Mandatory**: All contract tests written before implementation (constitutional requirement from plan.md)
- **Each user story** is independently completable and testable
- **Verify tests fail** (RED) before implementing (TDD red-green-refactor)
- **Commit frequently**: After each task or logical group
- **Stop at checkpoints** to validate story independently
- **Index-based architecture**: Type frames use Parent=0 (index) not pointers
- **Direct pointer storage**: Type frames stored in TypeRegistry (not on stack), Index=-1
- **Avoid**: Vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Task Count Summary

- **Total tasks**: 69
- **Setup (Phase 1)**: 5 tasks (T001-T005)
- **Foundational (Phase 2)**: 13 tasks (T006-T018) - BLOCKS all user stories
- **User Story 1 (Phase 3)**: 23 tasks (T019-T041) - Core MVP functionality
- **User Story 2 (Phase 4)**: 8 tasks (T042-T049) - Architecture validation
- **User Story 3 (Phase 5)**: 10 tasks (T050-T059) - Error handling
- **Polish (Phase 6)**: 10 tasks (T060-T069) - Validation and documentation

**Parallel opportunities**: 18 tasks marked [P] across all phases

**Suggested MVP scope**: Phases 1-3 (User Story 1 only) = 41 tasks → Delivers core polymorphic dispatch functionality

**Independent test criteria**:
- **US1**: Call `first [1 2 3]` → 1, `first "hello"` → "h" (polymorphic dispatch works)
- **US2**: TypeRegistry uses map, not hardcoded types (extensible architecture)
- **US3**: Call `first 42` → clear error "Action 'first' not defined for type integer!" (helpful errors)
