# Tasks: Viro Core Language and REPL

**Feature Branch**: `001-implement-the-core`
**Input**: Design documents from `/specs/001-implement-the-core/`
**Prerequisites**: ✅ plan.md, ✅ spec.md, ✅ research.md, ✅ data-model.md, ✅ contracts/

**Methodology**: TDD (Test-Driven Development) - Tests written FIRST, implementation AFTER per Constitution Principle I

**Organization**: Tasks are grouped by user story (P1-P6) to enable independent implementation and testing of each story.

---

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US6, SETUP, FOUND, POLISH)
- Include exact file paths in descriptions

## Path Conventions
Single project structure (per plan.md):
- `internal/` - Implementation packages (value, eval, stack, frame, native, error)
- `cmd/viro/` - CLI entry point and REPL
- `test/contract/` - Contract tests for native functions
- `test/integration/` - End-to-end interpreter tests

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic Go structure

- [X] T001 [SETUP] Create project directory structure: `internal/{value,eval,stack,frame,native,verror}`, `cmd/viro/`, `test/{contract,integration,fixtures}`
- [X] T002 [SETUP] Initialize Go module with `go mod init` and add `github.com/chzyer/readline` dependency
- [X] T003 [P] [SETUP] Configure Go linting (golangci-lint) and formatting (gofmt)
- [X] T004 [P] [SETUP] Create `.gitignore` for Go artifacts (vendor/, *.test, build/)
- [X] T005 [P] [SETUP] Create `README.md` with project overview and quickstart link

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Value System (Foundation)

- [X] T006 [P] [FOUND] Define `ValueType` constants (TypeNone through TypeFunction) in `internal/value/types.go`
- [X] T007 [P] [FOUND] Define core `Value` struct with Type and Payload in `internal/value/value.go`
- [X] T008 [P] [FOUND] Implement value constructor functions (NoneVal, LogicVal, IntVal, StrVal, WordVal, BlockVal, ParenVal, FuncVal) in `internal/value/value.go`
- [X] T009 [P] [FOUND] Implement type assertion helpers (AsInteger, AsBlock, AsString, etc.) in `internal/value/value.go`

### Error System (Foundation)

- [X] T010 [P] [FOUND] Define `ErrorCategory` constants (0, 100, 200, 300, 400, 500, 900) in `internal/verror/categories.go`
- [X] T011 [P] [FOUND] Define `Error` struct with Category, Code, ID, Args, Near, Where, Message in `internal/verror/error.go`
- [X] T012 [P] [FOUND] Implement error factory functions (NewSyntaxError, NewScriptError, NewMathError, NewAccessError, NewInternalError) in `internal/verror/error.go`
- [X] T013 [P] [FOUND] Implement error message formatting with interpolation (%1, %2, %3) in `internal/verror/error.go`
- [X] T014 [P] [FOUND] Implement `Error.Error()` method for Go error interface in `internal/verror/error.go`

### Stack System (Foundation)

- [X] T015 [P] [FOUND] Define `Stack` struct with Data slice, Top, CurrentFrame in `internal/stack/stack.go`
- [X] T016 [P] [FOUND] Implement stack operations: Push, Pop, Get, Set with index-based access in `internal/stack/stack.go`
- [X] T017 [P] [FOUND] Implement automatic stack expansion using Go slice growth in `internal/stack/stack.go`
- [X] T018 [P] [FOUND] Implement stack frame layout helpers (NewFrame, DestroyFrame, GetFrame) in `internal/stack/frame.go`

### Frame System (Foundation)

- [X] T019 [P] [FOUND] Define `FrameType` constants (FrameFunctionArgs, FrameClosure) in `internal/frame/frame.go`
- [X] T020 [P] [FOUND] Define `Frame` struct with Type, Words, Values, Parent in `internal/frame/frame.go`
- [X] T021 [P] [FOUND] Implement frame operations: Bind, Get, Set, HasWord in `internal/frame/frame.go`

### Series Types (Foundation)

- [X] T022 [P] [FOUND] Define `BlockValue` struct with Elements slice and Index in `internal/value/block.go`
- [X] T023 [P] [FOUND] Implement block operations: First, Last, At, Length, Append, Insert in `internal/value/block.go`
- [X] T024 [P] [FOUND] Define `StringValue` struct (wrapping []rune) in `internal/value/string.go`
- [X] T025 [P] [FOUND] Implement string series operations (First, Last, Append, Insert, Length) in `internal/value/string.go`

### Word System (Foundation)

- [X] T026 [P] [FOUND] Define word type payloads (symbol string) for Word, SetWord, GetWord, LitWord in `internal/value/word.go`
- [X] T027 [P] [FOUND] Implement word constructor functions (WordVal, SetWordVal, GetWordVal, LitWordVal) in `internal/value/value.go`

### Function System (Foundation)

- [X] T028 [P] [FOUND] Define `FunctionType` constants (FuncNative, FuncUser) in `internal/value/function.go`
- [X] T029 [P] [FOUND] Define `ParamSpec` struct with Name, Type, Optional, Refinement, TakesValue in `internal/value/function.go`
- [X] T030 [P] [FOUND] Define `FunctionValue` struct with Type, Name, Params, Body, Native in `internal/value/function.go`

**Checkpoint**: Foundation complete - Core types, error system, stack, frames, series, and function structures ready. User story implementation can now begin.

---

## Phase 3: User Story 1 - Evaluate Basic Expressions (Priority: P1) 🎯 MVP

**Goal**: Users can evaluate simple REBOL-style expressions in the REPL: literals, words, arithmetic, blocks/parens

**Independent Test**: Start REPL, enter `5`, `3 + 4`, `x: 10`, `x`, `[1 + 2]`, `(1 + 2)` and verify correct output

### Contract Tests for User Story 1 (TDD - Write FIRST)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T031 [P] [US1] Contract test for literal evaluation (integers, strings, none, logic) in `test/contract/eval_test.go`
- [X] T032 [P] [US1] Contract test for word evaluation (undefined word errors) in `test/contract/eval_test.go`
- [X] T033 [P] [US1] Contract test for set-word assignment (`x: 10`) in `test/contract/eval_test.go`
- [X] T034 [P] [US1] Contract test for block evaluation (deferred - returns self) in `test/contract/eval_test.go`
- [X] T035 [P] [US1] Contract test for paren evaluation (immediate - evaluates contents) in `test/contract/eval_test.go`
- [X] T036 [P] [US1] Contract test for arithmetic natives (+, -, *, /) in `test/contract/math_test.go`
- [X] T037 [P] [US1] Contract test for operator precedence (3 + 4 * 2 → 11) in `test/contract/math_test.go`

### Implementation for User Story 1

#### Evaluator Core

- [X] T038 [US1] Implement `Evaluator` struct with stack, current frame, context in `internal/eval/evaluator.go`
- [X] T039 [US1] Implement `Do_Next` function (single value evaluation with type dispatch) in `internal/eval/evaluator.go`
- [X] T040 [US1] Implement `Do_Blk` function (block sequence evaluation) in `internal/eval/evaluator.go`
- [X] T041 [US1] Implement type-based dispatch router (switch on ValueType) in `internal/eval/evaluator.go`
- [X] T042 [US1] Implement literal evaluation (integers, strings, none, logic return self) in `internal/eval/evaluator.go`

#### Word Evaluation

- [X] T043 [US1] Implement word evaluation (lookup in frame, return bound value) in `internal/eval/evaluator.go`
- [X] T044 [US1] Implement set-word evaluation (evaluate next, store in frame) in `internal/eval/evaluator.go`
- [X] T045 [US1] Implement get-word evaluation (fetch without evaluation) in `internal/eval/evaluator.go`
- [X] T046 [US1] Implement lit-word evaluation (return word itself) in `internal/eval/evaluator.go`

#### Block & Paren Evaluation

- [X] T047 [US1] Implement block evaluation (returns self - deferred evaluation) in `internal/eval/evaluator.go`
- [X] T048 [US1] Implement paren evaluation (evaluates contents, returns result) in `internal/eval/evaluator.go`

#### Math Natives (Basic Arithmetic)

- [X] T049 [P] [US1] Implement native `+` (add two integers with overflow detection) in `internal/native/math.go`
- [X] T050 [P] [US1] Implement native `-` (subtract two integers with overflow detection) in `internal/native/math.go`
- [X] T051 [P] [US1] Implement native `*` (multiply two integers with overflow detection) in `internal/native/math.go`
- [X] T052 [P] [US1] Implement native `/` (divide two integers with overflow check) in `internal/native/math.go`
- [X] T053 [US1] Implement division by zero error handling (Math error 400) in `internal/native/math.go`
- [X] T054 [US1] Register math natives in native dispatcher in `internal/native/registry.go`

#### Parser (Traditional Precedence)

- [X] T055 [US1] Implement tokenizer (scan input into tokens: numbers, words, operators, blocks, parens) in `internal/parse/tokenize.go`
- [X] T056 [US1] Implement parser with operator precedence (7 levels per contracts/math.md) in `internal/parse/parse.go`
- [X] T057 [US1] Implement AST construction respecting precedence (*, / before +, -) in `internal/parse/parse.go`
- [X] T058 [US1] Implement parentheses override for precedence in `internal/parse/parse.go`

#### Error Context Capture

- [X] T059 [P] [US1] Implement Near context capture (expression window) in `internal/verror/context.go`
- [X] T060 [P] [US1] Implement Where context capture (call stack from frames) in `internal/verror/context.go`

#### Basic REPL

- [X] T061 [US1] Implement REPL struct with evaluator, history in `cmd/viro/repl.go`
- [X] T062 [US1] Implement Read phase (read line from stdin) in `cmd/viro/repl.go`
- [X] T063 [US1] Implement Eval phase (parse and evaluate) in `cmd/viro/repl.go`
- [X] T064 [US1] Implement Print phase (display result or error) in `cmd/viro/repl.go`
- [X] T065 [US1] Implement Loop (repeat until exit command) in `cmd/viro/repl.go`
- [X] T066 [US1] Implement prompt display (`>>` for ready, `...` for continuation) in `cmd/viro/repl.go`
- [X] T067 [US1] Implement none result suppression (FR-044) in `cmd/viro/repl.go`
- [X] T068 [US1] Implement welcome message with version in `cmd/viro/repl.go`
- [X] T069 [US1] Create main entry point in `cmd/viro/main.go`

**Checkpoint**: At this point, User Story 1 should be fully functional - users can evaluate literals, arithmetic with precedence, words, blocks/parens in REPL

---

## Phase 4: User Story 2 - Control Flow with Native Functions (Priority: P2)

**Goal**: Users can use control flow constructs (when, if, loop, while) for decisions and repetition

**Independent Test**: Enter `when true [print "yes"]`, `if 1 < 2 ["less"] ["greater"]`, `loop 3 [print "hi"]` and verify execution

### Contract Tests for User Story 2 (TDD - Write FIRST)

- [X] T070 [P] [US2] Contract test for `when` (single branch conditional) in `test/contract/control_test.go`
- [X] T071 [P] [US2] Contract test for `if` (both branches required) in `test/contract/control_test.go`
- [X] T072 [P] [US2] Contract test for `loop` (repeat N times) in `test/contract/control_test.go`
- [X] T073 [P] [US2] Contract test for `while` (condition-based loop) in `test/contract/control_test.go`
- [X] T074 [P] [US2] Contract test for truthy conversion (none/false→false, others→true) in `test/contract/control_test.go`
- [X] T075 [P] [US2] Contract test for comparison operators (<, >, <=, >=, =, <>) in `test/contract/math_test.go`

### Implementation for User Story 2

#### Comparison & Logic Natives

- [X] T076 [P] [US2] Implement native `<` (less than) in `internal/native/math.go`
- [X] T077 [P] [US2] Implement native `>` (greater than) in `internal/native/math.go`
- [X] T078 [P] [US2] Implement native `<=` (less than or equal) in `internal/native/math.go`
- [X] T079 [P] [US2] Implement native `>=` (greater than or equal) in `internal/native/math.go`
- [X] T080 [P] [US2] Implement native `=` (equality with deep comparison) in `internal/native/math.go`
- [X] T081 [P] [US2] Implement native `<>` (not equal) in `internal/native/math.go`
- [X] T082 [P] [US2] Implement native `and` (logical AND with truthy conversion) in `internal/native/math.go`
- [X] T083 [P] [US2] Implement native `or` (logical OR with truthy conversion) in `internal/native/math.go`
- [X] T084 [P] [US2] Implement native `not` (logical negation) in `internal/native/math.go`

#### Control Flow Natives

- [X] T085 [P] [US2] Implement native `when` (condition + single block) in `internal/native/control.go`
- [X] T086 [P] [US2] Implement native `if` (condition + true-block + false-block) in `internal/native/control.go`
- [X] T087 [P] [US2] Implement native `loop` (count + body block with validation) in `internal/native/control.go`
- [X] T088 [P] [US2] Implement native `while` (condition block + body block, re-evaluate condition) in `internal/native/control.go`
- [X] T089 [US2] Register control flow natives in native dispatcher in `internal/native/registry.go`

#### Truthy Conversion

- [X] T090 [US2] Implement truthy conversion helper (none/false→false, all others→true) in `internal/eval/eval.go`

**Checkpoint**: User Stories 1 AND 2 complete - users can now evaluate expressions AND use control flow for logic

---

## Phase 5: User Story 3 - Series Operations (Priority: P3)

**Goal**: Users can create and manipulate series (blocks, strings) using first, last, append, insert, length?

**Independent Test**: Create `data: [1 2 3]`, use `first data`, `append data 4`, `insert data 0`, verify results

### Contract Tests for User Story 3 (TDD - Write FIRST)

- [X] T091 [P] [US3] Contract test for `first` (return first element, error on empty) in `test/contract/series_test.go`
- [X] T092 [P] [US3] Contract test for `last` (return last element, error on empty) in `test/contract/series_test.go`
- [X] T093 [P] [US3] Contract test for `append` (add to end, in-place modification) in `test/contract/series_test.go`
- [X] T094 [P] [US3] Contract test for `insert` (add at beginning, shift elements) in `test/contract/series_test.go`
- [X] T095 [P] [US3] Contract test for `length?` (return element count) in `test/contract/series_test.go`
- [X] T096 [P] [US3] Contract test for string series operations (character sequences) in `test/contract/series_test.go`

### Implementation for User Story 3

#### Series Natives

- [X] T097 [P] [US3] Implement native `first` (return element at index 0, validate non-empty) in `internal/native/series.go`
- [X] T098 [P] [US3] Implement native `last` (return element at index length-1, validate non-empty) in `internal/native/series.go`
- [X] T099 [P] [US3] Implement native `append` (add value to end, modify in-place) in `internal/native/series.go`
- [X] T100 [P] [US3] Implement native `insert` (add value at position 0, shift right) in `internal/native/series.go`
- [X] T101 [P] [US3] Implement native `length?` (return element count) in `internal/native/series.go`
- [X] T102 [US3] Register series natives in native dispatcher in `internal/native/registry.go`

#### Series Type Validation

- [X] T103 [US3] Implement series type checking (Block or String) in series natives in `internal/native/series.go`
- [X] T104 [US3] Implement empty series error handling (Script error 300) in `internal/native/series.go`

**Checkpoint**: User Stories 1, 2, AND 3 complete - users can now work with collections and data structures

---

## Phase 6: User Story 4 - Function Definition and Calls (Priority: P4)

**Goal**: Users can define custom functions with arguments and refinements, call them with proper scoping

**Independent Test**: Define `square: fn [n] [n * n]`, call `square 5`, verify returns 25. Test local-by-default scoping and refinements.

### Contract Tests for User Story 4 (TDD - Write FIRST)

- [X] T105 [P] [US4] Contract test for `fn` (parameter extraction, body capture) in `test/contract/function_test.go`
- [X] T106 [P] [US4] Contract test for function calls (argument binding, frame creation) in `test/contract/function_test.go`
- [X] T107 [P] [US4] Contract test for local-by-default scoping (local variables don't affect global) in `test/contract/function_test.go`
- [X] T108 [P] [US4] Contract test for flag refinements (--verbose → true/false) in `test/contract/function_test.go`
- [X] T109 [P] [US4] Contract test for value refinements (--title [] → value/none) in `test/contract/function_test.go`
- [X] T110 [P] [US4] Contract test for refinement order independence in `test/contract/function_test.go`
- [X] T111 [P] [US4] Contract test for nested function calls and recursion in `test/contract/function_test.go`

### Implementation for User Story 4

#### Function Native

- [X] T112 [US4] Implement native `fn` (parse parameters, extract refinements, capture body) in `internal/native/function.go`
- [X] T113 [US4] Implement parameter validation (words only, unique names, refinement syntax) in `internal/native/function.go`
- [X] T114 [US4] Implement refinement parsing (--flag vs --option []) in `internal/native/function.go`
- [X] T115 [US4] Register function native in native dispatcher in `internal/native/registry.go`

#### Function Call Evaluation

- [X] T116 [US4] Implement function call dispatch (recognize Function type) in `internal/eval/eval.go`
- [X] T117 [US4] Implement argument collection (scan args, separate positional from refinements) in `internal/eval/eval.go`
- [X] T118 [US4] Implement refinement collection (flags→true, values→next arg) in `internal/eval/eval.go`
- [X] T119 [US4] Implement argument count validation in `internal/eval/eval.go`

#### Frame Management for Functions

- [X] T120 [US4] Implement function frame creation (allocate on stack with layout) in `internal/stack/stack.go`
- [X] T121 [US4] Implement parameter binding (positional args + refinements to frame) in `internal/eval/eval.go`
- [X] T122 [US4] Implement local-by-default word binding (all body words bound to local frame) in `internal/frame/binding.go`
- [X] T123 [US4] Implement body evaluation in function frame context in `internal/eval/eval.go`
- [X] T124 [US4] Implement return value handling and frame cleanup in `internal/eval/eval.go`

#### Closure Support

- [X] T125 [US4] Implement parent frame capture for closures in `internal/frame/frame.go`
- [X] T126 [US4] Implement lexical scope chain traversal in `internal/frame/binding.go`

**Checkpoint**: User Stories 1-4 complete - users can now define and call functions with full scoping and refinements

---

## Phase 7: User Story 5 - Error Handling and Recovery (Priority: P5)

**Goal**: Users encounter structured errors with context, REPL remains usable after errors

**Independent Test**: Trigger errors (`undefined-word`, `1 / 0`, `+ "string" 5`), verify error messages and REPL continues

### Contract Tests for User Story 5 (TDD - Write FIRST)

- [X] T127 [P] [US5] Contract test for undefined word errors (Script error 300) in `test/contract/errors_test.go`
- [X] T128 [P] [US5] Contract test for division by zero errors (Math error 400) in `test/contract/errors_test.go`
- [X] T129 [P] [US5] Contract test for type mismatch errors (Script error 300) in `test/contract/errors_test.go`
- [X] T130 [P] [US5] Contract test for syntax errors during parsing (Syntax error 200) in `test/contract/errors_test.go`
- [X] T131 [P] [US5] Contract test for error context (Near and Where included) in `test/contract/errors_test.go`
- [X] T132 [P] [US5] Contract test for REPL error recovery (continues after error) in `test/integration/repl_test.go`

### Implementation for User Story 5

#### Error Propagation

- [X] T133 [US5] Implement error propagation in evaluator (attach context, return error) in `internal/eval/eval.go`
- [X] T134 [US5] Implement error propagation through function calls in `internal/eval/eval.go`

#### REPL Error Handling

- [X] T135 [US5] Implement error display in REPL (format with category, message, Near, Where) in `cmd/viro/repl.go`
- [X] T136 [US5] Implement REPL error recovery (catch error, display, continue loop) in `cmd/viro/repl.go`
- [X] T137 [US5] Implement state preservation after errors (maintain global context) in `cmd/viro/repl.go`

#### Error Validation

- [X] T138 [P] [US5] Implement undefined word error generation in word evaluation in `internal/eval/eval.go`
- [X] T139 [P] [US5] Implement type mismatch error generation in natives in `internal/native/*.go`
- [X] T140 [P] [US5] Implement syntax error generation in parser in `internal/parse/parse.go`

**Checkpoint**: User Stories 1-5 complete - robust error handling with clear messages and REPL stability

---

## Phase 8: User Story 6 - REPL Interactive Features (Priority: P6)

**Goal**: REPL provides command history, multi-line input, and helpful feedback

**Independent Test**: Use up/down arrows for history, enter incomplete expression with continuation prompt, multi-line blocks

### Contract Tests for User Story 6 (TDD - Write FIRST)

- [X] T141 [P] [US6] Integration test for command history (up arrow recalls previous) in `test/integration/repl_test.go`
- [X] T142 [P] [US6] Integration test for multi-line input (incomplete block shows `...`) in `test/integration/repl_test.go`
- [X] T143 [P] [US6] Integration test for exit commands (quit, exit) in `test/integration/repl_test.go`
- [X] T144 [P] [US6] Integration test for Ctrl+C interrupt in `test/integration/repl_test.go`

### Implementation for User Story 6

#### Command History (chzyer/readline)

- [X] T145 [US6] Integrate `github.com/chzyer/readline` library in `internal/repl/repl.go`
- [X] T146 [US6] Implement command history storage (up/down arrow navigation) in `internal/repl/repl.go`
- [X] T147 [US6] Implement history persistence (save to ~/.viro_history) in `internal/repl/repl.go`

#### Multi-line Input

- [X] T148 [US6] Implement incomplete expression detection (unclosed blocks/parens) in `cmd/viro/repl.go`
- [X] T149 [US6] Implement continuation prompt (`...`) for multi-line input in `cmd/viro/repl.go`
- [X] T150 [US6] Implement multi-line buffer accumulation in `cmd/viro/repl.go`
- [X] T151 [US6] Implement completion detection (closing brackets matched) in `cmd/viro/repl.go`

#### Exit & Interrupt

- [X] T152 [US6] Implement exit command recognition (`quit`, `exit`) in `cmd/viro/repl.go`
- [X] T153 [US6] Implement graceful REPL termination in `cmd/viro/repl.go`
- [X] T154 [US6] Implement Ctrl+C interrupt handling (cancel evaluation, return to prompt) in `cmd/viro/repl.go`

#### Data Operations Natives

- [X] T155 [P] [US6] Implement native `set` (assign value to word) in `internal/native/data.go`
- [X] T156 [P] [US6] Implement native `get` (retrieve value from word) in `internal/native/data.go`
- [X] T157 [P] [US6] Implement native `type?` (return datatype of value) in `internal/native/data.go`
- [X] T158 [US6] Register data natives in native dispatcher in `internal/native/registry.go`

#### I/O Natives

- [X] T159 [P] [US6] Implement native `print` (output value, reduce blocks, join with spaces) in `internal/native/io.go`
- [X] T160 [P] [US6] Implement native `input` (read line from stdin) in `internal/native/io.go`
- [X] T161 [US6] Register I/O natives in native dispatcher in `internal/native/registry.go`

**Checkpoint**: All user stories 1-6 complete - Viro Core Interpreter fully functional with interactive REPL

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories, validation, and documentation

### Integration Testing

- [X] T162 [P] [POLISH] Integration test for User Story 1 scenarios (spec acceptance scenarios) in `test/integration/us1_test.go`
- [X] T163 [P] [POLISH] Integration test for User Story 2 scenarios in `test/integration/us2_test.go`
- [X] T164 [P] [POLISH] Integration test for User Story 3 scenarios in `test/integration/us3_test.go`
- [X] T165 [P] [POLISH] Integration test for User Story 4 scenarios in `test/integration/us4_test.go`
- [ ] T166 [P] [POLISH] Integration test for User Story 5 scenarios in `test/integration/us5_test.go`
- [ ] T167 [P] [POLISH] Integration test for User Story 6 scenarios in `test/integration/us6_test.go`

### Performance Benchmarking

- [ ] T168 [P] [POLISH] Benchmark stack operations (<100ns per research.md) in `internal/stack/stack_bench_test.go`
- [ ] T169 [P] [POLISH] Benchmark arithmetic operations (<1µs per research.md) in `internal/native/math_bench_test.go`
- [ ] T170 [P] [POLISH] Benchmark simple expression evaluation (<10ms per SC-005) in `test/integration/eval_bench_test.go`
- [ ] T171 [P] [POLISH] Benchmark complex expression evaluation (<100ms per SC-005) in `test/integration/eval_bench_test.go`

### Success Criteria Validation

- [ ] T172 [POLISH] Validate SC-001: 20+ expression types evaluate correctly
- [ ] T173 [POLISH] Validate SC-002: 1000+ evaluation cycles without leaks (memory profiling)
- [ ] T174 [POLISH] Validate SC-003: Error messages enable diagnosis in <2 minutes (user testing)
- [ ] T175 [POLISH] Validate SC-004: Recursive functions to depth 100+ without overflow
- [ ] T176 [POLISH] Validate SC-005: Performance baselines met (benchmarks)
- [ ] T177 [POLISH] Validate SC-006: 95%+ type errors caught during validation
- [ ] T178 [POLISH] Validate SC-007: Command history supports 100+ commands
- [ ] T179 [POLISH] Validate SC-008: Multi-line input handles 10+ nested levels
- [ ] T180 [POLISH] Validate SC-009: Stack expansion <1ms (transparent)
- [ ] T181 [POLISH] Validate SC-010: Ctrl+C interrupt returns to prompt <500ms

### Documentation & Build

- [ ] T182 [P] [POLISH] Update `docs/interpreter.md` with architecture overview
- [ ] T183 [P] [POLISH] Create build instructions in `specs/001-implement-the-core/quickstart.md`
- [ ] T184 [P] [POLISH] Create REPL usage examples with all features
- [ ] T185 [P] [POLISH] Document operator precedence table for users
- [ ] T186 [P] [POLISH] Document local-by-default scoping vs REBOL differences

### Code Quality

- [ ] T187 [POLISH] Run golangci-lint and fix issues across all packages
- [ ] T188 [POLISH] Run gofmt across all Go files
- [ ] T189 [POLISH] Add package-level documentation comments
- [ ] T190 [POLISH] Review and refactor for code duplication
- [ ] T191 [POLISH] Optimize hot paths identified by profiling

### Final Validation

- [ ] T192 [POLISH] Run all contract tests (verify 100% pass)
- [ ] T193 [POLISH] Run all integration tests (verify 100% pass)
- [ ] T194 [POLISH] Run quickstart.md validation (manual REPL testing)
- [ ] T195 [POLISH] Verify constitution compliance (all 7 principles)
- [ ] T196 [POLISH] Code review and architectural validation
- [ ] T197 [POLISH] Create release notes for v1.0.0

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - ⚠️ BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - User stories proceed sequentially by priority: P1 → P2 → P3 → P4 → P5 → P6
  - Or can proceed in parallel if team capacity allows (after Foundational)
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1) - Basic Expressions**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P2) - Control Flow**: Depends on US1 (evaluator and math natives) - Adds control flow on top
- **User Story 3 (P3) - Series Operations**: Independent of US2, depends on US1 (evaluator) - Can parallel with US2
- **User Story 4 (P4) - Functions**: Depends on US1 (evaluator) - Can parallel with US2/US3
- **User Story 5 (P5) - Error Handling**: Depends on US1-US4 (all error sources) - Integrates errors across system
- **User Story 6 (P6) - REPL Features**: Depends on US1 (basic REPL exists) - Can parallel with US2-US5

### Within Each User Story

**TDD Order (Constitutional Requirement)**:
1. Write contract tests FIRST (must FAIL initially)
2. Implement features to make tests PASS
3. Refactor while keeping tests GREEN

**Implementation Order Within Story**:
- Tests (contract + integration) → FIRST, must fail
- Core types/structures
- Helper functions
- Main implementation
- Integration with evaluator
- Error handling
- Story validation

### Parallel Opportunities

**Phase 1 (Setup)**: All tasks marked [P] can run in parallel

**Phase 2 (Foundational)**: 
- Value constructors (T008, T009) - parallel
- Error factories (T012, T013, T014) - parallel
- Stack/Frame operations - parallel across packages
- Series types (T022-T025) - parallel

**Phase 3+ (User Stories)**:
- Once Foundational complete, stories CAN proceed in parallel with careful coordination:
  - US1 (basic eval) → Must complete first (others depend on it)
  - US2 (control flow) + US3 (series) → Can parallel after US1
  - US4 (functions) → Can parallel with US2/US3 after US1
  - US5 (errors) → Should integrate after US1-US4 mostly done
  - US6 (REPL features) → Can parallel throughout

**Within Each User Story**:
- All contract tests marked [P] can run in parallel
- All natives in same category marked [P] can run in parallel
- Different packages/files marked [P] can run in parallel

---

## Parallel Execution Examples

### Phase 2 Foundational (Parallel Foundation)

```bash
# All these can run simultaneously (different files):
T006: "Define ValueType constants in internal/value/types.go"
T010: "Define ErrorCategory constants in internal/error/categories.go"
T015: "Define Stack struct in internal/stack/stack.go"
T019: "Define FrameType constants in internal/frame/types.go"
T022: "Define BlockValue struct in internal/value/block.go"
```

### Phase 3 User Story 1 (Parallel Tests)

```bash
# Write all these tests first, in parallel:
T031: "Contract test for literal evaluation in test/contract/eval_test.go"
T032: "Contract test for word evaluation in test/contract/eval_test.go"
T036: "Contract test for arithmetic natives in test/contract/math_test.go"
T037: "Contract test for operator precedence in test/contract/math_test.go"
```

### Phase 3 User Story 1 (Parallel Natives)

```bash
# After evaluator core, implement all math natives in parallel:
T049: "Implement native + in internal/native/math.go"
T050: "Implement native - in internal/native/math.go"
T051: "Implement native * in internal/native/math.go"
T052: "Implement native / in internal/native/math.go"
```

### Phase 4 User Story 2 (Parallel Natives)

```bash
# Implement all comparison operators in parallel:
T076: "Implement native < in internal/native/math.go"
T077: "Implement native > in internal/native/math.go"
T078: "Implement native <= in internal/native/math.go"
T079: "Implement native >= in internal/native/math.go"

# Implement all control flow natives in parallel:
T085: "Implement native when in internal/native/control.go"
T086: "Implement native if in internal/native/control.go"
T087: "Implement native loop in internal/native/control.go"
T088: "Implement native while in internal/native/control.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

**Minimal Viable Product Scope**:
1. Complete Phase 1: Setup (T001-T005)
2. Complete Phase 2: Foundational (T006-T030) ⚠️ CRITICAL - blocks everything
3. Complete Phase 3: User Story 1 (T031-T069)
4. **STOP and VALIDATE**: 
   - Test all US1 acceptance scenarios from spec.md
   - Verify: literals, arithmetic, words, blocks/parens, basic REPL
   - Performance check: simple expressions <10ms
5. Deploy/demo if ready

**MVP Delivers**:
- Working REPL with prompt and result display
- Literal evaluation (integers, strings, logic, none)
- Word binding and evaluation (set-word assignment)
- Arithmetic with traditional operator precedence (3 + 4 * 2 → 11)
- Block/paren distinction (deferred vs immediate evaluation)
- Basic error handling (undefined words, math errors)
- Foundation for all future features

**Why Stop at US1?**:
- Proves core architecture (evaluator, stack, frames, type dispatch)
- Validates TDD approach and constitution compliance
- Demonstrates operator precedence parser works correctly
- Establishes baseline for performance benchmarking
- Provides tangible demo-able artifact
- Lowest risk validation of technical decisions

### Incremental Delivery (Story by Story)

**Delivery Sequence**:

1. **Foundation** → Complete Setup + Foundational (T001-T030)
   - Validation: All foundational tests pass, types and structures ready
   
2. **MVP: User Story 1** → Add basic expression evaluation (T031-T069)
   - Validation: REPL works, can evaluate arithmetic and assignments
   - **DEMO POINT**: Show working interpreter with math
   
3. **User Story 2** → Add control flow (T070-T090)
   - Validation: Conditionals and loops work independently
   - **DEMO POINT**: Show simple programs with logic
   
4. **User Story 3** → Add series operations (T091-T104)
   - Validation: Can manipulate blocks and strings
   - **DEMO POINT**: Show data structure handling
   
5. **User Story 4** → Add functions (T105-T126)
   - Validation: User-defined functions with refinements work
   - **DEMO POINT**: Show custom functions and local scoping
   
6. **User Story 5** → Add error handling (T127-T140)
   - Validation: Errors display properly, REPL recovers
   - **DEMO POINT**: Show robust error messages
   
7. **User Story 6** → Add REPL features (T141-T161)
   - Validation: History, multi-line, interrupts work
   - **DEMO POINT**: Show polished REPL experience
   
8. **Polish** → Quality and documentation (T162-T197)
   - Validation: All success criteria met, ready for release

**Each increment**:
- Builds on previous (no breaking changes)
- Independently testable
- Adds clear user value
- Can be demoed/deployed

### Parallel Team Strategy

**With 3+ Developers After Foundation**:

**Team A (Core Developer)**: 
- Focus: User Story 1 (critical path)
- Tasks: T031-T069 (evaluator, parser, basic REPL)
- Timeline: 2-3 weeks

**Team B (Control Flow)**: 
- Starts after US1 evaluator core (T038-T042)
- Focus: User Story 2
- Tasks: T070-T090 (control flow and comparisons)
- Timeline: 1-2 weeks

**Team C (Series & Functions)**: 
- Starts after US1 evaluator core
- Focus: User Stories 3 & 4 (can be split)
- Tasks: T091-T126 (series operations, functions)
- Timeline: 2-3 weeks

**Merge Strategy**:
- US1 merges first (establishes base)
- US2, US3, US4 merge independently (minimal conflicts)
- US5 integrates errors across all (coordinate with all teams)
- US6 adds REPL polish (coordinate with Team A)

---

## Task Execution Notes

### TDD Discipline (Constitutional)

**Every implementation task MUST**:
1. Have tests written FIRST (contract tests before implementation)
2. Verify tests FAIL before implementing
3. Implement minimum code to make tests PASS
4. Refactor while keeping tests GREEN
5. Commit only when tests pass

**Contract Test Pattern**:
```go
// test/contract/math_test.go
func TestNativeAdd(t *testing.T) {
    tests := []struct {
        name     string
        args     []Value
        want     Value
        wantErr  bool
    }{
        {"positive", []Value{IntVal(3), IntVal(4)}, IntVal(7), false},
        {"negative", []Value{IntVal(-5), IntVal(10)}, IntVal(5), false},
        {"type error", []Value{StrVal("x"), IntVal(4)}, nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NativeAdd(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error status")
            }
            if !ValuesEqual(result, tt.want) {
                t.Errorf("got %v, want %v", result, tt.want)
            }
        })
    }
}
```

### File Path Precision

Every task includes exact file path:
- `internal/value/value.go` - Value types and constructors
- `internal/eval/eval.go` - Evaluator core (Do_Next, Do_Blk)
- `internal/native/math.go` - Math natives (+, -, *, /, comparisons)
- `internal/native/control.go` - Control flow natives (when, if, loop, while)
- `test/contract/math_test.go` - Math contract tests
- `cmd/viro/repl.go` - REPL implementation

### Commit Strategy

**After each task or logical group**:
```bash
git commit -m "[US1] T038: Implement Do_Next with type dispatch"
git commit -m "[US2] T085-T088: Implement control flow natives"
```

**After each checkpoint**:
```bash
git commit -m "[CHECKPOINT] User Story 1 complete - basic expression evaluation working"
```

### Validation at Checkpoints

**After each user story phase**:
1. Run all contract tests for that story: `go test ./test/contract/...`
2. Run integration tests: `go test ./test/integration/...`
3. Manual REPL testing: validate acceptance scenarios from spec.md
4. Performance check: run relevant benchmarks
5. Update checklist: mark story as complete

**Constitution Compliance Check**:
- [ ] TDD followed (tests before implementation)?
- [ ] Incremental layering respected (no skipping phases)?
- [ ] Type dispatch working correctly?
- [ ] Stack/frame using index-based access (no pointers)?
- [ ] Errors structured with categories and context?
- [ ] Behavior observable (REPL provides feedback)?
- [ ] YAGNI respected (no premature features)?

---

## Total Task Summary

**Total Tasks**: 197 tasks organized across 9 phases

**Task Breakdown by Phase**:
- Phase 1 (Setup): 5 tasks
- Phase 2 (Foundational): 25 tasks ⚠️ BLOCKING
- Phase 3 (US1 - Basic Expressions): 39 tasks 🎯 MVP
- Phase 4 (US2 - Control Flow): 21 tasks
- Phase 5 (US3 - Series): 14 tasks
- Phase 6 (US4 - Functions): 22 tasks
- Phase 7 (US5 - Error Handling): 14 tasks
- Phase 8 (US6 - REPL Features): 17 tasks
- Phase 9 (Polish): 40 tasks

**Parallel Opportunities**: 80+ tasks marked [P] can run in parallel within their phase

**Independent Test Criteria per Story**:
- US1: REPL evaluates literals, arithmetic, words, blocks/parens correctly
- US2: Control flow (when, if, loop, while) executes as specified
- US3: Series operations (first, last, append, insert, length?) work on blocks and strings
- US4: User-defined functions with refinements and local scoping work correctly
- US5: Errors display with context, REPL recovers and continues
- US6: Command history, multi-line input, interrupts function properly

**Suggested MVP Scope**: Phases 1-3 (Setup + Foundational + User Story 1) = 69 tasks

**Estimated Timeline** (single developer, TDD):
- MVP (Phases 1-3): 3-4 weeks
- Full feature (Phases 1-8): 8-10 weeks
- With polish (All phases): 10-12 weeks

**Estimated Timeline** (3 developers, parallel after foundation):
- Foundation (Phases 1-2): 1-2 weeks
- User Stories (Phases 3-8 parallel): 3-4 weeks
- Polish (Phase 9): 1-2 weeks
- **Total**: 5-8 weeks

---

## Success Validation

**Phase 1-2 Complete When**:
- ✅ All foundational types defined (Value, Error, Stack, Frame, Block, Function)
- ✅ Constructor functions working
- ✅ Error factory functions ready
- ✅ Stack/frame operations tested

**User Story 1 Complete When**:
- ✅ All T031-T037 contract tests pass
- ✅ REPL starts and displays prompt
- ✅ Can evaluate: `5` → `5`, `3 + 4` → `7`, `x: 10  x` → `10`
- ✅ Block vs paren: `[1 + 2]` → `[1 + 2]`, `(1 + 2)` → `3`
- ✅ Precedence: `3 + 4 * 2` → `11` (not 14)
- ✅ Errors: `undefined-word` shows error, REPL continues

**User Story 2 Complete When**:
- ✅ All T070-T075 contract tests pass
- ✅ Can use: `when true [42]` → `42`, `if false [1] [2]` → `2`
- ✅ Loops: `loop 3 [42]` executes 3 times
- ✅ Comparisons: `1 < 2` → `true`, `3 = 3` → `true`

**User Story 3 Complete When**:
- ✅ All T091-T096 contract tests pass
- ✅ Series ops: `data: [1 2 3]  first data` → `1`, `append data 4` → `[1 2 3 4]`
- ✅ Strings: `first "hello"` → `h`, `length? "hello"` → `5`

**User Story 4 Complete When**:
- ✅ All T105-T111 contract tests pass
- ✅ Functions: `square: fn [n] [n * n]  square 5` → `25`
- ✅ Local scoping: global variables unaffected by function locals
- ✅ Refinements: `fn [name --title []] [...]` with `--title "Dr."` works

**User Story 5 Complete When**:
- ✅ All T127-T132 contract tests pass
- ✅ Errors show category, message, Near, Where context
- ✅ REPL recovers: error → display → continue accepting input

**User Story 6 Complete When**:
- ✅ All T141-T144 integration tests pass
- ✅ History: up arrow recalls previous commands
- ✅ Multi-line: incomplete expression shows `...` prompt
- ✅ Exit: `quit` terminates REPL
- ✅ Interrupt: Ctrl+C cancels evaluation

**All Success Criteria Met When**:
- ✅ SC-001 through SC-010 validated (per Phase 9 tasks T172-T181)
- ✅ All contract tests pass (100%)
- ✅ All integration tests pass (100%)
- ✅ Performance benchmarks meet targets
- ✅ Quickstart.md validation succeeds
- ✅ Constitution compliance verified

---

**Ready to implement!** Start with Phase 1 (Setup), then Phase 2 (Foundation - CRITICAL), then User Story 1 for MVP. 🚀
