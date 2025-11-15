# Implementation Plan: has? series membership

**Branch**: `[039-has-block-support]` | **Date**: 2025-11-15 | **Spec**: `specs/002-implement-deferred-features/contracts/reflection.md` (FR-022) + `specs/001-implement-the-core/data-model.md` (§9 Series)

## Summary

Extend the `has?` native beyond objects so it can answer "does this series contain this value?" for every implemented series type. The call signature stays `has? target value`, but the first argument may now be an object (current behavior) or any series (`block!`, `paren!`, `string!`, `binary!`, and future series such as `path!`, `lit-path!`, `get-path!`, `set-path!`, `list!`). The function must return `true` only when an element in the series equals the sought value via the existing `core.Value.Equals` semantics, scanning linearly without mutating the series. Edge cases—empty series, `none` values, values that appear multiple times, series whose current index is not the head—must be documented and exercised in tests. Documentation and contracts must state that the operation is an O(n) membership test.

## Technical Context

**Language/Version**: Go 1.25.1 (per `go.mod`)
**Primary Dependencies**: Go stdlib, `internal/core`, `internal/value` (Series + equality helpers)
**Storage**: N/A (in-memory interpreter state only)
**Testing**: `go test` with emphasis on `test/contract` suites (per AGENTS TDD mandate)
**Target Platform**: Cross-platform CLI interpreter (Linux/macOS/Windows) built from this repo
**Project Type**: Single Go project producing the `viro` CLI/interpreter
**Performance Goals**: Preserve O(n) scan, avoid allocations (no copying series contents) and keep membership checks cheap enough for REPL use
**Constraints**: TDD first, no inline Go comments, use constructor helpers (`value.NewLogicVal`, etc.), obey `viro-coder`/`viro-reviewer` workflow
**Scale/Scope**: Scoped to one native, shared helper, docs/spec updates, and contract tests (no evaluator or parser changes)

## Constitution Check

1. **TDD Gate** – Add/adjust `test/contract` suites (reflection + new series coverage) before touching `internal/native`. Ensure failures demonstrate the missing series support.
2. **Agent Workflow Gate** – All Go edits must be executed through the `viro-coder` agent and reviewed with `viro-reviewer` per `AGENTS.md`.
3. **Style Gate** – No inline comments in Go files; rely on clear naming and shared helpers. Use constructors from `value` and keep interfaces pointer-free per constitution principles (`specs/001-implement-the-core/data-model.md`). Re-check compliance after design in case helper extraction affects other natives.

## Project Structure

```
plans/
└── 039_has-block-support.md              # this plan

specs/
├── 001-implement-the-core/
│   └── data-model.md                     # Series interface + principles (§9)
└── 002-implement-deferred-features/
    └── contracts/
        └── reflection.md                 # Reflection/native contracts (add has?)

internal/
└── native/
    ├── reflection.go                     # has? implementation (object-only today)
    ├── series_helpers.go                 # assertSeries + ideal place for membership helper
    ├── series_block.go / series_string.go# reference implementations of find/select
    └── register_help.go                  # user-facing docs for has?

docs/
└── viro_architecture_knowledge.md        # interpreter reference (add edge cases + search semantics)

test/
└── contract/
    ├── reflection_test.go                # existing has? coverage (objects)
    └── series_has_test.go (new)          # focused series membership cases (or extend reflection_test)
```

**Structure Decision**: Single-project layout already in place. Changes remain confined to `internal/native` (logic/doc registration), `specs/002` (contract doc), `docs/` (edge-case notes), and `test/contract` (TDD). No new packages or binaries required.

## Complexity Tracking

_No additional complexity beyond existing structure._

## Feature Summary

- `has? target value` stays a two-argument native but must branch on the type of `target`: objects use the existing prototype-aware lookup while all series types perform a linear membership check using strict equality.
- Membership works for every type that implements `value.Series` today (block!, paren!, string!, binary!) and must be easy to extend to upcoming series values (path!* variants, list!) by reusing a centralized helper.
- `none` is handled like any other value (true only when a literal `none` element exists); empty series return `false`. Searching for complex values (objects, functions, words) should succeed whenever `Equals` matches.
- Error semantics align with spec FR-022: `has?` with unsupported first argument yields the existing type mismatch/action-not-implemented errors. `field`/value argument accepts “any!” when `target` is series, but keeps the current `word!/string!` requirement when `target` is object.
- Docs/specs must clarify runtime costs (linear scan) and highlight that series membership ignores the current index (matches `find` semantics) unless future requirements dictate otherwise.

## Research Findings

1. `internal/native/reflection.go:244-275` implements `has?` with strict checks: `object!` first argument, `word!/string!` second argument, prototype lookup via `Object.GetFieldWithProto`.
2. Contract coverage lives in `test/contract/reflection_test.go:530-689`; every case assumes object input, so series behavior lacks TDD coverage as required by AGENTS (“tests first in test/contract”).
3. Series abstraction resides in `internal/value/series.go` and is referenced via `assertSeries` in `internal/native/series_helpers.go:97-103`. `value.IsSeries` (`internal/value/types.go:82-89`) currently treats `block!`, `paren!`, `string!`, `binary!` as series.
4. Existing search patterns (`BlockFind` in `internal/native/series_block.go:9-36`, `StringFind` in `internal/native/series_string.go:11-43`, `BinaryFind` analog) scan slices linearly, ignore the series cursor, and rely on `core.Value.Equals`. Reusing this approach keeps semantics consistent and avoids regressions.
5. No spec entry documents `has?`, but `specs/002-implement-deferred-features/contracts/reflection.md` is the canonical reference for reflection helpers. It must pick up a new section describing both object and series contracts; `specs/001-implement-the-core/data-model.md` (§9) already states that series are mutable, index-based sequences whose operations should use the `Series` interface.
6. Help metadata in `internal/native/register_help.go:320-339` describes `has?` purely as an object helper. Updating this doc (and optionally `docs/viro_architecture_knowledge.md`) is required to communicate the new polymorphic behavior and note edge cases (empty series, `none`, linear scan cost).

## Architecture Overview

- **Polymorphic Dispatch**: Keep the existing `Has` native signature but branch early: if `target.GetType() == value.TypeObject`, execute the current field lookup path. Otherwise attempt to treat `target` as `value.Series` via `assertSeries`. Non-series, non-object targets raise `type mismatch`/`action-no-impl` errors consistent with other series natives.
- **Shared Series Membership Helper**: Add `seriesHasValue(series value.Series, sought core.Value) bool` (and optional `seriesHasValueIndex(series value.Series, sought core.Value) (bool, int)`) to `internal/native/series_helpers.go`. It should iterate from `0` to `Length()-1`, fetch each element through `ElementAt`, and compare with `Equals`. This helper centralizes logic for potential reuse by other natives (e.g., future `contains?`).
- **Equality Semantics**: Use `core.Value.Equals` for block/paren comparisons, and rely on existing `ElementAt` conversions (strings produce single-character `string!`, binary yields `integer!`). This ensures `has?` respects strict type equality and handles function/object values transparently because they implement `Equals`.
- **Edge Case Handling**: The helper returns `false` immediately for empty series (length zero) and never mutates the series or its index. Searching for `none` is just another equality check. Because `ElementAt` expects in-range indices, guard loops with `Length()` to avoid panics.
- **Documentation + Contracts**: Update `specs/002-implement-deferred-features/contracts/reflection.md` with a subsection for `has?`, describing both object and series behaviors, edge cases, and performance. Update `internal/native/register_help.go` and `docs/viro_architecture_knowledge.md` to mention the polymorphic nature and the O(n) scan. Include caution that strings compare substrings exactly (single-character units) and binary compares byte integers.

## Implementation Roadmap

1. **Deep-Dive Research (Phase 0)**
   - Confirm which types currently satisfy `value.Series` and whether additional types (path variants, list!) require infrastructure changes. Document findings in this plan and align them with `specs/001` §9.
   - Decide whether the membership helper should ignore the series index (matching `find`) or honor it. Default to `find` semantics but call out if spec clarifications are needed.

2. **TDD – Contract Tests (Phase 1, before Go edits)**
   - Extend `test/contract/reflection_test.go::TestHas` or add `test/contract/series_has_test.go` with table-driven cases covering:
     - `has?` on `block!`, `paren!`, `string!`, `binary!` returning true/false appropriately.
     - Mixed types: values being integers, words, objects, functions, `none`, nested block values.
     - Empty series returning false, `none` search hitting literal `none` entries, repeated values ensuring first match suffices.
     - Non-series, non-object first argument raising `ErrIDActionNoImpl` (via `assertSeries`).
     - Object path remains intact (a regression test to ensure prototypes still work).
   - Reflect doc requirement by adding tests showing series membership is independent of current index (e.g., `has? next [1 2 3] 1` should still be true).

3. **Helper Extraction (Phase 1 Design)**
   - Create `seriesHasValue(series value.Series, sought core.Value) bool` in `internal/native/series_helpers.go`. It should:
     - Handle nil/zero-length series quickly.
     - Iterate deterministically using `Length()` and `ElementAt(i)` without cloning.
     - Use `Equals` for comparison; stop early when a match occurs.
     - Optionally expose a variant returning the found boolean and index for future reuse.
   - Add unit tests for the helper if feasible (e.g., `internal/native/utils_test.go`) or rely on contract tests to validate behavior indirectly.

4. **`has?` Native Update (Phase 2)**
   - In `internal/native/reflection.go`, rewrite `Has` to:
     - Accept two args (same arity enforcement).
     - When `target` is `object!`, reuse current logic untouched.
     - Otherwise, call `assertSeries` on `target`. If it fails, return the existing type mismatch script error (message updated to mention "object! or series!").
     - Accept `value` parameter of any type when `target` is a series; no additional validation required since `Equals` handles type safety.
     - Return `logic!` result from helper (`true` when found). Ensure `None` is treated like any other value (helper covers this) and empty series short-circuit to `false`.
     - Ensure no allocations beyond the helper loop.

5. **Documentation + Contracts (Phase 2)**
   - Update `internal/native/register_help.go` entry for `has?` to describe the polymorphic behavior. Include examples for block, string, binary, and mention O(n) scan + empty-series behavior.
   - Add a `has?` section to `specs/002-implement-deferred-features/contracts/reflection.md` summarizing:
     - Signature: `has? target value`.
     - Object behavior (existing) and new series behavior referencing `specs/001` §9.
     - Edge cases (`none`, empty series) and performance note.
   - Insert a short note into `docs/viro_architecture_knowledge.md` (Reflection or Series sections) documenting the membership semantics, providing guidance on when to use `has?` vs `find`.

6. **Testing + Validation (Phase 3)**
   - Run targeted tests once implementation is complete: `go test ./test/contract -run TestHas` and any new test file.
   - Run a broader sweep (`go test ./...`) before committing to ensure no regressions.
   - Document linear-scan complexity in release notes or docs if necessary.

7. **Workflow Compliance (Phase 3)**
   - Use `viro-coder` agent for Go edits and `viro-reviewer` before finalizing.
   - Keep commits focused on this feature (tests + code + docs). Mention plan reference (#039) in commit message if repo convention expects it.

## Integration Points

- **`internal/native/reflection.go`** remains the single entry point for `has?`. Changes here must not affect other reflection natives.
- **`internal/native/series_helpers.go`** is the shared location for new iteration logic. Confirm no existing helper performs similar work to avoid duplication.
- **`internal/native/register_help.go`** updates surface new behavior to CLI help + `help has?` output.
- **`specs/002-implement-deferred-features/contracts/reflection.md`** must include the new contract text so future work referencing FR-022 aligns with behavior.
- **`docs/viro_architecture_knowledge.md`** should have a new note in the reflection or series section describing `has?` membership semantics and the linear scan trade-off.
- **`test/contract` suites** validate both old and new paths; make sure existing reflection tests still pass after updates.

## Testing Strategy

- **Contract Tests (primary gate)**: Expand `TestHas` (or new `TestHasSeries`) with table-driven cases covering:
  - Block membership positive/negative (including duplicates, nested values, `none`).
  - Paren membership (since paren reuses block storage but immediate evaluation semantics may differ).
  - String membership using single-character string comparisons.
  - Binary membership using integer byte comparisons.
  - Complex values (objects, functions) stored inside series to prove `Equals` usage.
  - Non-series first argument error path and retained object behavior.
- **Regression Tests**: Keep or add object-specific cases (prototype chain, variable field names) to ensure backwards compatibility.
- **Go Test Commands**: `go test ./test/contract -run 'TestHas'` for quick iterations, followed by `go test ./...` before completion.

## Potential Challenges

- **Path/List Coverage**: `value.TypePath` and related forms currently don’t implement `value.Series`. Decide whether to defer support until those types exist or introduce an adapter. Document the decision so future work knows how to extend coverage to lit-/get-/set-path/list once implemented.
- **String/Binary Element Types**: `ElementAt` for strings returns `string!` (single rune) while binary returns `integer!`. Tests must set expectations accordingly; mismatched types will always return false due to `Equals` semantics.
- **Performance vs. Allocation**: Avoid cloning the series or converting it to slices; iterate in place to maintain O(1) extra space and keep GC pressure low.
- **Index Semantics**: Clarify (in tests and docs) that `has?` checks the entire series regardless of the current index (matching `find`). If future requirements need cursor-aware behavior, document how to adapt the helper without breaking compatibility.
- **Error Messages**: When broadening input types, ensure script errors still specify precise expectations ("object! or series!" vs the previous "object!"). Update tests to pin the exact `verror` IDs.

## Viro Guidelines Reference

- Follow `AGENTS.md`: write/adjust `test/contract` cases before touching interpreter code, perform all Go edits via `viro-coder`, and run `viro-reviewer` prior to completion.
- Abide by `specs/001-implement-the-core/data-model.md` (§9) series principles (mutable sequences, index-based API) when writing the helper.
- Align behavior and documentation with `specs/002-implement-deferred-features/contracts/reflection.md` by adding the missing `has?` contract entry.
- Keep Go modules comment-free, use constructor helpers (`value.NewLogicVal`, `value.NewNoneVal`), and ensure equality checks rely on existing `core.Value.Equals` rather than custom logic.
- Validate using the prescribed commands (`go test ./test/contract`, `go test ./...`) before handing changes back to reviewers.
