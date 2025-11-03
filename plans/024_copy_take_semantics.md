# 024 Copy / Take Semantics & Series Count Validation

## Context
Recent polymorphic series refactor introduced actions:
- copy (with --part refinement)
- take (count argument)
- remove (with --part refinement)

Current behaviors:
- copy --part N: validatePartCount enforces 0 <= N <= total length; CopyPart then clamps N to remaining elements from current index.
- take N: TakeCount clamps (negative -> 0, oversized -> remaining).
- remove --part N: validatePartCount enforces 0 <= N <= total length; RemoveCount errors when index+count exceeds length. Negative / oversized counts raise OutOfBounds.

Inconsistencies:
1. Negative counts:
   - copy/remove: produce OutOfBounds error via validatePartCount.
   - take: silently coerces negative to 0 (returns empty series) – inconsistent.
2. Error argument shape for OutOfBounds varies (some empty args, some include requested & length).
3. Advanced index copy semantics rely on clamp (acceptable) but not explicitly documented.
4. Lack of test coverage for take edge cases (negative, zero, oversized, advanced index, tail position).

## Goals
- Standardize count validation semantics across copy/remove/take.
- Clarify clamp vs error rules.
- Ensure consistent OutOfBounds error argument payload.
- Extend contract tests to cover take & copy edge cases and error args.
- Document semantics in native registration docs.

## Decisions
1. Count Domain:
   - All count-taking operations (copy --part, remove --part, take N) treat count as an absolute non-negative integer.
   - Negative count always raises OutOfBounds.
2. Oversized Count (greater than remaining from current index):
   - copy --part: allowed; result clamps to remaining (keeps existing tests). Validation only checks against total length, not remaining; clamp acts after validation.
   - take N: allowed; result clamps to remaining. (Change: remove negative->0 coercion; keep oversized clamp.)
   - remove --part: error if index+count exceeds length (destructive op must be precise). (No change.)
3. Zero Count:
   - copy --part 0 → empty series of same type.
   - take 0 → empty series of same type.
   - remove --part 0 → no-op, returns original series (already implemented & tested).
4. Error Arguments Standardization (OutOfBounds ErrIDOutOfBounds):
   - Args layout: [requested-count-or-index, series-length, current-index]
   - For validatePartCount: current-index included (series.GetIndex()). For negative counts, requested is the negative value.
   - For removal when index+count overflow: requested-count, length, current-index.
   - For back at head: requested-index (currentIndex-1), length, current-index.
   - For at action invalid index: requested-index, length, current-index (where current-index is 0 because not series-relative; we may supply "" for third arg in non-relative operations – decision: keep third empty for absolute index ops). Summary:
     - Relative (depends on current position): third arg = current-index.
     - Absolute (global index passed explicitly): third arg = "".
5. Empty Series Errors (ErrIDEmptySeries): keep current form; first/last supply context string already.
6. Take Negative Behavior Change:
   - Modify TakeCount implementations (StringValue, BlockValue, BinaryValue) to return error instead of clamping for negative; but TakeCount currently returns Series (no error). Approach: shift validation into seriesTake native action before calling TakeCount.
   - seriesTake: validate count >= 0, else OutOfBounds error with args [count, length, currentIndex]. Keep TakeCount clamp logic but remove negative branch (optional). Simpler: adjust seriesTake only (no interface change) to maintain minimal code churn.

## Implementation Plan
Step 1: Tests (contract/test)
- Add new tests in series_test.go:
  - take negative count error (block & string). Expect ErrIDOutOfBounds args ["-1", length, currentIndex].
  - take zero count returns empty series of correct type.
  - take oversized count clamps (from head and from advanced index).
  - take from tail position returns empty series.
  - copy --part error arg validation: ensure negative count returns standardized 3 args (include current index).
  - remove --part error arg shape updated (index+count overflow) to include current index.
  - back at head error arg shape updated (third arg current index 0).
- Adjust any existing assertions expecting 2 args to now expect 3 where applicable.

Step 2: Code Changes
- native/series_polymorphic.go: in seriesTake, add validation similar to validatePartCount but using new signature validateCountNonNegative(series, count). Error if count < 0 (with args). Keep oversized semantics (no error; rely on clamp in TakeCount).
- native/series_helpers.go: extend validatePartCount to include current index in args: [requested, totalLength, currentIndex].
- remove update: when validatePartCount fails or RemoveCount returns error due to overflow, ensure error args use [requestedCount, totalLength, currentIndex]. (RemoveCount internal errors need mapping – intercept error and wrap with standardized ScriptError.)
- back action OutOfBounds error: change args to [fmt.Sprintf("%d", currentIndex-1), fmt.Sprintf("%d", length), fmt.Sprintf("%d", currentIndex)].
- at action: keep third arg blank (absolute indexing semantics) – no change.

Step 3: Native Docs
- register_series.go: document count semantics for copy/remove/take (negative errors; oversized clamp differences; zero behaviors; error arg shape).

Step 4: Consistency Sweep
- Search for ErrIDOutOfBounds constructions in native series code and update args layout per rules.
- Ensure tests for previous OutOfBounds errors updated accordingly (back head, copy negative, remove negative/out-of-range, at invalid index maybe unchanged).

Step 5: Optional Refactor (Low Priority)
- Consolidate count validation into a single helper validateNonNegative(name string, count int, series Series, relative bool) to produce standardized errors. (Defer if increases churn.)

## Test Cases Detail
1. take negative:
   input: `data: [1 2 3]\n take data -1`
   expect: error ID OutOfBounds args ["-1", "3", "1"] (current index at head = 0 internally ⇒ presented as 1? Decide: current index raw (0) or 1-based? Current system returns 0-based for internal operations. Decision: expose raw zero-based; tests will assert "0". -> Need clarity: existing index? seriesIndex returns 1-based externally. For error args, we will use raw zero-based to avoid conversion overhead; document this.)
   Final: args ["-1", "3", "0"].
2. take oversized from advanced index:
   input: `data: [1 2 3 4 5]\n data: next next data\n take data 10`
   expect series [3 4 5].
3. take zero returns empty:
   input: `take [1 2 3] 0` expect `[]`.
4. take from tail:
   input: `data: [1 2 3]\n data: skip data 3\n take data 2` expect empty block.
5. copy negative (already present) update args to include current index.

## Open Questions (resolve before implementation)
1. Error args index base (0 vs 1). Decision: Use zero-based internal index for third arg to reflect actual slice position; external index queries (index?) remain 1-based.
2. Should validatePartCount be tightened to remaining length for copy? Decision: No – we retain clamp semantics enabling larger count relative to remaining without error when within total length.

## Risks
- Changing error arg shapes may break existing tests relying on previous layout; addressed by updating tests.
- Confusion over index base in errors vs index? action; mitigated by documentation update.

## Acceptance Criteria
- All new tests pass.
- No regression in existing series tests beyond expected error arg adjustments.
- Documentation updated with semantics and error arg layout.
- Negative take now errors; other counts behave as specified.

## Effort Estimate
- Tests: Small (30-45 min)
- Code changes: Small (30 min)
- Docs update: Small (15 min)
- Refactor helper (optional): Medium (1 hr)

## Next Steps
1. ✅ Implement tests (contract) per plan.
2. ✅ Run test suite; confirm failures correspond to planned code changes.
3. ✅ Implement code changes using viro-coder agent.
4. ✅ Re-run tests; update docs.
5. Review via viro-reviewer agent.

## Implementation Complete
All steps have been successfully implemented following TDD principles:
- Tests were added first and confirmed to fail
- Code changes were implemented to make tests pass
- Documentation was updated
- All existing tests continue to pass
- Error argument standardization is complete across copy/remove/take operations

