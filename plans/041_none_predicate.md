# Plan 041: `none?` predicate

## Feature Summary
- Introduce a native predicate `none? value` that returns `true` only when the evaluated argument is the canonical `none` value and `false` otherwise.
- Expose the predicate through the existing native registration pipeline so it is available globally, documented via `help`, and follows the usual boolean-return semantics for predicates.
- Back the change with contract tests (TDD) and an updated data-natives spec entry to keep documentation and automation in sync.

## Research Findings
- `internal/native/data.go:60-83` already hosts the `type?` implementation (single-arg introspection) and is the natural location for another simple predicate that just inspects `args[0].GetType()`.
- `internal/native/register_data.go:26-185` binds all Data-category natives (incl. `type?`) with `value.NewNativeFunction`, parameter specs, and `NativeDoc` metadata that feeds the help system.
- Contract coverage for data natives lives in `test/contract/data_test.go:1-370`, using table-driven tests plus the `Evaluate` helper; this file is the right place for `TestData_NonePredicate`.
- `specs/001-implement-the-core/contracts/data.md:1-384` enumerates Data natives; adding a `none?` section keeps requirements authoritative and prevents checklist drift.
- `AGENTS.md` mandates TDD (tests/specs before runtime changes) and using the `viro-coder` / `viro-reviewer` agents for interpreter modifications.

## Architecture Overview
1. **Parser/Eval impact**: None; predicate is a standard native executed via `value.NewNativeFunction` and existing evaluator plumbing.
2. **Native implementation**: New `NoneQ` Go function returning `value.NewLogicVal(args[0].GetType() == value.TypeNone)` plus arity validation via `arityError` helper.
3. **Registration**: Add a single-parameter entry in `RegisterDataNatives` with `value.NewParamSpec("value", true)` and doc metadata (`Category: "Data"`). No dispatcher wiring beyond binding is needed.
4. **Documentation/spec**: Extend the Data contracts file to describe signature, semantics, examples, and tests; `NativeDoc` entry ensures `help none?` works automatically.
5. **Testing**: Contract tests evaluate small scripts to ensure `none?` returns logic! values for both true/false cases and respects arity errors.

## Implementation Roadmap
1. **Spec update (pre-code)**
   - Insert a new subsection under `specs/001-implement-the-core/contracts/data.md` describing `none?` (signature, parameters, truth table, sample scripts, and error cases e.g., arity mismatch). This keeps requirements first and documents expected behavior for future agents.
2. **Contract tests (TDD)**
   - Extend `test/contract/data_test.go` with `TestData_NonePredicate`, following the existing table-driven style.
   - Cover at minimum: `none? none` → `true`, `none? false` → `false`, `none? 0` → `false`, `none? []` → `false`, `none? (first [none 1])` → `true` to confirm normal evaluation.
   - Add a subtest ensuring the result type is `logic!` and a negative case for missing arguments (expect script error) if helpful.
3. **Native implementation**
   - Define `func NoneQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error)` in `internal/native/data.go` near `TypeQ`.
   - Enforce `len(args) == 1` via `arityError("none?", 1, len(args))` and return `value.NewLogicVal(args[0].GetType() == value.TypeNone)`; no additional helpers are required because `none` has a dedicated `ValueType` constant.
4. **Registration & docs**
   - In `RegisterDataNatives`, add a `registerAndBind("none?", value.NewNativeFunction(...))` entry with the single param spec, pointer to `NoneQ`, and a `NativeDoc` summarizing semantics (Category Data, Returns `[logic!]`, examples matching the tests, SeeAlso referencing `type?`).
   - Ensure the predicate is bound alongside other always-available Data natives so scripts can call it without extra imports.
5. **Verification**
   - Run focused tests: `go test ./test/contract -run TestData_NonePredicate` to iterate quickly.
   - Run the wider contract suite or `go test ./...` once the focused tests pass.
6. **Review workflow**
   - Delegate code edits to the `viro-coder` agent, then request a `viro-reviewer` pass before finalizing, per `AGENTS.md`.

## Integration Points
- **Spec ↔ tests**: The contract update should enumerate the same scenarios covered by `TestData_NonePredicate` to keep QA tooling aligned.
- **Native registry**: Binding via `RegisterDataNatives` ensures both the interpreter root frame and help/doc systems see the new predicate; forgetting this would leave the Go function unreachable.
- **Value helpers**: Uses only `core.Value.GetType()` and `value.TypeNone`, so no additional helper exposure is necessary, maintaining consistency with existing predicates such as `empty?` (series).

## Testing Strategy
- Table-driven tests verifying:
  - Inputs of different types return `logic!` booleans with expected truth table.
  - `none?` participates in expressions: e.g., `if none? none [1] [0]` yields `1` to ensure `logic!` output integrates with control flow.
  - Negative/edge: calling without arguments should raise an arity error (assert `wantErr: true`).
- Optional integration snippet: `Evaluate("values: reduce [none 42]\nnone? first values")` ensures the predicate handles evaluated block contents.
- After implementation, rerun nearby suites that touch data natives to guard against regressions.

## Potential Challenges
- **Arity enforcement**: Predicate must gracefully error on zero or multiple args; rely on `arityError` to keep messaging consistent.
- **Truthiness vs equality**: Ensure we only check `GetType() == value.TypeNone` rather than converting via `ToTruthy`, so `false` and `none` remain distinct.
- **Doc consistency**: Keep `NativeDoc` wording in sync with the new spec section to avoid conflicting help output; double-check SeeAlso references.

## Viro Guidelines Reference
- Follow `AGENTS.md` workflow: spec/tests first, code via `viro-coder`, review via `viro-reviewer`, and keep code comment-free.
- Constructor usage (`value.NewLogicVal`) and table-driven Go tests align with the style guides noted in `AGENTS.md` and `specs/001.../contracts/data.md`.
- Ensure new plan numbering stays sequential (`041`) and store plan in `plans/` per repo planning convention.
