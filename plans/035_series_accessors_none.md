# Plan 035: Soft-Fail Series/Object Accessors

## Feature Summary
- Ensure every read-only accessor over series-like values (blocks, strings, binaries) and object path lookups yields `none` when the requested element or field does not exist, rather than propagating script errors.
- Maintain existing error behaviour for true misuse (type mismatches, invalid refinements, assignments to immutable targets) so only "missing data" scenarios are softened.
- Update contract tests and documentation to describe and verify the new semantics, covering ordinal natives (`first` through `tenth`), positional lookups (`at`, path indexes), and word-based object accessors (dot and slash notation, `path`/`get-path`).

## Research Findings
- `internal/native/series_polymorphic.go` implements `seriesFirst`, `seriesLast`, and ordinal helpers via `seriesOrdinalAccess`; they currently translate empty/out-of-range conditions into `ErrIDEmptySeries` or `ErrIDOutOfBounds`.
- `internal/native/series_helpers.go` hosts `seriesAt` (used by `block/string/binary` `at`), `seriesPick`, and `seriesBack`; `seriesAt` raises `ErrIDOutOfBounds` whenever `absoluteIndex` is outside available elements from head positions.
- Path evaluation is centralized in `internal/eval/evaluator.go` (`traversePath`, `traverseWordSegment`, `traverseIndexSegment`, `checkIndexBounds`). Missing fields or out-of-range indexes surface as `ErrIDNoSuchField` / `ErrIDOutOfBounds`, causing `path`/`get-path` to error.
- Contract coverage expecting errors resides in:
  - `test/contract/series_test.go` (`TestSeries_First`, `TestSeries_Last`, `TestSeries_At`).
  - `test/contract/series_new_functions_test.go` (per-ordinal suites for `second` through `tenth`).
  - `test/contract/path_eval_test.go` and `test/contract/objects_test.go` (multiple cases asserting `no-such-field` / `out-of-bounds` errors for reads).
- Specs still document the old behaviour (`specs/002-implement-deferred-features/contracts/objects.md` states missing fields raise errors); `register_series.go` native docs omit the "returns none when missing" detail for ordinal/`at` actions.

## Architecture Overview
- Introduce a consistent "safe accessor" policy in the series helpers: instead of letting `seriesFirst`/`seriesOrdinalAccess` synthesize `verror` instances, detect empty/out-of-range scenarios and immediately return `value.NewNoneVal()` with `nil` error. Type validation (non-series arguments, non-integer indexes) continues to raise existing errors.
- Extend `seriesAt` to treat any index ≤0 or beyond remaining length as "missing" and return `none`, regardless of the current index. The same helper will be used by block/string/binary `at` natives so callers observe uniform behaviour from head or advanced positions.
- Leave mutation helpers (`poke`, `change`, `remove`, `copy --part`, etc.) untouched—they still rely on strict index validation so we avoid silently ignoring write mistakes.
- For path traversal, keep using `traversePath` but add a "lenient" mode triggered whenever `stopBeforeLast == false` (read-only evaluations). In this mode:
  - `traverseWordSegment` returns `none` without error if the field is absent after prototype lookup.
  - `traverseIndexSegment` short-circuits with `none` when the 1-based index is outside `[1..length]` instead of calling `checkIndexBounds`.
  - `ErrIDNonePath`, type mismatches, eval-segment issues, and all `set-path` traversals continue to surface as script errors so assignment semantics stay strict.
- Keep the `value.Series` implementations unchanged (they may still return errors internally); the native wrappers will translate those to `none`, confining the softer semantics to language-visible actions.

## Implementation Roadmap
1. **Contract Tests (TDD first):**
   - Update `test/contract/series_test.go`:
     - `TestSeries_First` / `TestSeries_Last`: expect `none` for empty series and add cases like `first tail data` → `none`.
     - `TestSeries_At`: convert all out-of-range expectations (index ≤0, >length, empty series) to `value.NewNoneVal()` results.
   - Update `test/contract/series_new_functions_test.go` for every ordinal (`second`→`tenth`): change "short series" cases from `wantErr` to `want == none`, and add explicit assertions for zero/negative indexes via `skip`/`next` combos if missing cases are not already covered.
   - Extend `path_eval_test.go` success table with scenarios like `obj.missing` and `data.10` returning `none`, and move prior `ErrIDNoSuchField` / `ErrIDOutOfBounds` cases out of the "errors" suite.
   - Revise `test/contract/objects_test.go` to treat missing-field reads (both direct path and `get-path`) as successful evaluations producing `none`, while keeping `none-path`, type mismatches, and assignment failures in the error table.
   - Consider adding a regression in `objects_test.go` to ensure `obj.missing` still returns `none` even when prototypes are involved.
2. **Series Native Adjustments:**
   - In `internal/native/series_polymorphic.go`:
     - Refactor `seriesFirst`/`seriesLast` to check `len == 0` or `err != nil` and return `none` early.
     - Rework `seriesOrdinalAccess` to guard `ordinal < 0` and `ordinal >= remaining length`; instead of returning `ErrIDOutOfBounds`, just return `value.NewNoneVal()` (with no error). Maintain existing logic for heads vs advanced indices.
     - Ensure `seriesSecond`–`seriesTenth` benefit automatically.
   - In `internal/native/series_helpers.go`:
     - Update `seriesAt` to return `none` when `zeroBasedIndex < 0`, when the series is empty, or when `absoluteIndex >= length` (regardless of `currentIndex`). Only series/type assertion errors should still error.
     - Keep `seriesPick` unchanged (already soft-fails) but review negative index handling to ensure it matches `seriesAt` semantics.
3. **Path Traversal Changes:**
   - Modify `traverseWordSegment` to return a `(core.Value, bool, error)` or similar so callers can distinguish "not found" without mapping to `ErrIDNoSuchField`. Alternatively, keep signature but return a sentinel `errMissingField` and let `traversePath` translate it to `none` when lenient.
   - Update `traverseIndexSegment` to avoid `checkIndexBounds` in lenient mode: compute `index < 1 || index > length` and signal "missing" rather than error.
   - Enhance `traversePath` to detect these sentinel conditions when `stopBeforeLast == false`, push `value.NewNoneVal()` onto `tr.values`, and break traversal. When `stopBeforeLast == true`, preserve the existing error propagation so `set-path` still validates intermediate containers.
   - Keep `checkIndexBounds` usage within assignment helpers (`assignToIndexTarget`) untouched for write contexts.
4. **Documentation Updates:**
   - Revise `specs/002-implement-deferred-features/contracts/objects.md` to describe that path lookups now yield `none` for missing fields while `set-path` keeps strict validation.
   - Consider annotating the relevant native docs in `internal/native/register_series.go` (e.g., `first`, `second`, `at`) to mention "returns `none` when no such element exists" for parity with `pick`.
5. **Validation & Cleanup:**
   - Run focused contract suites (`go test ./test/contract -run 'Series_|Path|Objects'`) before a full `go test ./...` to ensure behaviour changes are isolated.
   - Double-check that `ErrIDNoSuchField` is still reachable via `set-path` error cases; if not, document that it's now exclusive to write contexts.

## Integration Points
- Series actions are registered via `internal/native/register_series.go`; no API changes are needed, but doc strings there should stay accurate after the behaviour shift.
- Path traversal changes impact both `path` and `get-path` evaluation, as well as any native that internally builds `value.PathExpression`; ensure those callers still get the intended value/none without extra handling.
- Object field access also affects `select` for objects (already soft-fails); this change aligns path semantics with `select` to minimize surprise.

## Testing Strategy
- Use updated contract tests as the authoritative specification; they cover user-visible semantics across series natives and path lookups (blocks, strings, binaries, objects).
- Add regression tests when necessary (e.g., `first tail data`, `obj.profile.missing` returning `none`) to pin down corner cases.
- After code changes, execute `go test ./test/contract/...` and any affected internal packages (`internal/native`, `internal/eval`) to ensure no regressions.

## Potential Challenges
- Balancing "missing data" vs "invalid request": be explicit that type mismatches, negative counts, or traversal through `none` remain errors while simple absence returns `none`.
- Updating `traversePath` without breaking `set-path`: ensure the lenient shortcut is only applied to read paths, and continue surfacing helpful errors when assignments fail.
- Avoiding silent failures on mutation helpers (`poke`, `remove`, etc.), which must still error so users notice invalid writes.

## Viro Guidelines Reference
- Follow the mandated workflow: update contract tests first (TDD), then code, always using the `viro-coder` agent for interpreter changes and `viro-reviewer` afterward.
- Maintain coding conventions (no inline comments, constructor helpers such as `value.NewIntVal`, strict error categories).
- Keep documentation within the repo up to date (spec + native docs) to reflect the new, softer accessor semantics.
