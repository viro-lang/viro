# Plan 049: Nested Block Sort

## Feature Summary
- Extend the `sort` native for block!/paren! series so it can order blocks whose elements are themselves blocks, using lexicographic comparison semantics analogous to string sorting.
- Permit nested blocks to contain arbitrary value types; comparisons only occur when both sub-blocks provide a value at the same position, and those values must share the same type.
- Emit `verror.ErrIDNotComparable` whenever two compared positions either have mismatched types or hold a shared type for which the runtime cannot define an ordering.
- Preserve existing behaviour for integer and string blocks, as well as sort error cases for mixed top-level types or mismatched series types.

## Research Findings
- `internal/native/series_block.go` implements `BlockSort`, currently restricted to integer or string elements (lines 53-72). It guards against mixed types with `ErrIDNotComparable`.
- `value.SortBlock` (internal/value/block.go:330+) sorts block elements via `sort.SliceStable` but only knows how to compare integers and strings.
- `sort` is registered for both `block!` and `paren!` in `internal/native/register_series.go`, so any change must work for both series types without duplicating logic.
- Contract expectations live in `test/contract/series_test.go` (see `TestSeries_SortReverse` around lines 2160-2220); there are currently no tests covering nested block comparisons.
- Error identifiers come from `internal/verror`, and `ErrIDNotComparable` with context `{"sort", "mixed types", ""}` is already used for invalid block sort inputs.
- No general-purpose value comparison utility exists today; lexicographic behaviour must be implemented inside the value layer (likely `internal/value/block.go`).

## Architecture Overview
- Keep `BlockSort` responsible for validating the top-level block (type checking, empty short-circuit, refinement handling) and for constructing any comparison schema necessary to preflight nested blocks.
- Introduce a recursive "comparison schema" that records the expected type per element index at each nesting depth. When processing the list to sort, traverse each inner block to ensure that whenever two blocks expose elements at the same depth/index, those elements share the same `core.ValueType`. Nested block positions maintain their own schema nodes.
- Extend `value.SortBlock` with lexicographic comparisons for nested block/paren elements. Use helper functions (`compareBlockLex`, `compareScalarValuesForSort`) to encapsulate recursion and scalar comparisons, returning ordering consistent with Viro semantics (numerics by value, strings/binaries lexicographically, word-like values by symbol, logic false<true, etc.). Unsupported scalar types should trigger a comparison error detected during schema validation.
- Maintain sort stability (Go's `sort.SliceStable`) even though the user called out that stability is unnecessary. Existing behaviour already uses the stable variant, so leave it untouched to minimize behavioural changes elsewhere.

## Implementation Roadmap
1. **TDD – Contract Coverage (`test/contract/series_test.go`):**
   - Augment `TestSeries_SortReverse` with new cases before the current error table entries:
     - `name: "sort block of blocks lexicographic"` using data like `data: [[2][1 5][1 2 3]]` and assert the result is `[[1 2 3] [1 5] [2]]` (shorter block sorts first after equal prefixes).
     - `name: "sort block of heterogeneous inner values"` verifying `[ [1 "a"] [1 "b"] [2 [3]] ]` sorts correctly even when later positions are ignored.
     - `name: "sort nested blocks deeper levels"` ensuring recursion works, e.g. `[[1 [2 3]] [1 [2 4]] [1 [1]]]`.
   - Add explicit negative tests:
     - `name: "sort nested blocks mismatched element types"`, `input: "sort [[\"a\"] [1]]"`, expecting `ErrIDNotComparable`.
     - `name: "sort nested blocks incomparable scalar"`, e.g. two inner blocks each containing a `function!` at the same position to assert we surface `ErrIDNotComparable` and a helpful context string ("unsupported type function!").
     - `name: "sort nested blocks mismatched nested type"`, such as `sort [[1 [2]] [1 "a"]]` to ensure nested schema catches block vs string inconsistency when the nested segment needs to be compared.
   - Keep existing integer/string/binary test cases intact to guard regressions.

2. **Schema Validation Helper (new functions in `internal/native/series_block.go` or a dedicated helper file in `internal/native`):**
   - Introduce a structure like `type blockComparisonSchema struct { positions map[int]*schemaNode }`, where `schemaNode` records the expected `core.ValueType` and (if the type itself is block/paren) a nested `blockComparisonSchema` for deeper levels.
   - Implement `func buildBlockSchema(blocks []*value.BlockValue) (*blockComparisonSchema, error)` that:
     - Iterates over every top-level block element, ensuring each is `TypeBlock` or `TypeParen`.
     - Recursively records the type for each positional index; if another block supplies a value at the same (depth,index) with a different type, return `verror.ErrIDNotComparable` describing the mismatch (e.g., `{"sort", "mismatched types at position 2", "block! vs string!"}`).
     - Tracks whether a `schemaNode` represents a scalar type that the comparator knows how to order; if not, immediately error with context `{"sort", "unsupported type", value.TypeToString(t)}` so we fail before invoking the sorter.
   - Ensure the helper treats `TypeBlock` and `TypeParen` interchangeably when building nested schemas. Other series types (string/binary) should be treated as scalars.

3. **Lexicographic Comparator Enhancements (`internal/value/block.go`):**
   - Refactor `SortBlock` to delegate based on the detected element type (integer, string, block/paren). Leave the existing integer/string comparator now factored into helper closures for reuse.
   - Introduce `func compareBlockValuesLex(a, b *BlockValue) int` that:
     - Iterates index-wise up to `min(len(a.Elements), len(b.Elements))`.
     - For each position, dispatch to a scalar comparator when type is integer/decimal/string/binary/logic/word/path/etc. If the type is block/paren, recurse.
     - When scalar values differ, return comparison result; when elements are equal, continue; when all compared elements match, shorter block precedes longer.
   - Add `func compareScalarForSort(a, b core.Value) (int, error)` that covers supported types:
     - integers via `AsIntValue`.
     - decimals via `value.AsDecimal` and `Cmp` on magnitudes.
     - strings via rune comparison on `Form()` (to stay consistent with existing sort semantics).
     - binary via byte slices.
     - logic (false<true), word-like values (compare symbol strings), `none` (always equal, but mismatched lengths still break ties), and potentially datatype! by name.
     - For unsupported types (function!, object!, port!, etc.), return an error so caller can convert it to `ErrIDNotComparable` before sort begins.
   - Modify `SortBlock` to accept the schema (if needed) or, alternatively, rely on the pre-validation to guarantee comparator never receives unsupported types. If the comparator still surfaces an unexpected type mismatch, panic or recover is unacceptable; convert it into a best-effort fallback by checking the type inline and returning deterministic ordering (or degrade to `Form()` comparison) only if spec explicitly allows.

4. **`BlockSort` Integration (`internal/native/series_block.go`):**
   - Update the type guard logic:
     - Continue short-circuiting empty blocks.
     - Distinguish the first element type: if integer or string, keep the existing validation (ensuring every element matches the type).
     - If block or paren, assert every element is block-ish, then pass the slice into the new schema builder. On error, propagate `verror.ErrIDNotComparable`.
     - Reject any other element type combinations with the current `mixed types` error.
   - Update the call to `value.SortBlock` to capture potential errors if the signature is adjusted (if `SortBlock` remains void, ensure schema builder handles all failure modes ahead of time).
   - Ensure error contexts remain meaningful (operation name `sort`, reason `mixed types` or `unsupported type`, third slot describing the conflicting types or position).

5. **Optional Documentation Touchpoints:**
   - After code/test changes pass, consider adding a short note to `docs/literal-sharing-semantics.md` or `specs/002-implement-deferred-features/tasks.md` describing the new nested-block sorting capability, if the team maintains feature parity docs.
   - Update inline native documentation snippet within `register_series.go` (sort entry) to mention "Lexicographically sorts block!/paren! values, including nested block elements" so CLI help stays accurate.

6. **Validation & Tooling:**
   - Run `go test ./test/contract -run TestSeries_SortReverse` frequently during development, followed by `go test ./test/contract/...` once implementation is complete.
   - Execute `go test ./internal/native -run TestBlockSort` if such unit tests exist; otherwise rely on contract coverage plus `go test ./internal/value` to ensure comparator helpers compile and existing functionality remains intact.

## Integration Points
- `internal/native/register_series.go` automatically applies to both `block!` and `paren!`, so schema validation and sorting logic must treat both identically.
- `value.SortBlock` is invoked only from `BlockSort` today, but any other callers added later will inherit the new behaviour—keep the API stable and document that it now handles nested blocks.
- Error propagation flows through `verror`; ensure new helper functions exclusively create script errors via `verror.NewScriptError` so upstream callers do not need special handling.

## Testing Strategy
- Primary verification via new/updated contract cases inside `TestSeries_SortReverse` to lock in success and failure semantics for nested block sorting.
- Add regression-style snippets to `examples/` or `docs/` if desired to demonstrate the new capability (optional but useful for manual QA).
- Post-implementation, run `go test ./...` (or at least `./test/contract` + `./internal/value`) to catch unintended regressions in value comparisons or other series operations.

## Potential Challenges
- Designing a comparison schema that both enforces type consistency and allows optional positions without false positives; ensure mismatches are only flagged when two blocks need to compare a specific index.
- Providing deterministic ordering for scalar types beyond integers/strings without contradicting existing language semantics; confirm with Viro guidelines before inventing new orderings for exotic types.
- Avoiding panics within the sorter; all error conditions must be detected before `sort.SliceStable` executes, because its callback cannot return an error.
- Maintaining performance for large nested blocks—the schema builder should be linear in the total number of elements to avoid O(n²) validation.

## Viro Guidelines Reference
- Follow the mandated workflow: write/adjust contract tests first (TDD), implement interpreter changes via the `viro-coder` agent, and run `viro-reviewer` afterward (per `AGENTS.md`).
- Respect coding conventions: no inline comments, always construct values via helpers (`value.NewIntVal`, `value.NewBlockVal`), and use `verror.NewScriptError` with accurate context tuples.
- Keep series-native code in `internal/native` free of allocations when possible, and avoid mutating shared state outside the provided block unless explicitly required.
