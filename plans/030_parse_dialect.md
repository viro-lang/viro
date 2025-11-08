# Implementation Plan: Parse Dialect

**Branch**: `030-parse-dialect` | **Date**: 2025-11-08 | **Spec**: `specs/030-parse-dialect/spec.md`

## Feature Summary

Viro needs a user-facing `parse` native that mirrors the Rebol dialect described in https://www.rebol.com/docs/core23/rebolcore-15.html. The feature must evaluate declarative rule blocks against both `string!` and `block!` inputs, support captures and mutations, and return boolean success/failure with descriptive errors on mismatches. The existing `parse` native is currently a stage-two parser helper; this initiative introduces a new dialect engine while preserving existing tooling by renaming or aliasing the legacy entry point.

## Research Findings

### Existing parser native

- `internal/native/parse.go` exposes a tokenizer-oriented helper. `cmd/viro/help.go` and `test/contract/parser_test.go` rely on its current parameters.
- Parser natives are wired through `internal/native/register_io.go`. Any renaming or aliasing must keep registrations, help output, and docs consistent.
- Plans `001_parser_refactor.md` and `027_two_stage_parser.md` show earlier investments in tokenizer/parse separation, so the new dialect must not regress that workflow.

### Series and evaluator support

- `internal/value/series.go` plus `internal/value/block.go` expose cursor semantics that work for strings and blocks; they already manage copying, slicing, and share detection.
- Evaluator helpers (`internal/eval/evaluator.go`, especially `DoBlock` and path traversal helpers) provide safe hooks for evaluating parens or words encountered inside rule blocks.
- Path/series handling (e.g., `value.PathSegment`) demonstrates how backtracking and mutations are typically orchestrated in the interpreter.

### Missing dialect primitives

- `value/types.go` has no `bitset!/charset` type, so Rebol-style charsets will require a new value plus constructor native.
- There is no reusable backtracking engine; we need a dedicated `internal/parse/dialect` package with cursors, stacks, and evaluator bridges.
- Capture/mutation helpers (`copy`, `set`, `insert`, `change`, `collect`) currently exist outside parse contexts; new glue code is required to reuse them safely during parsing.

### Documentation and guidelines

- AGENTS.md enforces TDD (contract tests first), no inline Go comments, and the viro-coder/viro-reviewer workflow for interpreter code.
- There is no parse dialect documentation in `docs/`; new guides and examples are necessary.
- `.specify/memory/constitution.md` is a placeholder, so constitution checks live in this plan/spec bundle (TDD + integration-test gates).

## Architecture Overview

1. **Native boundary** – Introduce a user-facing `parse` native defined in `internal/native/parse_dialect.go`. It accepts the input series, rules (block or string), and refinement flags `--case`, `--all`, `--any`, `--part`. The legacy helper ships as `parse-values` (or similar) for tokenizer consumers.

2. **Dialect engine** – Build `internal/parse/dialect` housing:
   - `SeriesCursor` abstraction over strings, blocks, and eventually binaries.
   - `MatchState` and `BacktrackFrame` stacks for recursion depth, repeat bounds, and marks.
   - `EvaluatorHook` so actions (`set`, `copy`, parens) can call into `core.Evaluator` under sandbox controls.

3. **Rule evaluation** – Rules remain ordinary Viro values. The engine dispatches on datatype words, literal values, quoted literals, and special words (`any`, `some`, `into`, etc.). Control words (`fail`, `reject`, `accept`, `break`, `if`, `while`, `not`) return explicit engine codes.

4. **Value support** – Add `bitset!/charset` to represent character sets. Block parsing uses datatype checks and literal word matching. Strings honor `--case` and `--all` and use charsets for classes.

5. **Compatibility story** – Provide a shim so existing scripts using the old helper continue to work. Document migration steps in release notes and CLI help.

## Implementation Roadmap

1. **Task 1 – Specifications & planning (this change)**  
   Publish `plans/030_parse_dialect.md` plus `specs/030-parse-dialect/` (spec, research, data model, quickstart, checklists, contracts) capturing requirements, constitution checks, and testing strategy.

2. **Task 2 – Preserve current parser helper**  
   Rename/alias the existing native to `parse-values`, update registrations/help/tests, and add release-note guidance.

3. **Task 3 – Dialect engine scaffold**  
   Implement `internal/parse/dialect` core types (cursor, matcher, evaluator bridge) with focused unit tests.

4. **Task 4 – Literal + alternation core**  
   Expose the new native, supporting literal sequences, `|`, and boolean return semantics, with initial `test/contract` suites for strings and blocks.

5. **Task 5 – Navigation & repetition**  
   Add `any`, `some`, `opt`, numeric repeat ranges, `skip`, `to`, `thru`, `--all`, and `--part` behavior; grow contract coverage around phone numbers, HTML tags, and whitespace handling.

6. **Task 6 – Capture & evaluation**  
   Implement `copy`, `set`, `word:` marks, `:word` jumps, and paren evaluation via the evaluator.

7. **Task 7 – Block semantics**  
   Support datatype tests, literal word matches, recursive rules, and `into` for nested block parsing.

8. **Task 8 – Charsets & refinements**  
   Add `bitset!/charset`, the `charset` native, and `--case`/`--any` handling for strings.

9. **Task 9 – Control flow**  
   Implement `fail`, `reject`, `accept`, `break`, `if`, `while`, `not`, tying them to `verror` diagnostics.

10. **Task 10 – Mutation actions & collection**  
    Add `insert`, `remove`, `change`, `collect`, `keep`, ensuring copy-on-write safety.

11. **Task 11 – Integration & documentation**  
    Update CLI help, docs, examples, release notes, and run `go test ./...` plus `make build`.

## Integration Points

- `internal/native/register_io.go` and `cmd/viro/help.go` for exposing the new native and aliasing the old helper; define refinement flags via `value.NewRefinementSpec` (e.g., `--case`, `--all`, `--part`).
- `docs/` (architecture knowledge, new parse guide) and `examples/` for canonical recipes.
- `test/contract/` suites for strings, blocks, control flow, and mutation; `test/integration/` for end-to-end DSLs.
- `internal/verror` for new parse failure IDs and diagnostics.

## Testing Strategy

- Follow TDD: create contract tests (e.g., `parse_string_basic_test.go`, `parse_block_semantics_test.go`, `parse_control_test.go`, `parse_mutation_test.go`) before implementing each capability.
- Add `internal/parse/dialect` unit tests for cursor math, backtracking, and charset evaluation.
- Provide integration scripts that exercise CLI usage, ensuring `parse` works in REPL, `do`, and `--check` modes.
- Keep fixtures small and focused to simplify diffs and reviews.

## Potential Challenges

- **Compatibility** – Renaming the old helper risks regressions; aliasing plus release notes must be part of Task 2.
- **Datatype expansion** – Adding `bitset!/charset` touches value tables, molding, equality, and serialization.
- **Backtracking cost** – Complex grammars can explode; instrumentation and optional depth limits may be required.
- **Mutation safety** – Actions must respect series sharing rules to avoid corrupting other references.

## Viro Guidelines Reference

- Tests first, per AGENTS.md.
- No inline Go comments; document intent in specs/docs.
- Use value constructors (`value.StringVal`, `value.BlockVal`) instead of raw structs.
- Route errors through `verror.NewScriptError` / `verror.NewInternalError` with specific IDs.

## Open Questions & Assumptions

1. Assume `parse-values` is an acceptable name for the legacy helper; confirm with maintainers before Task 2.
2. Assume bitsets warrant a dedicated value type rather than reusing `binary!`; revisit after prototype measurements.
3. Performance targets default to "parity with Rebol" unless future profiling dictates otherwise.

## Implementation Checklist

- [ ] Spec + research + contract scaffolding committed (`specs/030-parse-dialect/`).
- [ ] Legacy parser helper renamed/aliased without breaking tokenizer workflows.
- [ ] Dialect engine package merged with unit tests.
- [ ] Contract suites cover strings, blocks, navigation, captures, control flow, mutation, and charsets.
- [ ] Docs/examples/release notes updated and `go test ./...` + `make build` succeed.