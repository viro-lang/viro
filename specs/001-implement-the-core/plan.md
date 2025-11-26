# Implementation Plan: Viro Core Language and REPL

**Branch**: `001-implement-the-core` | **Date**: 2025-01-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-implement-the-core/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a homoiconic programming language interpreter in Go with a working REPL. The core requirement is to create a type-based dispatch system that evaluates expressions (literals, words, functions, blocks) following strict TDD methodology. The implementation follows a layered architecture: Core Evaluator → Type System → Native Functions (~50 initially) → Frame/Context System → Error Handling → REPL Interface. This phase establishes the foundational architecture for the interpreter, supporting basic expressions, control flow, series operations, user-defined functions, and interactive REPL features with command history.

## Technical Context

**Language/Version**: Go 1.21+ (requires generics and improved error handling)  
**Primary Dependencies**: Go standard library + github.com/chzyer/readline (REPL command history and multi-line input)  
**Storage**: N/A (in-memory interpreter state only)  
**Testing**: Go standard library `testing` package with table-driven test pattern (subtests via t.Run())  
**Target Platform**: macOS primary, Linux/Windows compatible (cross-platform Go)  
**Project Type**: Single CLI application (interpreter + REPL)  
**Performance Goals**: Stack ops <100ns, simple eval <10ms, complex eval <100ms per spec SC-005, tracked via Go benchmarks  
**Constraints**: Interactive response time <100ms for 95th percentile operations, stack depth 100+ function calls, command history 100+ entries, supports 10,000 values in memory simultaneously  
**Scale/Scope**: ~29 native functions initially (expanding to 600+ long-term), 8 core value types, 7 error categories, 1000+ REPL evaluation cycles without leaks

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

✅ **Principle I - TDD (NON-NEGOTIABLE)**: Planning includes Phase 1 contract definitions before implementation. Task breakdown will enforce test-first ordering.

✅ **Principle II - Incremental Layering**: Plan follows exact architecture sequence from constitution: Core Evaluator → Type System → Minimal Natives → Frame/Context → Error Handling → Extended Natives → Parse (deferred).

✅ **Principle III - Type Dispatch Fidelity**: Data model will define type-based dispatch with evaluation type maps. Each value type gets appropriate evaluation handler.

✅ **Principle IV - Stack and Frame Safety**: Stack design mandates index-based access (not pointers). Frame layout specified in contracts with return slot, prior frame, metadata, arguments.

✅ **Principle V - Structured Errors**: Spec defines 7 error categories (0-900 range) with required fields (code, type, id, arg1-3, near, where). Error contracts will enforce structure.

✅ **Principle VI - Observable Behavior**: REPL provides text I/O for evaluation results and errors. Error messages include context per FR-036. Trace capabilities deferred to later phase per clarification.

✅ **Principle VII - YAGNI**: Spec explicitly limits to ~50 native functions (FR-017 through FR-034). Parse dialect, module system, advanced features marked out-of-scope.

**Gate Status**: ✅ PASSED - All constitutional requirements satisfied by specification and planned architecture.

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
viro/
├── internal/
│   ├── value/              # Value types: integer, string, word, block, function, logic, none
│   │   ├── value.go        # Core Value type with type tag and payload
│   │   ├── block.go        # Block series implementation
│   │   ├── word.go         # Word types: word, lit-word, get-word, set-word
│   │   ├── function.go     # Native and user function types
│   │   └── types.go        # Type constants and type checking
│   ├── eval/               # Core evaluator
│   │   ├── eval.go         # Do_Next, Do_Blk - main evaluation loop
│   │   ├── dispatch.go     # Type-based dispatch routing
│   │   └── evaltype.go     # Evaluation type classification
│   ├── stack/              # Stack management
│   │   ├── stack.go        # Unified stack for data + frames
│   │   ├── expand.go       # Automatic expansion logic
│   │   └── index.go        # Index-based access helpers
│   ├── frame/              # Frame & context system
│   │   ├── frame.go        # Frame structure: word list + value list
│   │   ├── binding.go      # Bind/get/set operations
│   │   └── types.go        # Frame types: function args, closures
│   ├── native/             # Native function implementations
│   │   ├── control.go      # if, when, loop, while
│   │   ├── data.go         # set, get, type?
│   │   ├── io.go           # print, input
│   │   ├── math.go         # +, -, *, /, <, >, =, and, or, not
│   │   ├── series.go       # first, last, append, insert, length?
│   │   ├── function.go     # fn (function definition)
│   │   └── registry.go     # Native function registration and dispatch
│   ├── error/              # Structured error system
│   │   ├── error.go        # Error structure with categories
│   │   ├── categories.go   # Error codes 0-900
│   │   └── context.go      # Near/where context capture
│   └── parse/              # Parse dialect (deferred to later phase)
├── pkg/
│   └── viro/               # Public API (if needed for embedding)
├── cmd/
│   └── viro/               # CLI entry point
│       ├── main.go         # Program entry, REPL initialization
│       └── repl.go         # REPL loop: read-eval-print, history, multi-line
└── test/
    ├── contract/           # Contract tests for each native function
    │   ├── control_test.go
    │   ├── data_test.go
    │   ├── io_test.go
    │   ├── math_test.go
    │   └── series_test.go
    ├── integration/        # End-to-end interpreter tests
    │   ├── eval_test.go    # Core evaluation scenarios
    │   ├── errors_test.go  # Error handling scenarios
    │   └── repl_test.go    # REPL interaction scenarios
    └── fixtures/           # Test scripts
        └── examples/
```

**Structure Decision**: Single project layout selected. This is a standalone CLI interpreter, not a web application or mobile app. The internal/ directory contains implementation packages not intended for external import. The pkg/viro/ directory is reserved for future embedding API if needed. Testing follows Go conventions with _test.go files and separate contract/integration test organization.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

No violations detected. All constitutional principles are satisfied by the current plan.

---

## Phase 0: Research (Complete)

✅ **Completed**: All technical unknowns resolved

**Research Artifacts**:
- `research.md`: Comprehensive research covering 6 areas:
  1. REPL readline library selection (chzyer/readline chosen)
  2. Table-driven testing patterns (Go subtests with t.Run())
  3. Stack expansion strategy (Go slices with index-based access)
  4. Error context capture (Near: expression window, Where: call stack)
  5. UTF-8 string handling ([]rune for character series semantics)
  6. Interpreter benchmarking strategy (Go benchmarks with CI tracking)

**Decisions Made**:
- Primary dependency: github.com/chzyer/readline (pure Go, cross-platform)
- Testing approach: Go table-driven tests with subtests
- Performance targets: Stack ops <100ns, simple eval <10ms, complex eval <100ms
- **Local-by-default scoping**: Viro diverges from Viro's global-by-default semantics. All words in function bodies are local by default, preventing accidental modification of global state. This design is safer and more predictable than Viro's approach, aligning with modern language best practices (JavaScript let/const, Python local variables, etc.).
- **Bash-style refinements**: Viro replaces Viro's `/refinement` syntax with bash-style `--flag` and `--option []` syntax. This makes function calls more intuitive and self-documenting: it's immediately clear which refinements are boolean flags vs which accept values. The syntax is familiar to CLI users and eliminates confusion about refinement argument order. Example: `copy data --deep --limit 5` is clearer than `copy/deep/limit data 5`.
- **Simplified control flow**: Viro replaces Viro's confusing if/either distinction with a clearer when/if pair. `when condition [block]` executes a single branch (returns none if condition is false), while `if condition [true-block] [false-block]` requires both branches (like most mainstream languages). This eliminates the "why are there two conditional functions?" confusion and makes code intent clearer. Developers coming from other languages will immediately understand: `when` = optional single branch, `if` = mandatory both branches.
- **Paren type for immediate evaluation**: Viro includes paren `(...)` as a distinct type from block `[...]`. Blocks evaluate to themselves (deferred evaluation - data or code storage), while parens evaluate their contents immediately and return the result. This distinction is critical for: (1) controlling evaluation order in left-to-right evaluation `x: 3 + 4 * 2` (= 14) vs `x: 3 + (4 * 2)` (= 11), (2) dynamic path indices `array.(index + 1)`, (3) forcing evaluation in literal contexts `data: [name "Alice" age (current-year - birth-year)]`, and (4) computed refinement names `func (get refinement-word)`. Most expressions don't need parens—they evaluate naturally. Parens are only required when you need immediate evaluation in a deferred context (inside blocks) or to control evaluation order. Without parens, users would need verbose `do` calls for these cases: `data: [age do [current-year - birth-year]]`. Paren provides Viro's elegant immediate-evaluation syntax while keeping blocks as pure data/code containers.
- **Left-to-right evaluation**: Viro follows Viro's strict left-to-right evaluation model with **no operator precedence**. This critical design decision embraces Viro's philosophy: `3 + 4 * 2` evaluates to `14` (not 11), processing operators as they appear from left to right. This is simpler to implement and matches Viro's homoiconic purity. Users can control evaluation order with parentheses: `3 + (4 * 2)` = 11. The parser does not need precedence levels—it simply transforms infix notation to prefix calls in left-to-right order.
- **Integer-only arithmetic (Phase 1 scope)**: Viro implements **64-bit signed integer arithmetic only** in Phase 1, deferring decimal/floating-point types to Phase 2. This reduces initial implementation complexity and allows focus on proving the core architecture: type-based dispatch, stack/frame management, native function evaluation, and simple left-to-right parser. Decimal types introduce multiple design questions that are better addressed after core architecture is validated: (1) precision - float32 or float64?, (2) type coercion - should `3 + 4.5` be allowed and what's the result type?, (3) division semantics - should `5 / 2` return 2 (integer) or 2.5 (decimal)?, (4) special values - how to handle NaN, Infinity, -Infinity?, (5) string formatting - decimal places, scientific notation. Phase 1 uses truncating integer division: `10 / 3` → `3`, `-10 / 3` → `-3` (toward zero). Users needing decimal arithmetic can revisit this design in Phase 2 with full consideration of precision, coercion, and special value semantics.

**Next**: Phase 1 - Design & Contracts

---

## Phase 1: Design & Contracts (Complete)

✅ **Completed**: All design artifacts generated

**Design Artifacts**:
- `data-model.md`: Complete entity definitions for 9 core entities (Value, Block, Paren, Word, Function, Frame, Stack, Error, Series) with fields, validation rules, state transitions, relationships, and data flow examples
- `contracts/README.md`: Overview of 29 native functions organized by category
- `contracts/control-flow.md`: Contracts for if, when, loop, while
- `contracts/data.md`: Contracts for set, get, type?
- `contracts/io.md`: Contracts for print, input
- `contracts/math.md`: Contracts for +, -, *, /, <, >, <=, >=, =, <>, and, or, not
- `contracts/series.md`: Contracts for first, last, append, insert, length?
- `contracts/function.md`: Contract for fn (user-defined functions)
- `contracts/error-handling.md`: Error structure, categories (0-900), construction, propagation
- `quickstart.md`: Build, run, test instructions with REPL examples and development workflow

**Agent Context**:
- Updated `.github/copilot-instructions.md` with Go 1.21+, chzyer/readline, testing strategy

**Next**: Phase 2 - Task Breakdown (via `/speckit.tasks` command)

---

## Phase 2: Re-evaluation of Constitution Check

*Re-check after Phase 1 design to ensure no violations introduced*

✅ **Principle I - TDD**: Contracts define test cases before implementation. Quickstart documents TDD workflow.

✅ **Principle II - Incremental Layering**: Data model and contracts organized by architecture layer (Core → Types → Natives → Frames → Errors).

✅ **Principle III - Type Dispatch Fidelity**: Data model defines Value type with type tag and type-based dispatch via evaluator.

✅ **Principle IV - Stack and Frame Safety**: Data model specifies index-based access for Stack and Frame, avoiding pointer invalidation.

✅ **Principle V - Structured Errors**: Error contracts define 7 categories (0-900) with required fields (code, type, id, args, near, where).

✅ **Principle VI - Observable Behavior**: Print/input natives provide I/O. Error messages include near/where context per contracts.

✅ **Principle VII - YAGNI**: Contracts cover 29 natives (minimal set), parse dialect and advanced features deferred per spec out-of-scope.

**Gate Status**: ✅ PASSED - All constitutional requirements remain satisfied after design phase.

---

## Planning Summary

**Branch**: `001-implement-the-core`  
**Status**: Planning Complete, Ready for Task Breakdown

**Completed Phases**:
- ✅ Phase 0: Research (6 technical decisions resolved)
- ✅ Phase 1: Design & Contracts (10 artifacts generated)

**Generated Artifacts**:
```
specs/001-implement-the-core/
├── plan.md              ✅ This file (implementation plan)
├── research.md          ✅ Technical research and decisions
├── data-model.md        ✅ Entity definitions and relationships
├── quickstart.md        ✅ Build, run, test guide
└── contracts/           ✅ Native function specifications
    ├── README.md        ✅ Overview (29 natives)
    ├── control-flow.md  ✅ if, when, loop, while
    ├── data.md          ✅ set, get, type?
    ├── io.md            ✅ print, input
    ├── math.md          ✅ 13 math/logic operations
    ├── series.md        ✅ first, last, append, insert, length?
    ├── function.md      ✅ func (user-defined functions)
    └── error-handling.md ✅ Error structure and categories
```

**Agent Context Updated**: ✅ .github/copilot-instructions.md

**Constitution Compliance**: ✅ All 7 principles verified (pre- and post-design)

**Next Command**: `/speckit.tasks` to generate granular task breakdown with test-first ordering

---

## Notes for Implementation

**Architecture Sequence** (per constitution Principle II):
1. Core Infrastructure (Value, Type, Error, Stack, Frame)
2. Evaluator Foundation (Do_Next, Do_Blk, type dispatch)
3. Native Functions (Data → Math → Series → I/O → Control Flow → Function)
4. REPL Integration (readline, history, multi-line, interrupt)
5. Testing & Validation (contract tests, integration tests, benchmarks)

**TDD Enforcement** (per constitution Principle I):
- Every task must include: Write test → Verify failure → Implement → Verify pass
- Contract tests define expected behavior before implementation
- Integration tests validate end-to-end scenarios

**Key Dependencies**:
- `github.com/chzyer/readline`: REPL command history and multi-line input
- Go 1.21+: Generics for type-safe value handling, improved error handling

**Performance Targets**:
- Stack push/pop: <100 ns/op (verified via benchmarks)
- Native function call: <1 µs/op
- Simple expression evaluation: <10 ms (per spec SC-005)
- Complex expression evaluation: <100 ms (per spec SC-005)

**Quality Gates**:
- 29 natives fully implemented per contracts
- All contract tests passing (100%)
- Integration tests passing for user stories P1-P6
- Code coverage ≥80% for core systems
- Success criteria SC-001 through SC-010 met
- REPL features operational (history, multi-line, interrupt)

**Ready for Task Breakdown**: All planning complete, proceed to `/speckit.tasks` command.
