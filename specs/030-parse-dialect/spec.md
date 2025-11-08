# Feature Specification: Parse Dialect

**Feature Branch**: `030-parse-dialect`  
**Created**: 2025-11-08  
**Status**: Draft  
**Input**: User description: "Chciałbym dodać do Viro funkcję `parse`, która funkcjonalnością będzie odpowiadać tej z Rebola."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Validate Structured Text (Priority: P1)

Authors can validate and extract data from raw strings (log entries, HTML snippets, CSV fragments) using declarative rules instead of manual loops.

**Why this priority**: Without a string-focused `parse`, users cannot safely process external inputs or enforce schemas.

**Independent Test**: Run `parse email [copy local to "@" copy domain end]` and confirm success/failure across valid and invalid samples without mutating the input.

**Acceptance Scenarios**:

1. **Given** a string `"<title>Viro</title>"`, **When** a script evaluates `parse data [thru "<title>" copy text to "</title>" thru "</title>"]`, **Then** the call returns `true` and `text` captures `"Viro"`.
2. **Given** a malformed HTML snippet missing the closing tag, **When** the same rule runs, **Then** `parse` returns `false` and raises Syntax error (200) describing the failing rule with `near` and `where` info when tracing is enabled (e.g., wrap with `trace --on` ... `trace --off`).

---

### User Story 2 - Compose DSLs over Blocks (Priority: P2)

DSL authors can run rule blocks against nested block data (configuration ASTs, tokenized scripts) to validate and transform structures.

**Why this priority**: Many Viro workflows operate on block series; parity with Rebol enables existing dialect techniques.

**Independent Test**: Evaluate `parse script [some [set word word! set value integer! | into [word! integer!]]]` over nested blocks to ensure recursion and datatype guards behave correctly.

**Acceptance Scenarios**:

1. **Given** a block `[print "hi" repeat 3 [foo]]`, **When** a rule checks `[some [word! | into [word! integer!]]]`, **Then** `parse` succeeds and recursion respects depth limits.
2. **Given** a block with an unexpected datatype, **When** the same rule runs, **Then** `parse` returns `false` and the error message points to the offending value.

---

### User Story 3 - Capture and Mutate During Parsing (Priority: P3)

Power users can capture substrings, bind them to words, and mutate the source (insert/remove/change) while parsing to implement templating and reformatting pipelines.

**Why this priority**: Captures/mutations unlock DSL compilers and data cleanup flows that mirror Rebol behavior.

**Independent Test**: Run `parse data [collect [some [keep copy token to "," skip]]]` to build token lists, then `parse data [some ["foo" change "foo" "bar"]]` to mutate series.

**Acceptance Scenarios**:

1. **Given** `data: "a=1&b=2"`, **When** rules use `collect/keep` to gather key/value pairs, **Then** `parse` returns `["a" "1" "b" "2"]` without clobbering the source unless `change` is invoked.
2. **Given** shared series references, **When** `parse` executes `change` or `remove`, **Then** copy-on-write semantics ensure other references are unaffected.

### Edge Cases

- Empty rule blocks must return `true` without advancing the cursor; empty input must respect `any`/`some` semantics.
- `some []` or infinitely recursive rules must raise Script error (stack overflow) rather than hang; expose depth guard configuration if needed.
- `--case` must distinguish Unicode code points; combining `--case` with `charset` should not downcase input implicitly.
- Words bound inside rules must not leak their temporary values into caller frames unless explicitly `set`/`copy`ed.
- Paren evaluation may produce side effects; failures inside parens should bubble up as Script errors without corrupting parse state.
- Mutation actions must honor read-only series (e.g., string literals embedded in binaries) and raise Access error (500) when mutation is disallowed.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Provide a user-facing `parse` native that returns `true`/`false` and raises Syntax error (200) with `near`/`where` context when rules fail.
- **FR-002**: Accept `string!`, `block!`, and `binary!` (future) series as the first argument, and accept either a rule block or rule string. Rule strings are tokenized into blocks before evaluation, mirroring Rebol behavior.
- **FR-003**: Preserve the existing tokenizer helper by exposing it as `parse-values` (or equivalent) so downstream tools remain stable.
- **FR-004**: Support literal matches (strings, words, numbers), datatype tests (`integer!`, `time!`, etc.), quoted literals (`'word`, `"string"`), rule invocation via words, and rule grouping with brackets.
- **FR-005**: Implement repetition and alternation constructs: `any`, `some`, `opt`, `none`, integer repeat counts (`3 "a"`), range repeats (`3 5 word!`), and `|` alternation.
- **FR-006**: Implement navigation words: `skip`, `back`, `to`, `thru`, and refinement flags `--part`, `--all`, `--any`, `--case` affecting whitespace and case sensitivity.
- **FR-007**: Support capture/evaluation actions: `copy`, `set`, `word:` marks, `:word` jumps, and parenthesis evaluation by delegating to `core.Evaluator` with sandbox safeguards.
- **FR-008**: Support block semantics: `into` for nested parsing, recursive rule references, datatype predicates within block inputs, and literal word comparisons.
- **FR-009**: Introduce a `bitset!/charset` value plus a `charset` native that builds them from rule blocks, enabling `parse` to test character classes with `--case` awareness.
- **FR-010**: Implement control words `fail`, `reject`, `accept`, `break`, `if`, `while`, `not`, emitting deterministic engine return codes and mapping to `verror` categories where appropriate.
- **FR-011**: Implement mutation actions within parse rules: `insert`, `remove`, `change`, `collect`, and `keep`, all honoring copy-on-write guarantees and raising Access error (500) for immutable series.
- **FR-012**: Document the dialect via CLI help (`help parse`), `docs/parse-dialect.md`, examples, and release notes describing the legacy helper rename and migration guidance.

### Key Entities

- **ParseCall**: Captures arguments (input series, rule block, refinements) plus execution context (caller frame, evaluator reference).
- **SeriesCursor**: Wraps a `string!` or `block!`, tracking index, end, and snapshot/restore operations for backtracking.
- **RuleNode**: Intermediate representation describing each rule element (literal, datatype test, action, control word).
- **CharsetValue (`bitset!`)**: Represents sets of code points used by `charset`, `to`, `thru`, and class tests.
- **CaptureContext**: Stores word bindings, copy targets, and collected values that survive past the parse call.

## Success Criteria *(mandatory)*

- **SC-001**: The contract suites described in `specs/030-parse-dialect/contracts/` all pass, covering at least 50 representative grammars (strings, blocks, control, mutation).
- **SC-002**: Parsing a 100 KB log line with alternation and `some` completes within 150 ms on reference hardware (Apple M2) while keeping allocations under 5x input size.
- **SC-003**: Block-oriented DSLs with nested recursion (depth 32) complete without stack overflow and honor word bindings captured during parsing.
- **SC-004**: Documentation includes Quickstart, reference docs, and a CLI `help parse` entry; user feedback via examples demonstrates parity with Rebol semantics.
- **SC-005**: Regression suite ensures the legacy tokenizer helper remains accessible as `parse-values`, and prior parser tests pass unchanged.