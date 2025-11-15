## Feature Summary

Extend the `put` native so it can mutate `block!` association lists in addition to objects. New semantics mirror `select` on blocks: treat successive elements as key/value pairs, honor the current series index, accept any key/value types, append pairs when keys are missing, and remove pairs when `none` is assigned. Changes must be test-driven and documented.

## Research Findings

- `internal/native/data.go:476-518` hosts the current `Put` implementation which only accepts objects and word/string field names. Passing any other target type raises `typeError("put target", "object", targetVal)`.
- Block search semantics already exist in `internal/native/series_block.go:243-280` via `BlockSelect`. It matches keys by symbol when both operands are word-like and otherwise uses `core.Value.Equals`. However, it ignores the block's `Index` (always scans from the head) and returns `none` when a key appears without a corresponding value entry.
- Block values carry mutable backing slices (`value.BlockValue.Elements`) plus a current index pointer (`value.BlockValue.Index`). Removing or inserting elements by splicing `Elements` requires adjusting `Index` to keep it within `[0..len]`.
- `test/contract/series_test.go:858-905` only asserts block behavior for `select`, not `put`. There is no dedicated contract file for block mutation via `put`, so the new semantics currently lack coverage.
- `specs/002-implement-deferred-features/contracts/objects.md` documents `put` exclusively for objects, while `internal/native/register_data.go:287-315` describes and categorizes the native under "Objects" with object-focused examples.
- Helper candidates: `internal/native/series_helpers.go:isWordLike` and the comparison logic inside `BlockSelect` can be extracted to avoid duplicating word-matching rules. `series_polymorphic.go` already centralizes other series behaviors and is a good home for shared block key/value utilities if they aren't object-specific.

## Architecture Overview

- **Dual-mode `put`:** Keep the current object mutation path untouched, but extend the dispatcher in `Put` to detect `block!` targets. For blocks, delegate to a new helper (`putIntoBlock`) that encapsulates pair-aware mutation logic. This keeps object-specific trace instrumentation and manifest validation isolated from block operations.
- **Key matching:** Factor a reusable helper (e.g., `blockKeyMatches(key, sought core.Value) bool`) alongside `BlockSelect` so both `select` and `put` share identical comparison rules (word-like symbol comparison vs general `Equals`). This satisfies the "behave consistently with select" requirement and keeps future adjustments centralized.
- **Index-aware scanning:** When mutating a block, derive the starting search offset from `block.GetIndex()`. If the index points to a value element (odd offset), advance to the next key slot so earlier pairs remain untouched. Iterate in steps of two (`key` + `value`) until tail. This approach matches the "first occurrence after the current position" expectation without reordering data.
- **Update/append/remove logic:**
  - **Update:** When the key is found and has an existing value slot (`i+1 < len`), assign `newVal` into that slot.
  - **Append:** If the key is not found before tail, append the key and value (two elements) using `AppendValue` so the original block reference is mutated in place.
  - **Odd-length pairs:** If a key is found at the last element without a value partner, treat it as an incomplete pair: overwrite the key when removing, or append the missing value element when setting.
  - **Removal on `none`:** When `newVal` is `none`, remove both the key and the following value (if present). Use a dedicated helper to splice both `Elements` and `locations`, and adjust `Index` if the removal occurs before or overlapping the current index to maintain valid cursor positioning.
- **Return values:** Preserve the current contract of returning the assigned value. For removal (`none`), return `none` whether or not a matching key existed (since the caller supplied `none`).

## Implementation Roadmap

1. **Plan + Spec updates groundwork:** Capture this design in `plans/038_put_block_support.md` (this file) and note the upcoming behavioral change in `specs/002-implement-deferred-features/contracts/objects.md` so documentation aligns with the new capabilities before coding starts.
2. **TDD – Contract tests:**
   - Create `test/contract/block_put_test.go` (or extend an existing nearby suite) with table-driven cases covering:
     - Updating an existing key/value pair at head (`blk: [a 1 b 2] put blk 'a 99` ⇒ `blk = [a 99 b 2]`, result `99`).
     - Respecting the current index (`blk: next [a 1 a 2]` should update the second `a`). Include both update and removal scenarios to assert previous pairs remain intact.
     - Appending when key is absent (`put blk 'c 3` ⇒ block grows by two elements; ensure function returns `3`).
     - `none` removal when key exists (`put blk 'a none` ⇒ `[b 2]`) and when key is missing (block unchanged).
     - Odd-length blocks (e.g., `[a 1 b]`): updating `b` should append the missing value slot rather than duplicating the key; removing `b` should drop the dangling key without error.
     - First-occurrence-only behavior and multi-type keys (words vs strings vs integers). Include a case where later duplicates remain unchanged.
     - Mixed key/value types (e.g., key as string!, value as object!) to confirm there are no type limitations.
     - Return-value assertions for each path to ensure the native returns exactly the supplied third argument (or `none`).
   - Reuse `Evaluate` helper from other contract tests to assert both returned value and resulting block state. Prefer explicit `same?` or shape checks to confirm in-place mutation.
3. **Helper refactor:**
   - In `internal/native/series_block.go`, extract the comparison logic from `BlockSelect` into a new unexported helper (e.g., `blockKeyMatches(key, sought core.Value) bool`). Optionally, add a second helper to compute the first key index at or after a given cursor (`firstKeyIndexFrom(block *value.BlockValue, start int) int`). Ensure `BlockSelect` is updated to use these helpers so `select` also gains index-aware behavior if we decide to adopt it now (or document why it remains unchanged).
   - If aligning `BlockSelect` with cursor awareness is desirable, update it here; otherwise, clearly scope the helper so `Put` alone adjusts indices while `select` retains previous semantics (mention this in review notes).
4. **Block mutation logic:**
   - Extend `Put` in `internal/native/data.go` to branch on `targetVal.GetType()`. For block targets:
     - Convert to `*value.BlockValue` via `value.AsBlockValue`.
     - Invoke a new helper (e.g., `putBlockPair(block, key, newVal) (core.Value, error)`) that implements the algorithm described above.
     - The helper should:
       - Determine the iteration start index respecting `block.Index` parity.
       - Loop over key/value pairs using the shared comparison helper.
       - On match with `newVal` ≠ `none`, either assign or append missing value.
       - On match with `newVal` = `none`, remove key/value (handling odd tail) and adjust `block.Index`.
       - On miss, append pair unless `newVal` = `none` (no action).
       - Return the third argument (or `none`).
     - Ensure the helper mutates `block.Elements` and `block.locations` consistently—mirror the slice-manipulation style from `blockTrim*` functions or add a small removal helper to keep code readable.
   - Keep the existing object code path as-is (including trace events) to avoid regressions.
5. **Docs + native metadata:**
   - Update `internal/native/register_data.go` so the `put` doc string describes block behavior (parameters may now include `block!` targets and "key" instead of "field"). Add an example demonstrating `put [a 1] 'b 2` and removal with `none`.
   - Revise `specs/002-implement-deferred-features/contracts/objects.md` to include a "block association list" subsection under `put`, detailing alternating key/value semantics, `none` removal, and index-respecting search.
   - If `docs/viro_architecture_knowledge.md` or `docs/literal-sharing-semantics.md` mention `put`, add a short note referencing the new capability if relevant.
6. **Validation:** Run targeted and full test suites:
   - `go test ./test/contract -run 'BlockPut|Series_Select'` to iterate quickly on the new contract file and any select changes.
   - `go test ./...` (or at least `go test ./internal/native ./test/contract/...`) before finalizing to ensure no other packages regressed.

## Integration Points

- **Native registration:** `put` remains bound via `register_data.go`, but its doc metadata and parameter descriptions must now mention `block!` support. No changes are required to type frames or `CreateAction` since `put` is not a polymorphic action.
- **Series helpers:** Any new helper used by both `BlockSelect` and block `put` should live in `series_block.go` (or another shared file) to avoid duplication and keep block-specific logic centralized.
- **Trace + evaluator:** Object mutation still emits trace events (`trace.TraceObjectFieldWrite`). Block mutation does not currently emit trace events; document this decision so reviewers know it's intentional.
- **Spec + docs:** Updating `specs/002...` ensures future features referencing `put` will account for block semantics. If `examples/*.viro` demonstrate object mutation via `put`, consider adding a block example later (not required for this change but noted for future documentation work).

## Testing Strategy

- Primary validation through the new `test/contract/block_put_test.go` suite. Include head/tail/index coverage, append/removal flows, and mixed key/value types.
- If `BlockSelect` gains cursor-awareness or helper changes, update `TestSeries_Select` accordingly (maintain parity between select and put semantics).
- After unit tests pass, execute `go test ./test/contract -run 'BlockPut|Series_Select'` followed by `go test ./...` to ensure global stability.

## Potential Challenges

- **Index math bugs:** Removing or appending elements while respecting the current index can easily leave `block.Index` pointing past the tail. Always clamp the index after splicing.
- **Odd-length handling:** Ensure that finding a key without a following value does not panic or mis-append; treat the missing value as needing insertion (for updates) or just remove the key (for deletions).
- **Shared helper risk:** Refactoring `BlockSelect` to use a new helper must not accidentally change its observable behavior unless accompanied by matching tests/documentation updates.
- **Mutation verification:** Tests must confirm that `put` mutates the original block reference; forgetting to operate on the same pointer (e.g., by cloning) would violate the requirement.

## Viro Guidelines Reference

- Follow the mandated workflow: write contract tests first (`test/contract/block_put_test.go`), then implement interpreter changes via the `viro-coder` agent, and finish with a `viro-reviewer` pass.
- Honor style rules from `AGENTS.md`: no inline comments in Go code, use constructors like `value.NewNoneVal()`, and keep word comparisons consistent with `isWordLike` helpers.
- Update specs/docs alongside code so feature documentation matches implementation, and keep commit/test commands aligned with the recommended `go test ./...` workflow.
