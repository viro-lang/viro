# Feature Summary
Address reviewer findings around evaluated path segments by (1) rejecting leading eval segments in the grammar, (2) ensuring `Mold/Form` never produce unparsable `.(expr)` prefixes, (3) tightening evaluator semantics (type safety, caching, and clearer errors) for both read and assignment paths, and (4) documenting + testing the supported result types for eval segments across `path!`, `get-path!`, and `set-path!` usages.

# Research Findings
- `internal/parse/semantic_parser.go` accepts `PathSegmentEval` anywhere because `parsePath` only rejects numeric leading segments. `parsePathSegment` simply wraps paren contents in a `BlockValue`, so strings like `.(foo).bar` and `(foo).bar` round-trip despite not being officially supported.
- `internal/tokenize/tokenizer.go` treats `.(\n` as part of the literal regardless of position; there are no tokenizer tests for literals beginning with `.(`, so ambiguity around leading eval segments can slip through unnoticed.
- `internal/value/path.go` duplicates nearly identical `String`/`Mold`/`Form` logic for path/get-path/set-path, embeds inline comments (against `AGENTS.md`), and prepends `"."` when the first segment is eval—producing `.(expr)` even though such syntax should not be emitted.
- `internal/eval/evaluator.go` limits eval segment results to word/string/int via `materializeSegment`, re-resolves the first segment in both `traversePath` and `resolvePathBase`, re-evaluates the final segment during assignments, and uses unchecked type assertions (potential panics) across traversal helpers.
- Tests: `internal/parse/path_test.go` lacks negative cases for eval bases; `test/contract/objects_test.go` covers eval segments only for normal paths, not get-path/set-path or error cases (assignment failures, invalid result types, caching). There are no value-formatting tests ensuring molded paths stay parseable.
- Documentation (`specs/002-implement-deferred-features/data-model.md`, `docs/viro_architecture_knowledge.md`) describes eval segments but never mentions the unsupported leading-base scenario or the permitted result types.

# Architecture Overview
Tokenizer → parser builds `value.PathExpression` (segments typed as word/index/refinement/eval). Evaluator walks those segments, materialising eval segments on demand and traversing objects/series before optional assignment. `Mold`/`Form` convert path values back into source syntax, so any mismatch between parser and formatter breaks round-trip guarantees. Enforcing leading-segment rules therefore requires coordinated updates to tokenizer, parser, evaluator, and value-layer formatting, plus contract tests to lock behaviour.

# Implementation Roadmap
1. **Front-load test coverage (TDD requirement)**
   - Parser unit tests (`internal/parse/path_test.go`, `internal/tokenize/tokenizer_test.go`): add cases asserting that `.(foo).bar`, `(foo).bar`, `:.(foo).bar`, and `(foo):` variants raise a syntax error (`ErrIDInvalidPath` or new ID). Include tokenizer-focused tests ensuring literals beginning with `.(` stop at the right place and keep column info for error reporting.
   - Contract tests (`test/contract/path_eval_test.go` or extend `objects_test.go` / new dedicated file):
     * Positive: `obj.(field)` and `:obj.(field)` share coverage, including nested eval segments and multi-step indices.
     * Negative: get-path/set-path scenarios where eval segment returns unsupported types, non-existent fields, out-of-range indices, or attempts assignment through `none`.
   - Formatting tests (new `internal/value/path_test.go`): verify `Mold()` results for paths/get-paths/set-paths never start with `.(`, remain parseable by `parse.Parse`, and treat eval segments consistently.
   - Assignment error coverage: Add contract tests capturing script errors when eval segments appear in the last position for set-path, when they yield decimals/binaries, and when base resolves to numeric literal.

2. **Tokenizer + parser enforcement for leading eval segments**
   - Update `internal/tokenize/tokenizer.go` to track whether the literal started with `.(`; expose that state so parser can emit a targeted error (or prevent `.(` from beginning a literal unless preceded by a non-dot rune). Keep decimal handling intact by extending existing `depth` tracking tests.
   - Modify `parsePath` in `internal/parse/semantic_parser.go` to reject `segments[0].Type == PathSegmentEval` (and return syntax error with explanatory context). Ensure similar logic applies to get-path and set-path (since they reuse `parsePath`).
   - If tokenizer changes alone suffice, document decision; otherwise add a new syntax error ID (e.g., `ErrIDPathEvalBase`) in `internal/verror/categories.go` and use it for clearer diagnostics.
   - Update parser tests to expect the new error ID/text; assert location metadata when possible.

3. **Unify path formatting/molding and block leading `.(` outputs**
   - Introduce a helper (e.g., `renderPathSegments(opts)`) inside `internal/value/path.go` that both `String()` and `Mold()` variants can call for path/get-path/set-path, parameterised with prefix/suffix (colon, trailing colon) to remove duplicated logic.
   - Within that helper, detect `i == 0 && seg.Type == PathSegmentEval` and format it without a leading dot (e.g., wrap the expression as `(expr)` or emit a tagged placeholder like `#[invalid-eval-base ...]`). Document chosen behaviour and cover it with the new formatting tests.
   - Remove inline comments from the touched methods per `AGENTS.md`; rely on helper/variable naming to convey intent.
   - Ensure `BlockValue` mold caching is reused instead of calling both `Form()` and `MoldElements()` for the same segment when possible.

4. **Evaluator safety + result-type handling for eval segments**
   - Add typed accessors on `value.PathSegment` (e.g., `func (seg PathSegment) Word() (string, bool)`) or local helpers in `internal/eval/evaluator.go` to eliminate unchecked assertions throughout traversal and assignment. Convert failures into internal/script errors instead of panics.
   - Refactor `materializeSegment` to either (a) continue limiting results to word/string/int with a single exit point and explicit error text, or (b) optionally expand support for block/path results by flattening them into multiple segments; document the decision. If keeping the restriction, add tests confirming the script error message for unsupported types and mention it in docs.
   - Remove redundant first-segment handling by letting `traversePath` fully materialize `segments[0]` and passing the resolved base directly to a slimmed-down `resolvePathBase` (or inlining the logic). Ensure set-path traversal caches materialized segments so `assignToPathTarget` no longer re-evaluates eval segments.
   - Improve assignment error pathways: detect attempts to assign through eval segments that resolve to non-word/index early, and surface `ErrIDImmutableTarget` / `ErrIDPathTypeMismatch` with the molded path string for context.

5. **Documentation + cleanup**
   - Update `specs/002-implement-deferred-features/data-model.md` and `docs/viro_architecture_knowledge.md` path sections to state: leading eval segments are invalid, eval segments evaluate once per traversal, supported result types are `word!`, `string!`, and `integer!` (or whichever the team decides), and set-path/get-path share the same semantics.
   - Add a short note to `docs/viro_core_knowledge_rag.md` (or appropriate FAQ entry) describing why `.(expr)` paths are rejected and how to rewrite such expressions using explicit base values.
   - Remove any remaining inline comments you touch in Go files (value/path.go, evaluator.go, tokenizer.go) to comply with `AGENTS.md`.

6. **Optional stretch: caching/materialization reuse**
   - If time permits, extend `pathTraversal` to store `[]core.Value` for already materialized eval segments (e.g., map from segment index to cached `PathSegment`) so repeated traversals (especially set-path follow-up assignments) avoid duplicate evaluation. Guard this behind clear unit tests or benchmarks (`internal/eval/lookup_bench_test.go`).
   - Consider adding a small benchmark or instrumentation to ensure the caching does not regress existing parser performance (link to `docs/parser-performance.md`).

7. **Validation**
   - Run focused tests first: `go test ./internal/tokenize ./internal/parse ./internal/value -run Path`, `go test ./test/contract -run PathEval`, and any new benchmark packages touched.
   - Finish with `go test ./...` (or `make test` if CI expects it) to ensure no regressions elsewhere.

# Integration Points
- `internal/tokenize` ↔ `internal/parse`: literal/token boundaries must remain aligned after rejecting leading eval segments.
- `internal/value/path.go` ↔ `test/contract/parser_test.go`: molded output needs to stay parseable, so tests should round-trip using parser APIs.
- `internal/eval/evaluator.go` ↔ `test/contract/objects_test.go` & new path tests: traversal/assignment logic changes must maintain existing object/index semantics.
- Documentation updates must stay consistent with specs and user-facing guides in `docs/` to prevent conflicting instructions.

# Testing Strategy
- Unit tests: `go test ./internal/tokenize -run Eval`, `go test ./internal/parse -run Path`, `go test ./internal/value -run Path`.
- Contract tests: targeted `go test ./test/contract -run PathEval` (new suite) and existing `objects_test.go` cases.
- Full regression: `go test ./...` (or `make test`) before opening PR.
- Optional: add a parser round-trip test verifying `value.PathExpression.Mold()` parses back into an equivalent path for all supported segment types.

# Potential Challenges
- Adjusting tokenizer heuristics without breaking decimal literals or existing path/refinement parsing.
- Deciding whether to expand eval segment result support now vs. deferring; expanding requires rethinking how `PathSegment` arrays are mutated mid-traversal.
- Keeping molded output stable for non-eval paths while refactoring shared helpers.
- Ensuring removal of inline comments complies with the style guide without reducing clarity; helper naming must carry the intent instead.

# Viro Guidelines Reference
- `AGENTS.md`: mandates no inline comments in Go code, tests-before-code (TDD), and using viro-coder/viro-reviewer for implementation.
- `specs/002-implement-deferred-features/data-model.md`: authoritative description of path semantics that must be updated to mention eval restrictions.
- `docs/viro_architecture_knowledge.md`: reference section on path evaluation to keep user-facing behaviour accurate.
