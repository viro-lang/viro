# Plan 037: Foreach Object Iteration

## Feature Summary

- Extend `foreach` so it can iterate `object!` values directly, binding keys (and optionally values) per iteration while preserving existing behavior for series arguments.

## Research Findings

- `internal/native/control.go:633-754` currently restricts `foreach` to `value.IsSeries` types and consumes `value.Series` via `GetIndex`/`ElementAt`. Multi-word var blocks chunk the series and bind trailing vars to `none`.
- `test/contract/control_test.go:614-756` covers `foreach` success/error cases and `--with-index` refinement but only for series inputs.
- Object layout (`internal/value/object.go:16-180`) stores fields in frames and exposes `GetAllFieldsWithProto` for manifest+prototype ordering; reflection natives (`internal/native/reflection.go:131-190`) already rely on this ordering, and docs note `words-of`/`values-of` guarantee consistent order (`docs/observability.md:320`).
- No existing spec text covers `foreach`; `specs/001-implement-the-core/contracts/control-flow.md` only describes `when/if/loop/while`, so new semantics must be documented there (and anywhere else referencing `foreach` such as CLI help in `internal/native/register_control.go:145-175`).

## Architecture Overview

- Keep current `value.Series` flow untouched; introduce an alternate object iteration path that activates when `args[0]` is `object!`.
- For objects, derive the iteration order from a snapshot of field names (`obj.GetAllFieldsWithProto()` → `[]core.Binding`). Record only the word order; fetch current values each iteration via `obj.GetFieldWithProto(name)` so mutations in the body are visible while still iterating a fixed key set.
- Per iteration, bind variables as follows:
  - Single variable (word or single-word block) receives the key name (word value).
  - Two or more variables: first gets key, second gets the current field value, remaining names get `none` (mirrors existing “not enough source values” semantics and satisfies odd var-count requirement).
- The `--with-index` refinement continues to bind a zero-based iteration counter per object entry.
- Break/continue handling and return semantics remain unchanged; only the loop that feeds variables differs depending on argument type.

## Implementation Roadmap

1. **Spec/Docs Prep**
   - Extend `specs/001-implement-the-core/contracts/control-flow.md` (or create a `foreach` section if missing) to describe object iteration semantics, including key order, multi-variable binding rules, prototype inclusion, and `--with-index`.
   - Update user-facing docs/help (`internal/native/register_control.go`, potentially `docs/repl-usage.md` or `docs/execution-model.md`) with examples that show `foreach` over objects.
2. **Contract Tests (TDD)**
   - Add new table-driven cases to `test/contract/control_test.go`:
     - Single-word iteration over an object returns keys in manifest/proto order.
     - Block `[key value]` form binds both pieces.
     - Extra variables (e.g., `[key value extra]`) receive `none`.
     - `--with-index` works with object iteration.
     - Empty object yields `none` without executing body.
     - Prototype fields appear before child overrides; overriding values reflect latest bindings.
     - Error cases remain unchanged (non-object + non-series still error).
3. **Foreach Implementation**
   - Refactor `internal/native/control.go:633-754`:
     - After parsing var names, branch on `seriesVal.GetType() == value.TypeObject`.
     - Extract ordered field names once, skip if zero length.
     - Iterate names; per iteration bind vars as described and evaluate body. Maintain `iteration` counter shared with index refinement.
     - Keep existing `value.Series` logic for non-object series; update the type-mismatch error string to mention `object!`.
     - Consider extracting binding code into helper functions to avoid duplication between branches.
4. **Helper/Utility Enhancements**
   - If needed, add a small helper (internal-only) to convert `[]core.Binding` to ordered key slices while ensuring duplicates are deduplicated the same way as `GetAllFieldsWithProto`.
   - Reuse `value.NewWordVal` and `currentFrame.Bind` mechanisms already present.
5. **Documentation & Release Notes**
   - Add a release-note entry (e.g., `RELEASE_NOTES.md`) referencing object iteration support.
   - Ensure any relevant guides or examples mention the new capability.

## Integration Points

- `internal/native/control.go`: main logic change; must respect existing frame binding and continue/break handling.
- `test/contract/control_test.go`: add new contract cases (and update existing type-mismatch test message if wording changes).
- `specs/001-implement-the-core/contracts/control-flow.md`: document `foreach` behavior, especially object semantics.
- `internal/native/register_control.go`: CLI help examples to showcase object usage.
- Optional docs (`docs/repl-usage.md`, `examples/*.viro`) to illustrate usage.

## Testing Strategy

- Contract tests described above; especially:
  - Keys iteration order equals `words-of` output (compare blocks).
  - Prototype scenario: child overrides parent field and verifies value seen.
  - Mutation during iteration: change a field value inside body and assert subsequent iterations read updated values (ensures runtime `GetFieldWithProto` lookup).
  - `--with-index` binding increments correctly and resets between loops.
  - Ensure legacy series behavior unaffected (regression tests already exist).
- If needed, add integration test script under `test/integration/` to run a short scenario mixing series and object iterations.

## Potential Challenges

- **Value snapshots vs live reads**: storing `binding.Value` would miss later mutations; mitigate by storing only field names and calling `GetFieldWithProto` each time.
- **Prototype overrides**: `GetAllFieldsWithProto` already deduplicates; ensure added helper does not reintroduce duplicates or reorder fields.
- **Var-count semantics ambiguity**: document and test the “extra vars receive `none`” rule so behavior is deterministic.
- **Performance**: `GetFieldWithProto` walks prototypes; acceptable for iteration counts equal to object field count, but avoid repeated `GetAllFieldsWithProto` recomputation mid-loop.
- **Error messaging**: updating the “series only” error text must keep existing tests in sync.

## Viro Guidelines Reference

- Follow TDD mandate from `AGENTS.md`: write `test/contract` cases before touching `internal/native/control.go`.
- Use `viro-coder` for code edits and `viro-reviewer` for review per workflow policy.
- Keep code comment-free per style guide; rely on clear structure instead.
- Maintain constructor usage and binding patterns (e.g., `value.NewWordVal`, `currentFrame.Bind`); avoid direct struct field access outside approved APIs.
- Document behavior changes in specs before implementation to stay aligned with specification-first process.
