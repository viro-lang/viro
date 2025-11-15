# 036_empty_tail_alignment

## Feature Summary
- Align the `empty?` native with `tail?` so both report truthiness based on the current series index, not only structural length.
- Applies to every value registered as a series action (blocks, strings, binaries, paths, etc.) so that slices or moved cursors behave consistently.
- Preserve existing error behaviour for non-series inputs and keep backward-compatible semantics elsewhere.

## Research Findings
- `internal/native/series_helpers.go:109-131` implements `seriesEmpty`, `seriesHeadQ`, and `seriesTailQ`. `seriesEmpty` currently returns `seriesVal.Length() == 0`, ignoring the cursor, while `seriesTailQ` checks `GetIndex() == Length()`.
- `internal/native/register_series.go` binds `empty?`, `head?`, and `tail?` for `TypeBlock`, `TypeString`, `TypeBinary`, plus other series-like types lower in the file, so any change in `seriesEmpty` automatically covers all these value types.
- Contract coverage for `empty?`/`tail?` lives in `test/contract/series_test.go:2422-2528`, but `empty?` tests only cover raw head states; none verify cursor movement (`next`, `skip`, `back`). `tail?` already verifies tail/head positions, so we can mirror those expectations.
- `internal/value/series.go` defines the `Series` interface and exposes both `Length()` and `GetIndex()` so the helper can rely on existing API without touching concrete implementations.
- `specs/001-implement-the-core/data-model.md` (Series section) clarifies that series expose both `Length` and `Index`. Even though the spec text emphasises `0 ≤ index < length`, the interpreter already allows `index == length` for tails (per `seriesTailQ`), which is the behaviour we need to mirror.
- `AGENTS.md` mandates writing contract tests before interpreter changes and requires using the `viro-coder` / `viro-reviewer` agents for implementation/review.

## Architecture Overview
- Introduce a shared helper in `internal/native/series_helpers.go` that encapsulates the cursor-vs-length comparison (e.g., `isSeriesAtOrBeyondTail(series value.Series) bool`).
- Reuse this helper inside both `seriesTailQ` and `seriesEmpty` to guarantee identical semantics and avoid divergence in the future.
- No evaluator/runtime changes should be necessary because the series interface already exposes the needed state.

## Implementation Roadmap
1. **Extend Contract Coverage** (`test/contract/series_test.go`):
   - Augment `TestSeries_QueryFunctions` with additional `empty?` cases that cover: head vs tail for blocks (`empty? [1 2 3]` remains false, `empty? tail [1 2 3]` → true), cursor shifts via `next`, `back`, and `skip` (`empty? skip [1 2 3] 3` → true, `empty? back tail [1 2 3]` → false).
   - Mirror the same scenarios for another series type (strings are the simplest) to prove polymorphic correctness (`empty? tail "hello"`, `empty? next "a"`, etc.).
   - Ensure at least one test covers a subseries/head vs tail transition to confirm slices respect the cursor (e.g., `empty? next tail [1 2 3]` or `empty? head tail "a"`).
   - Keep the existing non-series error test unchanged.
2. **Add Tail-State Helper** (`internal/native/series_helpers.go`):
   - After `assertSeries` (or near `seriesTailQ`), add a small unexported helper (no comments per style) that returns whether `series.GetIndex() >= series.Length()` so it remains safe even if future operations allow indexes beyond the tail.
3. **Update `seriesEmpty` Implementation** (`internal/native/series_helpers.go`):
   - Replace the `Length()==0` check with the helper so the logic uses the cursor. An empty series at head still returns true because both index and length are zero.
4. **Update `seriesTailQ` to Use Helper** (`internal/native/series_helpers.go`):
   - Have `seriesTailQ` delegate to the same helper to prevent future drift and to satisfy the “same boolean result” requirement literally.
5. **Housekeeping**:
   - Confirm no other files reference `seriesEmpty` directly; registration already fans it out, so no additional wiring is necessary.
   - Ensure imports remain sorted/alphabetised and no prohibited comments are introduced.
6. **Quality Gates**:
   - Run targeted tests: `go test ./test/contract -run TestSeries_QueryFunctions`.
   - Run broader smoke if needed (`go test ./...`) to ensure no regressions for other natives relying on `empty?`.

## Integration Points
- `internal/native/register_series.go` automatically applies the new semantics to every registered series type (blocks, strings, binaries, paths, parens, etc.), so no per-type code paths require edits.
- `test/contract/series_test.go` is the canonical place for contract expectations affecting evaluator behaviour; make sure new scenarios live alongside the current query tests for discoverability.
- The helper touches only `Series` interface consumers, so no change is required in `internal/value/*` implementations, but any behaviour change immediately affects evaluators, REPL, and CLI users through shared natives.

## Testing Strategy
- Expand `TestSeries_QueryFunctions` with:
  - Block cases: head (false), `tail` (true), `skip ... 3` (true), `back tail ...` (false) to prove cursor sensitivity.
  - String or binary cases covering `next`, `skip`, and tail creation, ensuring at least two series families are exercised.
  - (Optional) Additional guard verifying slices/subseries derived via `next/skip` behave identically when passed into `empty?` multiple times.
- Execute `go test ./test/contract -run TestSeries_QueryFunctions` first, then `go test ./...` if runtime permits, per `AGENTS.md` preference for automated validation.

## Potential Challenges
- Some specs still describe `index < length`; relying on `>=` avoids panics if other natives ever push the cursor beyond tail. Document reasoning in code review notes since inline comments are disallowed.
- Existing user code might rely on the current `empty?` semantics (length-only). Contract updates will reveal expectation changes; highlight this in release notes later if necessary.
- Ensure no helper introduces allocation or interface assertion overhead in tight loops; keep it as a simple inline function so the compiler can optimise it away.

## Viro Guidelines Reference
- `AGENTS.md`: mandates TDD (tests in `test/contract` before interpreter changes) and use of `viro-coder`/`viro-reviewer` agents for edits/reviews.
- `specs/001-implement-the-core/data-model.md` (Series section) for canonical description of `Series` capabilities and cursor semantics.
- `docs/viro_architecture_knowledge.md` (Series and native patterns) to ensure helper placement/routing follows established interpreter structure.
