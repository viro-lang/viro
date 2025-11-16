# Plan 040: `fn --no-scope`

## Feature Summary
- Extend the `fn` native with a `--no-scope` refinement so callers can opt into executing a function body within the caller's frame instead of pushing a new one (e.g., `fn --no-scope [a] [a + b]`).
- When the refinement is present, parameter bindings, refinement locals, and any subsequent assignments should reuse the caller's scope so side effects are visible immediately and the body can reference call-site locals.
- Maintain existing behavior (local-by-default) when the refinement is omitted; avoid widening evaluator responsibilities beyond this toggle.

## Research Findings
- `internal/native/function.go:25-67` parses fn parameters, clones the body block, and captures the defining frame via the `frameProvider` evaluator interface; currently no refinements or metadata beyond `Parent` are stored.
- `internal/native/register_control.go:181-204` registers `fn` with two positional parameters and a documentation entry lacking refinement details; this is the hook for exposing `--no-scope` in help output.
- `internal/value/function.go:18-127` defines `value.ParamSpec` and `value.FunctionValue`; `FunctionValue` lacks a flag for scope behavior today, so new metadata (e.g., `NoScope bool`) has to be added here and propagated through constructors and helpers.
- `internal/eval/evaluator.go:907-933` (`executeFunction`) always allocates a new `frame.FrameFunctionArgs`, binds parameters (`bindFunctionParameters`), evaluates the body, and guarantees cleanup—this is the critical place to branch when `NoScope` is requested.
- `docs/scoping-model.md` and `specs/001-implement-the-core/contracts/function.md` describe local-by-default semantics; both need updates clarifying the opt-in escape hatch and its trade-offs.
- Existing scoping regressions/tests (`test/contract/native_scoping_test.go`, `test/contract/function_eval_test.go`, `test/contract/function_test.go`) give coverage for local scopes, refinement parsing, and block cloning but do not exercise dynamic scope reuse.

## Architecture Overview
1. **Function metadata**: Introduce a boolean on `value.FunctionValue` (e.g., `NoScope bool`) plus constructor plumbing so `fn --no-scope` instances carry this intent. Keep default `false` to avoid touching existing call sites.
2. **Native surface**: Update `value.NewNativeFunction` registration for `fn` to include a `value.NewRefinementSpec("no-scope", false)` and modify `native.Fn` to inspect `refValues` and set the metadata flag. The refinement behaves like a boolean flag (no payload, defaults to `false`).
3. **Execution mode switch**:
   - Extend `Evaluator.executeFunction` with a guarded branch: when `fn.NoScope` is true, obtain the current frame (falling back to root if needed) and execute the body without allocating a new `FrameFunctionArgs`.
   - Parameter/refinement binding must occur directly on the caller frame. Introduce a helper that: (a) snapshots any pre-existing bindings for affected names (value + existence), (b) `Bind`s/`Set`s the argument values, (c) defers restoration of previous bindings only if the caller still wants isolation for positional names, or explicitly chooses to leave them mutated (decision described below).
   - Because `Return` relies on `*ReturnSignal`, the execution branch can still call `DoBlock` and convert `ReturnSignal` to result exactly like the current path.
4. **Scope sharing strategy**:
   - *Option A (full reuse)*: Do not restore positional/refinement words. Every `Bind` persists exactly as the body left it. This maximizes transparency but leaks parameter names into the caller (acceptable per "reuse scope entirely" directive). Tests will codify this.
   - *Option B (temporary params)*: Restore only the pre-call bindings for formal parameters/refinements while leaving any other body assignments in place. This keeps parameter leakage minimal and still honors body side effects. Following maintainer guidance, **Option B is confirmed** so parameters do not remain bound after the call, while other scope mutations remain visible.
5. **Closure handling**: Even with shared scopes, functions should still respect captured lexical parents when a word is missing in the caller frame. To achieve this without touching evaluator internals, wrap the caller frame in a lightweight `sharedFrame` that implements `core.Frame` by delegating `Bind`/`Set` to the caller while overriding `GetParent` to point at the function's captured parent instead of the caller's parent. The wrapper can hold both the caller frame index (for assignment) and the captured parent index, giving lookup chains access to both the call-site chain and lexical closure (call-site frame becomes the first hop; the wrapper's `Parent` is `fn.Parent`).
6. **Documentation/help**: Expand `NativeDoc` for `fn` to describe `--no-scope`, including warning language about shared scope semantics and examples. Update scoping docs/specs accordingly.

## Implementation Roadmap
1. **Spec & docs first (TDD)**
   - Update `specs/001-implement-the-core/contracts/function.md` with a new subsection describing the refinement, execution semantics, examples (mutating caller state, accessing invocation locals), and error constraints.
   - Extend `docs/scoping-model.md` (and any other relevant dev guides) with an "Opting out via --no-scope" section covering use cases, trade-offs, and best practices.
   - Amend `register_control.go`'s `NativeDoc` entry so `help fn` lists the new refinement definition.
2. **Tests** (`test/contract/`)
   - Add/extend contract tests (likely a dedicated `function_no_scope_test.go` or additions under `native_scoping_test.go`) verifying:
     - Assignments inside `fn --no-scope` affect outer variables (including parameter words).
     - Body can read caller-local words defined after the function definition but before invocation.
     - Nested/shared-scope functions properly mutate the expected frame when invoked from inside other functions (dynamic scope demonstration).
     - Recursion/rest reentry either supported (with parameter restoration) or explicitly documented as undefined; tests should codify whichever approach we adopt.
     - Interactions with refinements/argument order remain intact when the refinement is absent (regression test).
3. **Runtime changes**
   - Modify `internal/value/function.go` to carry the new flag and expose setters/getters if needed.
   - Update `value.NewUserFunction` invocation in `native.Fn` to accept the flag; ensure clones/equals/mold handle the new field.
   - Implement the shared-frame helper (new file in `internal/frame` or `internal/eval`) and integrate it into `executeFunction`.
   - Adjust `executeFunction` logic to branch on `fn.NoScope`, invoke the helper to bind/unbind params, and skip the `PushFrameContext`/`defer pop` path while preserving `where` stack info.
   - Ensure `bindFunctionParameters` can target the shared frame (may need to export helper or reuse existing function).
4. **Help/UX polish**
   - Add examples to `examples/*.viro` or `docs/repl-usage.md` if necessary.
   - Verify `viro --help`/`words` output updates automatically from the `NativeDoc` change.
5. **Validation & cleanup**
   - Run focused contract tests (`go test -run TestFunctionNoScope ./test/contract`) then the entire contract suite, followed by `go test ./...` if feasible.
   - Document the feature in `RELEASE_NOTES.md` if the repo tracks such changes.

## Integration Points
- **Evaluator/Frame coupling**: The shared-scope path still relies on `Evaluator.currentFrame()` and frame indices for lookups; changes must keep `frameStore` lifetimes intact to avoid dangling references.
- **Native registration/help**: `register_control.go` update ensures REPL `help fn` stays authoritative; missing this would confuse users.
- **Spec/docs**: Contracts drive downstream automated validation (`speckit`), so spec updates must precede code.
- **Agents**: Follow `AGENTS.md` by delegating code edits to the `viro-coder` agent and requesting a `viro-reviewer` pass before finalizing.

## Testing Strategy
- Contract tests covering:
  - **Mutation visibility**: `x: 1  bump: fn --no-scope [] [x: x + 1]  bump  x` ⇒ `2`.
  - **Parameter overwrites**: Outer `a` retains/changes according to the selected restoration policy; include assertions for both pre- and post-call values.
  - **Call-site variable access**: Define `adder` once, call it inside another function that declares locals, and assert the body sees them.
  - **Nested/no-scope combos**: `outer` defined normally, `inner` defined with `--no-scope` and invoked multiple times to ensure frames are not duplicated.
  - **Recursion or reentrancy**: If recursion is supported, test that successive calls do not leak stale parameter values (requires restoration). If unsupported, add a test ensuring recursion triggers a descriptive script error.
  - **Default path regression**: A normal `fn` definition behaves identically when `--no-scope` is absent.
- Optional integration tests: Add an example script under `examples/` or `test/scripts` demonstrating the behavior end-to-end.

## Potential Challenges & Open Questions
- **Parameter binding lifetime**: Decide whether formal parameters should persist in the caller frame post-call. Option B (restore) is safer for recursion and prevents silent leakage but slightly deviates from "reuse entirely"; plan recommends clarifying expectation with maintainers before implementation and encoding the decision in tests/spec.
- **Closure lookup order**: Without a new frame, lexical captures might become inaccessible; the shared-frame wrapper approach keeps `fn.Parent` reachable, but ensure lookups traverse both call-site scope and captured parent chain correctly.
- **Recursion/Concurrency**: Shared-scope recursion risks reusing the same words concurrently. If we restore parameter bindings per call, recursion remains feasible; otherwise, document recursion as unsupported for `--no-scope` functions.
- **Frame lifecycle**: Bypassing `PushFrameContext` must not break `MarkFrameCaptured` or tracing; double-check call-stack annotations and trace events still fire.
- **Error propagation**: Ensure `Return`, `break`, and script errors bubble the same way even when no frame is pushed.

## Viro Guidelines Reference
- Adhere to TDD: update specs and add contract tests before touching runtime code (`AGENTS.md`, "Workflow" section).
- Use `viro-coder` for all interpreter changes and request a `viro-reviewer` pass before completion.
- Maintain code style (no inline comments, constructor helpers like `value.IntVal`, table-driven tests in Go).
- Validate via targeted `go test` invocations before broader suites, per `AGENTS.md` build/test guidance.
- Keep documentation changes ASCII-only and ensure new plan numbering follows sequential convention (this file: `plans/040_fn_no_scope.md`).
