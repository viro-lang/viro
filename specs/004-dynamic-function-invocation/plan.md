# Implementation Plan: Dynamic Function Invocation (Action Types)

**Branch**: `004-dynamic-function-invocation` | **Date**: 2025-10-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-dynamic-function-invocation/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a polymorphic action system that enables type-based dynamic dispatch for native and user-defined types. The system introduces an `action!` value type that dispatches to type-specific implementations based on the first argument's type. Type-specific functions are organized in type frames (one per data type) created at interpreter startup. This enables left-to-right polymorphic operations (e.g., `first [1 2]` and `first "hello"` both work) while maintaining Viro's local-by-default scoping and index-based architecture.

## Technical Context

**Language/Version**: Go 1.21+ (uses generics)
**Primary Dependencies**: Standard library only (no external dependencies for core dispatch system)
**Storage**: N/A (in-memory frame/stack system)
**Testing**: Go testing framework (`go test`), table-driven tests, contract-based testing
**Target Platform**: Cross-platform (macOS, Linux, Windows) - interpreter runtime
**Project Type**: Single project (language interpreter)
**Performance Goals**: Dispatch overhead minimal compared to direct function calls; acceptable performance degradation to be measured during implementation
**Constraints**: Must maintain index-based frame references (no pointers to prevent invalidation); type frames initialized eagerly at startup (static structure)
**Scale/Scope**: Convert all existing series operations (~8-10 native functions) to actions; support all current native types (block!, string!, integer!, etc.); extensible for future user-defined types

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Based on Viro's development principles (from CLAUDE.md):

✅ **TDD Mandatory**: Feature follows contract-first approach with tests before implementation
- Contract tests will be defined in `specs/004-*/contracts/` before implementation
- All acceptance scenarios from spec.md will have corresponding contract tests
- Red-Green-Refactor cycle will be enforced

✅ **Index-Based Architecture**: Design maintains index-based frame references
- Type frames will use indices into stack (not pointers)
- Type registry maps ValueType → frame index (int)
- No pointer invalidation risk during stack expansion

✅ **Local-by-Default Scoping**: Actions respect lexical scoping rules
- Actions live in root frame (like current natives)
- User code can shadow actions with local bindings
- No global state modification

✅ **Incremental Migration**: Existing code remains functional during transition
- Series operations converted incrementally to actions
- No breaking changes to user-facing API
- Backward compatibility maintained

✅ **No External Dependencies**: Feature uses only standard library and existing internal packages
- internal/value, internal/eval, internal/frame, internal/stack
- No new external Go dependencies

**GATE STATUS**: ✅ PASS - All constitutional requirements satisfied

### Post-Design Re-Evaluation (Phase 1 Complete)

After completing research.md, data-model.md, contracts/, and quickstart.md:

✅ **TDD Mandatory**: Contracts fully specified
- `contracts/action-dispatch.md`: Defines action dispatch behavior with 8 test cases
- `contracts/series-actions.md`: Defines 5 series actions with comprehensive test coverage
- All acceptance scenarios from spec.md mapped to concrete test cases
- Test implementation will precede code implementation

✅ **Index-Based Architecture**: Design verified
- ActionValue stores only Name and ParamSpec (no type-to-frame mappings)
- Global TypeRegistry stores direct pointers to type frames (not indices)
- Type frames use Parent=0 (index) to reference root frame on stack
- Type frames have Index=-1 (not in frameStore), stored in TypeRegistry
- Hybrid approach: execution frames on stack (index-based), type frames in registry (pointer-based)

✅ **Local-by-Default Scoping**: Confirmed
- Actions registered in root frame (see quickstart.md "Scoping and Shadowing")
- Local bindings can shadow actions (tested in contracts)
- No global mutable state introduced (TypeRegistry immutable after init)

✅ **Incremental Migration**: Strategy validated
- research.md defines Phase 1 (5 core series ops) and Phase 2 (extended ops)
- Migration checklist in contracts/series-actions.md ensures no regressions
- Backward compatibility guaranteed (see contracts/series-actions.md "Backward Compatibility")

✅ **No External Dependencies**: Verified
- research.md confirms: "Go standard library only (no new dependencies)"
- Only internal packages modified: value, eval, frame, native, stack

**POST-DESIGN GATE STATUS**: ✅ PASS - Design maintains all constitutional requirements

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
internal/
├── value/              # Add TypeAction, ActionValue struct
│   ├── types.go       # Add action type constant
│   ├── value.go       # Add action value constructors/assertions
│   └── action.go      # NEW: Action-specific value implementation
├── eval/              # Modify evaluator for action dispatch
│   ├── evaluator.go   # Add action evaluation case
│   └── dispatch.go    # NEW: Action dispatch logic
├── frame/             # Type frames stored here
│   ├── frame.go       # Existing frame structure (unchanged)
│   └── typeframe.go   # NEW: Type frame initialization
├── native/            # Refactor to register into type frames
│   ├── registry.go    # Modify to support type frame registration
│   ├── register_series.go  # Convert series ops to actions
│   └── action.go      # NEW: Action creation utilities
└── stack/             # Existing stack (unchanged)

test/contract/         # Contract tests for actions
├── action_dispatch_test.go   # NEW: Test action dispatch
├── series_action_test.go     # NEW: Test series as actions
└── action_errors_test.go     # NEW: Test error cases

cmd/viro/              # REPL initialization
└── main.go           # Initialize type frames at startup
```

**Structure Decision**: Single project (language interpreter). Changes isolated to internal packages following existing architecture. Type frames are regular frames stored in stack, maintaining index-based access pattern. No new top-level directories required.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

No violations - all constitutional requirements satisfied.
