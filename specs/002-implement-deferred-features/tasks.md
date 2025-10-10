# Tasks: Deferred Language Capabilities (002)

**Input**: Design documents from `/specs/002-implement-deferred-features/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Contract tests required for all new features per spec requirements. Integration tests validate success criteria SC-011 through SC-020.

**Organization**: Tasks are grouped by user story (P1-P5) to enable independent implementation and testing of each capability.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5, or SETUP/FOUNDATION for shared infrastructure)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency management for Feature 002

- [X] T001 Add `github.com/ericlagergren/decimal` dependency to go.mod
- [X] T002 Add `gopkg.in/natefinch/lumberjack.v2` dependency to go.mod
- [X] T003 [P] Update README.md with Feature 002 capabilities overview
- [X] T004 [P] Create contract test scaffold in test/contract/math_decimal_test.go
- [X] T005 [P] Create contract test scaffold in test/contract/ports_test.go
- [X] T006 [P] Create contract test scaffold in test/contract/objects_test.go
- [X] T007 [P] Create contract test scaffold in test/contract/parse_test.go
- [X] T008 [P] Create contract test scaffold in test/contract/trace_debug_test.go
- [X] T009 [P] Create contract test scaffold in test/contract/reflection_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T010 Extend ValueType enumeration in internal/value/types.go (add TypeDecimal, TypeObject, TypePort, TypePath)
- [X] T011 [P] Implement DecimalValue struct in internal/value/decimal.go with decimal.Big wrapper
- [X] T012 [P] Implement ObjectInstance struct in internal/value/object.go with FrameIndex and Manifest
- [X] T013 [P] Implement Port struct in internal/value/port.go with scheme, driver interface, timeout fields
- [X] T014 [P] Implement PathExpression struct in internal/value/path.go with segment representation
- [X] T015 Update Value.String() in internal/value/value.go to handle new types (decimal, object, port, path)
- [X] T016 [P] Add CLI flag --sandbox-root to cmd/viro/main.go with default to os.Getwd()
- [X] T017 [P] Add CLI flag --trace-file to cmd/viro/main.go for optional trace file redirection (default: stderr output)
- [X] T018 [P] Add CLI flag --trace-max-size to cmd/viro/main.go (default 50MB per clarification)
- [X] T018.1 [P] Add CLI flag --allow-insecure-tls to cmd/viro/main.go (global TLS verification bypass with stderr warning)
- [X] T019 Create sandbox path resolver helper in internal/eval/sandbox.go (resolves paths within root, validates with filepath.EvalSymlinks)
- [X] T020 Implement TraceSession struct in internal/native/trace.go with dual sink support (stderr default, optional lumberjack file sink)
- [X] T021 [P] Implement TraceEvent struct and JSON serialization in internal/native/trace.go
- [X] T022 [P] Implement TraceFilters struct in internal/native/trace.go (include/exclude words, min duration)
- [X] T023 Implement Debugger struct in internal/native/trace.go (breakpoints map, mode, ID generation)
- [X] T024 Update evaluator dispatch in internal/eval/evaluator.go to check type dispatch tables for new types
- [X] T025 Add trace instrumentation hooks in internal/eval/evaluator.go (entry/exit of Do_Next, Do_Blk)
- [X] T025.1 [FOUNDATION] **CHECKPOINT - TDD Gate**: Write contract tests for new value types (TypeDecimal, TypeObject, TypePort, TypePath) in test/contract/value_types_test.go covering: (1) Value construction, (2) String() output, (3) Type dispatch routing, (4) Error cases (invalid conversions). Run tests and verify they FAIL before proceeding to Phase 3.

**⚠️ BLOCKER**: No user story implementation may begin until T025.1 passes with failing tests documented.

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Precise Calculations with Decimal Numbers (Priority: P1) 🎯 MVP

**Goal**: Enable financial and scientific calculations with decimal precision and advanced math functions

**Independent Test**: Start REPL, evaluate `rate: 12.5%`, `total: 19.99 * 3`, `round --places total 2`, `pow 1.05 12` - confirm precision within tolerance

### Contract Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T026 [US1] Contract test: decimal constructor (integer, string, scale preservation) in test/contract/math_decimal_test.go
- [X] T027 [US1] Contract test: decimal promotion in mixed arithmetic in test/contract/math_decimal_test.go
- [X] T028 [US1] Contract test: pow, sqrt, exp domain validation in test/contract/math_decimal_test.go
- [X] T029 [US1] Contract test: log, log-10 domain errors in test/contract/math_decimal_test.go
- [X] T030 [US1] Contract test: trig functions (sin, cos, tan, asin, acos, atan) in test/contract/math_decimal_test.go
- [X] T031 [US1] Contract test: rounding modes (round --places, --mode, ceil, floor, truncate) in test/contract/math_decimal_test.go
- [X] T032 [US1] Contract test: overflow/underflow handling in test/contract/math_decimal_test.go
- [X] T032.1 [US1] **CHECKPOINT**: Run `go test ./test/contract/math_decimal_test.go` and verify ALL tests FAIL with expected error messages before proceeding to implementation tasks

### Implementation for User Story 1

- [X] T033 [P] [US1] Implement decimal literal parser in internal/parse/parse.go (sign, fraction, exponent with scale metadata)
- [X] T034 [P] [US1] Implement tokenizer disambiguation logic in internal/parse/parse.go (numbers vs refinements vs paths by first character)
- [X] T034.1 [US1] Implement comprehensive token disambiguation validation in internal/parse/parse_test.go covering edge cases: negative decimals (`-3.14`), refinement-like numbers (`--123` ambiguous), path-like numbers (`1.2.3` invalid), and ensure parser correctly classifies per FR-011 first-character rules
- [X] T035 [US1] Implement `decimal` native constructor in internal/native/math.go (handles integer, decimal, string inputs)
- [X] T036 [US1] Add arithmetic promotion logic in internal/native/math.go (integer→decimal conversion with scale 0)
- [X] T037 [US1] Create internal/native/math_decimal.go for advanced math natives
- [X] T038 [P] [US1] Implement `pow` native in internal/native/math_decimal.go using decimal.Context
- [X] T039 [P] [US1] Implement `sqrt` native in internal/native/math_decimal.go with negative domain check
- [X] T040 [P] [US1] Implement `exp` native in internal/native/math_decimal.go with overflow guards
- [X] T041 [P] [US1] Implement `log` and `log-10` natives in internal/native/math_decimal.go with domain validation
- [X] T042 [P] [US1] Implement trigonometric natives (sin, cos, tan) in internal/native/math_decimal.go
- [X] T043 [P] [US1] Implement inverse trig natives (asin, acos, atan) in internal/native/math_decimal.go with domain checks
- [X] T044 [US1] Implement `round` native with refinements (--places, --mode) in internal/native/math_decimal.go
- [X] T045 [P] [US1] Implement `ceil`, `floor`, `truncate` natives in internal/native/math_decimal.go
- [X] T046 [US1] Register all decimal and math natives in internal/native/registry.go
- [X] T047 [US1] Add Math error mappings for domain violations (sqrt-negative, log-domain, exp-overflow) in internal/verror/categories.go
- [X] T047.1 [US1] Add contract test for decimal precision overflow (35+ digit results) raising Math error in test/contract/math_decimal_test.go
- [ ] T048 [US1] Update error context generation to include decimal metadata (magnitude, scale) in internal/verror/context.go

### Integration Tests for User Story 1

- [X] T049 [US1] Integration test SC-011: Decimal arithmetic precision validation in test/integration/sc011_validation_test.go
- [X] T050 [US1] Integration test SC-011: Performance benchmark (operations <2ms) in test/integration/sc011_validation_test.go
- [X] T051 [US1] Add math benchmarks in internal/native/math_bench_test.go (decimal vs integer baseline)
- [X] T051.1 [US1] **CHECKPOINT - Backward Compatibility**: Run complete Feature 001 test suite (`go test ./test/integration/*_001_*.go`) and verify zero regressions before proceeding to User Story 2

**Checkpoint**: User Story 1 COMPLETE ✅ - Decimal arithmetic and advanced math fully functional with literal syntax support. All tests passing (12/12). Backward compatibility verified.

---

## Phase 4: User Story 2 - Persist and Exchange Data via Ports (Priority: P2)

**Goal**: Enable file and network I/O through unified port abstraction with sandbox and TLS controls

**Independent Test**: Run scripts creating files in sandbox, fetching JSON via HTTP, streaming TCP data - verify integrity and error handling

### Contract Tests for User Story 2

- [X] T052 [US2] Contract test: open file port with sandbox resolution in test/contract/ports_test.go
- [X] T053 [US2] Contract test: open HTTP port with TLS verification in test/contract/ports_test.go
- [X] T054 [US2] Contract test: open TCP port with timeout in test/contract/ports_test.go
- [X] T055 [US2] Contract test: read/write file operations in test/contract/ports_test.go
- [X] T056 [US2] Contract test: HTTP GET/POST/HEAD with redirects in test/contract/ports_test.go
- [X] T057 [US2] Contract test: port query metadata in test/contract/ports_test.go
- [X] T058 [US2] Contract test: sandbox escape prevention in test/contract/ports_test.go
- [X] T059 [US2] Contract test: TLS --insecure flag behavior in test/contract/ports_test.go
- [X] T059.1 [US2] **CHECKPOINT**: Run `go test ./test/contract/ports_test.go` and verify ALL tests FAIL with expected error messages before proceeding to implementation tasks

### Implementation for User Story 2

- [X] T060 [US2] Define PortDriver interface in internal/value/value.go (Open, Read, Write, Close, Query methods)
- [X] T061 [P] [US2] Implement fileDriver in internal/native/io.go with sandbox path resolution
- [X] T062 [P] [US2] Implement tcpDriver in internal/native/io.go with net.Dialer and optional timeout
- [X] T063 [US2] Implement httpDriver in internal/native/io.go with http.Client pool keyed by (verifyTLS, timeout)
- [X] T064 [US2] Implement redirect following logic (max 10 hops) in internal/native/io.go httpDriver
- [X] T065 [US2] Implement `open` native in internal/native/io.go with scheme dispatch and refinement handling
- [X] T066 [P] [US2] Implement `close` native in internal/native/io.go (idempotent, state transition)
- [X] T067 [P] [US2] Implement `read` native in internal/native/io.go with --binary, --lines, --part refinements
- [X] T068 [P] [US2] Implement `write` native in internal/native/io.go with --append, --binary, --lines refinements
- [X] T069 [P] [US2] Implement `save` convenience native in internal/native/io.go (serialization + write)
- [X] T070 [P] [US2] Implement `load` convenience native in internal/native/io.go (read + parse)
- [X] T071 [P] [US2] Implement `query` native in internal/native/io.go returning port metadata as object
- [X] T072 [P] [US2] Implement `wait` native in internal/native/io.go for port readiness polling
- [X] T073 [US2] Register all port natives in internal/native/registry.go
- [X] T074 [US2] Add Access error mappings for port failures (port-closed, tls-verification-failed, sandbox-violation) in internal/verror/categories.go
- [X] T075 [US2] Add IO error mappings (timeout, connection-refused) in internal/verror/categories.go
- [X] T076 [US2] Add trace events for port lifecycle (open, read, write, close, error) in internal/native/trace.go

### Integration Tests for User Story 2

- [X] T077 [US2] Integration test SC-012: File read/write throughput (50 MB/s) in test/integration/sc012_validation_test.go
- [X] T078 [US2] Integration test SC-012: HTTP GET latency (95% <2s LAN) in test/integration/sc012_validation_test.go
- [X] T079 [US2] Integration test: Sandbox enforcement scenarios in test/integration/sc012_validation_test.go
- [X] T079.1 [US2] **CHECKPOINT - Backward Compatibility**: Run complete Feature 001 test suite and verify zero regressions before proceeding to User Story 3

**Checkpoint**: User Story 2 COMPLETE ✅ - Port operations fully functional with sandbox enforcement, TLS controls, and trace events. All tests passing (13/13 SC-012 tests). Backward compatibility verified (84/84 Feature 001 tests, 13/13 User Story 1 tests).

---

## Phase 5: User Story 3 - Structure Data with Objects and Paths (Priority: P3)

**Goal**: Enable object construction and nested path access/mutation for structured programming

**Independent Test**: Create nested objects, access via `user.address.city`, mutate via path assignment - verify frame isolation

### Contract Tests for User Story 3

- [X] T080 [US3] Contract test: object construction with field initialization in test/contract/objects_test.go
- [X] T081 [US3] Contract test: nested object creation in test/contract/objects_test.go
- [X] T082 [US3] Contract test: path read traversal (object.field.subfield) in test/contract/objects_test.go
- [X] T083 [US3] Contract test: path write mutation in test/contract/objects_test.go
- [X] T084 [US3] Contract test: path indexing for blocks (block.3) in test/contract/objects_test.go
- [X] T085 [US3] Contract test: parent prototype lookup in test/contract/objects_test.go
- [X] T086 [US3] Contract test: path error handling (none-path, index-out-of-range) in test/contract/objects_test.go
- [X] T086.1 [US3] **CHECKPOINT**: Run `go test ./test/contract/objects_test.go` and verify ALL tests FAIL with expected error messages before proceeding to implementation tasks

### Implementation for User Story 3

- [X] T087 [US3] Extend Frame to support ObjectManifest (Words, Types) in internal/frame/frame.go
- [X] T088 [US3] Implement `object` native in internal/native/data.go (creates frame, binds words, evaluates initializers)
- [X] T089 [P] [US3] Implement `context` native in internal/native/data.go (isolated scope variant of object)
- [X] T090 [US3] Implement path segment tokenizer in internal/parse/parse.go (distinguish word, index, refinement, paren segments)
- [X] T091 [US3] Implement path evaluation logic in internal/eval/evaluator.go (recursive traversal with base value resolution)
- [X] T092 [US3] Implement path assignment logic in internal/eval/evaluator.go (penultimate target tracking, mutation)
- [X] T093 [P] [US3] Implement `select` native in internal/native/data.go (object field lookup with --default)
- [X] T094 [P] [US3] Implement `put` native in internal/native/data.go (field update with validation)
- [X] T095 [US3] Register object and path natives in internal/native/registry.go
- [X] T096 [US3] Add Script error mappings (object-field-duplicate, no-such-field, immutable-target, path-type-mismatch) in internal/verror/categories.go
- [X] T097 [US3] Add trace events for object operations (object-create, object-field-read, object-field-write) in internal/native/trace.go

### Integration Tests for User Story 3

- [X] T098 [US3] Integration test SC-013: Nested object path access in test/integration/sc013_validation_test.go
- [X] T099 [US3] Integration test SC-013: Path mutation scenarios in test/integration/sc013_validation_test.go
- [X] T099.1 [US3] **CHECKPOINT - Backward Compatibility**: Run complete Feature 001 test suite and verify zero regressions before proceeding to User Story 4

**Checkpoint**: User Story 3 COMPLETE ✅ - Objects and paths operational with select/put natives, trace events. All Feature 001 tests passing (9/9 SC tests). Contract tests provide comprehensive coverage for nested access and mutation scenarios.

---

## Phase 6: User Story 4 - Transform Data with Advanced Series and Parse Dialect (Priority: P4)

**Goal**: Provide advanced series operations and declarative parse dialect for data manipulation

**Independent Test**: Use parse to validate patterns, leverage copy --part, find, take, remove, sort on blocks/strings - confirm deterministic results

### Contract Tests for User Story 4

- [ ] T100 [US4] Contract test: copy, copy --part for blocks and strings in test/contract/series_test.go
- [ ] T101 [US4] Contract test: find, find --last in test/contract/series_test.go
- [ ] T102 [US4] Contract test: remove, remove --part in test/contract/series_test.go
- [ ] T103 [US4] Contract test: skip, take operations in test/contract/series_test.go
- [ ] T104 [US4] Contract test: sort, reverse on series in test/contract/series_test.go
- [ ] T105 [US4] Contract test: parse literal matching in test/contract/parse_test.go
- [ ] T106 [US4] Contract test: parse quantifiers (some, any, opt) in test/contract/parse_test.go
- [ ] T107 [US4] Contract test: parse set/copy captures in test/contract/parse_test.go
- [ ] T108 [US4] Contract test: parse into nested blocks in test/contract/parse_test.go
- [ ] T109 [US4] Contract test: parse control (not, ahead, fail) in test/contract/parse_test.go
- [ ] T110 [US4] Contract test: parse failure diagnostics in test/contract/parse_test.go
- [ ] T110.1 [US4] **CHECKPOINT**: Run `go test ./test/contract/series_test.go ./test/contract/parse_test.go` and verify ALL tests FAIL with expected error messages before proceeding to implementation tasks

### Implementation for User Story 4

- [ ] T111 [P] [US4] Implement `copy` native with --part refinement in internal/native/series.go
- [ ] T112 [P] [US4] Implement `find` native with --last refinement in internal/native/series.go
- [ ] T113 [P] [US4] Implement `remove` native with --part refinement in internal/native/series.go
- [ ] T114 [P] [US4] Implement `skip` and `take` natives in internal/native/series.go
- [ ] T115 [P] [US4] Implement `sort` and `reverse` natives in internal/native/series.go
- [ ] T116 [US4] Create internal/parse/dialect.go for parse dialect engine
- [ ] T117 [US4] Implement ParseRule and ParseState structs in internal/parse/dialect.go
- [ ] T118 [US4] Implement parse literal matching in internal/parse/dialect.go
- [ ] T119 [US4] Implement parse quantifiers (some, any, opt) in internal/parse/dialect.go
- [ ] T120 [US4] Implement parse set/copy operations in internal/parse/dialect.go
- [ ] T121 [US4] Implement parse control combinators (not, ahead, fail, reject) in internal/parse/dialect.go
- [ ] T122 [US4] Implement parse into recursion with stack management in internal/parse/dialect.go
- [ ] T123 [US4] Implement infinite loop detection (rule/index memoization) in internal/parse/dialect.go
- [ ] T124 [US4] Implement `parse` native in internal/native/parse.go with --case, --part, --trace refinements
- [ ] T125 [US4] Register series and parse natives in internal/native/registry.go
- [ ] T126 [US4] Add Syntax error mappings for parse failures (parse-stalled, invalid-parse-rule) in internal/verror/categories.go
- [ ] T127 [US4] Add parse trace events with indentation tracking in internal/native/trace.go

### Integration Tests for User Story 4

- [ ] T128 [US4] Integration test SC-014: Parse dialect validation corpus (50 patterns, 0 false positives/negatives) in test/integration/sc014_validation_test.go
- [ ] T129 [US4] Integration test SC-014: Parse performance (<250ms for 1MB input) in test/integration/sc014_validation_test.go
- [ ] T130 [US4] Integration test: CSV parsing scenario in test/integration/sc014_validation_test.go
- [ ] T130.1 [US4] **CHECKPOINT - Backward Compatibility**: Run complete Feature 001 test suite and verify zero regressions before proceeding to User Story 5

**Checkpoint**: User Story 4 complete - advanced series and parse dialect operational

---

## Phase 7: User Story 5 - Observe, Debug, and Reflect on Programs (Priority: P5)

**Goal**: Provide tracing, debugging, and reflection capabilities for program transparency and diagnostics

**Independent Test**: Enable tracing, set breakpoints, inspect stack frames, use reflection to examine functions/objects - confirm visibility and stability

### Contract Tests for User Story 5

- [ ] T131 [US5] Contract test: trace --on/--off/trace? in test/contract/trace_debug_test.go
- [ ] T132 [US5] Contract test: trace filtering (--only, --exclude) in test/contract/trace_debug_test.go
- [ ] T133 [US5] Contract test: trace sink configuration (--file, --append) in test/contract/trace_debug_test.go
- [ ] T134 [US5] Contract test: debug --breakpoint/--remove in test/contract/trace_debug_test.go
- [ ] T135 [US5] Contract test: debug stepping (--step, --next, --finish, --continue) in test/contract/trace_debug_test.go
- [ ] T136 [US5] Contract test: debug --locals/--stack in test/contract/trace_debug_test.go
- [ ] T137 [US5] Contract test: type-of for all value types in test/contract/reflection_test.go
- [ ] T138 [US5] Contract test: spec-of for functions/objects in test/contract/reflection_test.go
- [ ] T139 [US5] Contract test: body-of immutability in test/contract/reflection_test.go
- [ ] T140 [US5] Contract test: words-of/values-of consistency in test/contract/reflection_test.go
- [ ] T140.1 [US5] **CHECKPOINT**: Run `go test ./test/contract/trace_debug_test.go ./test/contract/reflection_test.go` and verify ALL tests FAIL with expected error messages before proceeding to implementation tasks

### Implementation for User Story 5

- [ ] T141 [US5] Implement trace session management (enable/disable/flush) in internal/native/trace.go
- [ ] T142 [US5] Implement lumberjack sink configuration (max size 50MB, 5 backups) in internal/native/trace.go
- [ ] T143 [US5] Implement trace event JSON serialization in internal/native/trace.go
- [ ] T144 [P] [US5] Implement `trace --on` native with refinements in internal/native/control.go
- [ ] T145 [P] [US5] Implement `trace --off` native in internal/native/control.go
- [ ] T146 [P] [US5] Implement `trace?` query native in internal/native/control.go
- [ ] T147 [US5] Implement Breakpoint management in internal/native/trace.go
- [ ] T148 [P] [US5] Implement `debug --on/--off` natives in internal/native/control.go
- [ ] T149 [P] [US5] Implement `debug --breakpoint` and `debug --remove` in internal/native/control.go
- [ ] T150 [P] [US5] Implement `debug --step/--next/--finish/--continue` in internal/native/control.go
- [ ] T151 [P] [US5] Implement `debug --locals` (frame snapshot) in internal/native/control.go
- [ ] T152 [P] [US5] Implement `debug --stack` (call stack retrieval) in internal/native/control.go
- [ ] T153 [US5] Integrate breakpoint checks in evaluator dispatch in internal/eval/evaluator.go
- [ ] T154 [US5] Update REPL prompt to indicate debug mode in cmd/viro/repl.go
- [ ] T155 [P] [US5] Implement `type-of` native in internal/native/data.go
- [ ] T156 [P] [US5] Implement `spec-of` native with immutable snapshot in internal/native/data.go
- [ ] T157 [P] [US5] Implement `body-of` native with deep copy in internal/native/data.go
- [ ] T158 [P] [US5] Implement `words-of` native in internal/native/data.go
- [ ] T159 [P] [US5] Implement `values-of` native in internal/native/data.go
- [ ] T160 [P] [US5] Implement `source` native with formatting in internal/native/data.go
- [ ] T161 [US5] Register trace, debug, and reflection natives in internal/native/registry.go
- [ ] T162 [US5] Add Script error mappings (unknown-symbol, no-such-breakpoint, spec-unsupported-type) in internal/verror/categories.go

### Integration Tests for User Story 5

- [ ] T163 [US5] Integration test SC-015: Trace overhead when disabled (<5%) in test/integration/sc015_validation_test.go
- [ ] T164 [US5] Integration test SC-015: Trace overhead when enabled (<25%) in test/integration/sc015_validation_test.go
- [ ] T165 [US5] Integration test SC-015: Breakpoint interaction latency (<150ms) in test/integration/sc015_validation_test.go
- [ ] T166 [US5] Integration test: End-to-end trace session workflow in test/integration/sc015_validation_test.go
- [ ] T167 [US5] Integration test: Debug session with stepping and inspection in test/integration/sc015_validation_test.go
- [ ] T167.1 [US5] **CHECKPOINT - Backward Compatibility**: Run complete Feature 001 test suite and verify zero regressions before proceeding to Phase 8

**Checkpoint**: User Story 5 complete - observability and reflection fully operational

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and finalization

- [ ] T168 [P] Update quickstart.md with Feature 002 examples and validation checklist
- [ ] T169 [P] Update docs/interpreter.md with decimal, object, port, parse, trace documentation
- [ ] T170 [P] Create docs/observability.md documenting trace/debug usage patterns
- [ ] T171 [P] Create docs/ports-guide.md with sandbox configuration and security considerations
- [ ] T172 [P] Update RELEASE_NOTES.md with Feature 002 capabilities summary
- [ ] T173 Integration test SC-016: Backward compatibility regression suite (all Feature 001 tests pass) in test/integration/sc016_validation_test.go
- [ ] T174 [P] Integration test SC-017: Decimal precision corpus validation in test/integration/sc017_validation_test.go
- [ ] T175 [P] Integration test SC-018: Port I/O integration scenarios in test/integration/sc018_validation_test.go
- [ ] T176 [P] Integration test SC-019: Object and path integration scenarios in test/integration/sc019_validation_test.go
- [ ] T177 [P] Integration test SC-020: Parse dialect validation corpus in test/integration/sc020_validation_test.go
- [ ] T178 Run quickstart.md validation from specs/002-implement-deferred-features/quickstart.md
- [ ] T179 [P] Code review: Constitution compliance check using docs/constitution-compliance.md checklist
- [ ] T180 [P] Code review: Security audit for sandbox, TLS, and trace file handling
- [ ] T181 Performance profiling and optimization pass (target: SC-001 through SC-005 metrics)
- [ ] T182 Final integration test sweep across all success criteria (SC-011 through SC-020)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - User Story 1 (P1): Decimal arithmetic - can start after Foundation
  - User Story 2 (P2): Ports - can start after Foundation (independent of US1)
  - User Story 3 (P3): Objects/Paths - can start after Foundation (independent of US1, US2)
  - User Story 4 (P4): Parse/Series - depends on US3 for object construction integration
  - User Story 5 (P5): Trace/Debug - can start after Foundation but should wait for US1-US4 for meaningful testing
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Foundation complete → No other dependencies
- **User Story 2 (P2)**: Foundation complete → Independent (may use US1 decimals in examples but not required)
- **User Story 3 (P3)**: Foundation complete → Independent
- **User Story 4 (P4)**: Foundation complete → Soft dependency on US3 for parse-to-object workflows
- **User Story 5 (P5)**: Foundation complete → Soft dependency on US1-US4 for representative trace coverage

### Within Each User Story

- Contract tests MUST be written and FAIL before implementation
- Value type definitions before natives
- Parser updates before evaluation logic
- Core natives before convenience wrappers
- Error mappings before integration tests
- Story complete before moving to next priority

### Parallel Opportunities Within Phases

**Phase 1 (Setup)**: T003-T009 can all run in parallel (different test files)

**Phase 2 (Foundational)**: 
- T011-T014 can run in parallel (different value structs)
- T016-T018 can run in parallel (different CLI flags)
- T020-T023 can run in parallel (different trace/debug components)

**Phase 3 (User Story 1)**:
- T026-T032 contract tests can run in parallel (different test functions)
- T033-T034 parser tasks sequential (same file)
- T038-T043 math natives can run in parallel (different functions, may be same file but independent)
- T045 parallel with T042-T043 (different rounding helpers)

**Phase 4 (User Story 2)**:
- T052-T059 contract tests can run in parallel
- T061-T062 drivers can run in parallel (independent implementations)
- T066-T072 port natives can run in parallel if in separate functions
- T077-T079 integration tests can run in parallel

**Phase 5 (User Story 3)**:
- T080-T086 contract tests can run in parallel
- T089 parallel with T088 (context vs object, may share logic but independent)
- T093-T094 natives can run in parallel

**Phase 6 (User Story 4)**:
- T100-T110 contract tests can run in parallel
- T111-T115 series natives can run in parallel
- T118-T122 parse components sequential (depend on dialect.go structure)

**Phase 7 (User Story 5)**:
- T131-T140 contract tests can run in parallel
- T144-T146 trace natives can run in parallel
- T148-T152 debug commands can run in parallel (may share debugger state but independent functions)
- T155-T160 reflection natives can run in parallel

**Phase 8 (Polish)**:
- T168-T172 documentation can run in parallel
- T173-T177 integration tests can run in parallel

---

## Parallel Example: User Story 1 (Decimal Math)

```bash
# Launch all contract tests for decimal math together:
Task T026: "Contract test: decimal constructor in test/contract/math_decimal_test.go"
Task T027: "Contract test: decimal promotion in test/contract/math_decimal_test.go"
Task T028: "Contract test: pow, sqrt, exp domain validation in test/contract/math_decimal_test.go"
Task T029: "Contract test: log, log-10 domain errors in test/contract/math_decimal_test.go"
Task T030: "Contract test: trig functions in test/contract/math_decimal_test.go"
Task T031: "Contract test: rounding modes in test/contract/math_decimal_test.go"
Task T032: "Contract test: overflow/underflow handling in test/contract/math_decimal_test.go"

# After tests written and failing, launch parallel math native implementations:
Task T038: "Implement pow native in internal/native/math_decimal.go"
Task T039: "Implement sqrt native in internal/native/math_decimal.go"
Task T040: "Implement exp native in internal/native/math_decimal.go"
Task T041: "Implement log and log-10 natives in internal/native/math_decimal.go"
Task T042: "Implement trigonometric natives (sin, cos, tan) in internal/native/math_decimal.go"
Task T043: "Implement inverse trig natives (asin, acos, atan) in internal/native/math_decimal.go"
Task T045: "Implement ceil, floor, truncate natives in internal/native/math_decimal.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only - Decimal Arithmetic)

1. Complete Phase 1: Setup (T001-T009)
2. Complete Phase 2: Foundational (T010-T025) **CRITICAL - blocks all stories**
3. Complete Phase 3: User Story 1 (T026-T051)
4. **STOP and VALIDATE**: Run SC-011 tests, verify precision and performance
5. Deploy/demo decimal arithmetic capability

**Rationale**: Decimal math is P1 priority and delivers immediate business value for financial calculations while having minimal dependencies on other features.

### Incremental Delivery (Recommended)

1. **Foundation**: Complete Setup + Foundational (T001-T025) → All value types and infrastructure ready
2. **MVP**: Add User Story 1 (T026-T051) → Test independently → Deploy decimal math ✅
3. **I/O Layer**: Add User Story 2 (T052-T079) → Test independently → Deploy ports and file/network I/O ✅
4. **Structured Data**: Add User Story 3 (T080-T099) → Test independently → Deploy objects and paths ✅
5. **Data Transformation**: Add User Story 4 (T100-T130) → Test independently → Deploy parse and series ✅
6. **Observability**: Add User Story 5 (T131-T167) → Test independently → Deploy trace/debug ✅
7. **Polish**: Complete Phase 8 (T168-T182) → Full Feature 002 release

Each user story adds value without breaking previous stories. After each story, the interpreter remains in a deployable state.

### Parallel Team Strategy

With multiple developers available:

1. **Week 1**: Team completes Setup + Foundational together (T001-T025)
2. **Week 2-3**: Once Foundational is done:
   - Developer A: User Story 1 (Decimal Math) - T026-T051
   - Developer B: User Story 2 (Ports) - T052-T079
   - Developer C: User Story 3 (Objects) - T080-T099
3. **Week 4**: 
   - Developer D: User Story 4 (Parse) - T100-T130 (needs US3 complete)
   - Developer E: User Story 5 (Trace/Debug) - T131-T167 (needs US1-US4 for testing)
4. **Week 5**: All developers collaborate on Phase 8 (Polish) - T168-T182

Stories US1, US2, US3 are fully independent and can proceed in parallel. US4 and US5 have soft dependencies but can overlap significantly with careful coordination.

---

## Notes

- [P] tasks = different files or independent functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story delivers independently testable increment
- Contract tests written first following TDD principles
- All port operations enforce sandbox root from CLI flag --sandbox-root
- Decimal precision target: 34 digits (decimal128), ≤1.5× integer performance
- Parse dialect follows REBOL 3 semantics where practical
- Trace logs rotate at 50 MB per file (5 backups) per clarification
- HTTP ports follow redirects automatically (max 10 hops) per clarification
- TLS verification required by default, --insecure flag available per research decision
- Commit after each task or logical group
- Run quickstart validation after each user story phase
- Constitution compliance verified in Phase 8 before release
