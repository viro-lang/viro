# Research: Parse Dialect

## Primary Sources

- Rebol Core 2.3 — Chapter 15 "Parse" (https://www.rebol.com/docs/core23/rebolcore-15.html).
- Existing Viro parser work: `internal/native/parse.go`, `plans/027_two_stage_parser.md`, `test/contract/parser_test.go`.

## Interpreter Baseline

- `internal/native/parse.go` is a tokenizer helper, not a dialect evaluator.
- `internal/value/series.go` and `internal/value/block.go` provide series abstractions with copy-on-write semantics reused by the parse engine.
- Evaluator APIs (`core.Evaluator`, `internal/eval/evaluator.go`) already expose `DoBlock` and `DoParen` primitives we can call from parse actions.

## Rebol Dialect Expectations

- Rules are plain values: literal words/strings, datatype words, block groupings, special words (`any`, `some`, etc.), parenthesis expressions.
- Parse returns `true`/`false` and in Viro should expose refinement flags `--case`, `--any`, `--all`, `--part`, `--skip` (where applicable) plus `collect`, `keep`, and mutation actions.
- Character classes rely on `charset` and `bitset!` values; parse also recognizes `'word` for literal matches, `quote`/`thru`, and recursion via rule words.

## Gap Analysis

| Capability | Rebol | Current Viro | Gap |
|------------|-------|--------------|-----|
| User-facing `parse` | Yes | No (tokenizer helper only) | Need new native + engine |
| Charsets/bitsets | Yes | No `bitset!` type | Add type + native |
| Backtracking engine | Built-in | None | Implement `internal/parse/dialect` |
| Capture/mutation in parse | Yes | Not supported | Integrate with evaluator + series APIs |
| Diagnostics (`near`, `where`) | Yes | Parser-level only | Extend `verror` data |

## Tooling & Docs

- Documentation gaps: no parse reference, no Quickstart, no examples. Need `docs/parse-dialect.md` plus entries in README/examples.
- Testing gaps: contract suites must be added under `test/contract/` before implementation to satisfy TDD policy.

## Risks & Unknowns

1. Bitset storage size may inflate memory footprint; must evaluate compression or reuse existing binary representation.
2. Unlimited recursion can DOS the interpreter; consider configurable depth limits or instrumentation.
3. Mutation semantics on shared series need precise copy-on-write rules to avoid surprising behavior.
4. Interaction between `--case` and Unicode normalization has not been defined; assume byte-wise comparison until clarified.

## Next Steps

- Finalize naming for the legacy helper (`parse-values`) with maintainers.
- Draft contract files enumerated in `specs/030-parse-dialect/contracts/`.
- Prototype cursor/backtracking structures in `internal/parse/dialect` before wiring the native.