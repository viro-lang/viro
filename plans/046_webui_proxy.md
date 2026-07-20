# Plan 046: Remove internal WebUI manager and proxy calls directly to go-webui

## Feature Summary
- Drop the bespoke `internal/webui` manager (async render queues, handler bookkeeping, stub builds) and call `github.com/webui-dev/go-webui/v2` directly from the WebUI natives.
- Delete *all* WebUI-specific tests (contract and unit) plus the stub build/tag plumbing (`webui_stub`) so only the real CGO-backed integration remains.
- Re-implement every `webui.*` native as a thin synchronous proxy around the underlying go-webui call, letting the library determine blocking behavior and return values.
- Update documentation, specs, and examples so they describe the new proxy-style usage model (no manager, no polling loop) and explain that `webui.on`/`webui.poll` now map to go-webui bindings/`Wait()` semantics instead of Viro-managed event queues.

## Research Findings
- `internal/native/webui.go:13-382` currently expects an injected `webui.Manager` and performs significant validation/workflow orchestration (spec parsing, content-type enforcement, handler queues, explicit `poll`). This entire file must be rewritten once the manager disappears.
- `internal/webui/manager*.go` implement the asynchronous manager (real + stub) plus tests (`manager_real_async_test.go`). Removing them also eliminates the only reason for the `webui_stub` build tag.
- `internal/bootstrap/webui_registration*.go` guard native registration behind `//go:build webui || webui_stub`; this logic must change to a single `webui` build since headless stubs go away.
- Contract coverage for the existing behavior lives in `test/contract/webui_test.go:1-436`, which manipulates `webui.poll`, ready flags, and handler semantics that no longer apply.
- Public docs/specs (`docs/webui.md:1-143`, `docs/repl-usage.md:770-790`, `specs/002-implement-deferred-features/contracts/webui.md`) all describe the manager-driven queue semantics, poll loops, and stub tag usage that will become incorrect.
- Example `examples/10_webui.viro` demonstrates the polling loop + handler registration DSL that has to be simplified to show the blocking proxy usage.

## Architecture Overview
- Collapse the manager indirection: `internal/native/register_webui.go` will instantiate a lightweight `webuiRuntime` owned by the native package. The runtime simply tracks `nextID`, an index of `ui.Window` handles, and (optionally) currently bound callbacks so they can be released on close. No goroutines or queues remain.
- Each native (`webui.window`, `render`, `inject`, `send`, `close`, `ready?`, etc.) will perform basic argument validation and then call the corresponding go-webui method synchronously (e.g., `Window.Show`, `Window.Run`, `Window.Close`, `Window.IsShown`). Return values mirror what go-webui returns (mostly bool/error) without additional wrapping beyond converting into Viro values.
- `webui.on` is redefined as a thin shim over `window.Bind`: it binds the provided element/selector string to a Go callback that immediately executes the supplied handler block via the captured evaluator/frame. Because no queue exists, handlers run inline on whatever goroutine go-webui invokes; guard interpreter re-entry with a package-level mutex to serialize access and document that callbacks block the UI thread.
- `webui.poll` no longer drains an internal queue. Instead it becomes a simple proxy to `webui.Wait()` (blocking until all windows close) when called with `none`, or a no-op when a specific window handle is provided (since go-webui lacks per-window waits). Docs explain that scripts should call `webui.poll none` once to keep the process alive if they need event callbacks.
- Build tags simplify to `//go:build webui` on files that need CGO; there is no fallback stub. When `webui` tag is absent, the interpreter simply never binds `webui` natives (existing noop registration already handles this via `registerWebUINativesIfAvailable`).

## Implementation Roadmap
1. **Delete the manager layer and stub plumbing**
   - Remove `internal/webui/` entirely (manager structs, types, stub, tests) along with `viro_webui_test` helper binaries if they only existed for the manager.
   - Strip `webui_stub` references from the repo (`docs`, plans, specs, CI configs) and drop build tags that depended on it.

2. **Restructure native registration/state**
   - Replace `internal/native/register_webui.go` so it no longer accepts a `webui.Manager`. Instead, build a singleton runtime (e.g., `var webuiRT = newWebUIRuntime()`) that exposes methods invoked by the native wrappers.
   - Update `internal/bootstrap/webui_registration.go` to call `native.RegisterWebUINatives` without constructing a manager and guard it with `//go:build webui` only. The `_noop` variant remains for non-webui builds.

3. **Rewrite `internal/native/webui.go`**
   - Implement `webuiRuntime` with `sync.Mutex`, `nextID`, `map[uint32]ui.Window`, and optional `map[uint32][]handlerBinding` to keep track of `ui.Bind` closures for cleanup.
   - `webui.window`: parse the spec (still needed to translate Viro block into width/height/title/etc.), call `ui.NewWindow`, apply attributes via go-webui setters, store the window, and return a `webui-window!` whose payload is simply the assigned ID.
   - `webui.render`: evaluate markup argument into string/binary, pass it directly into `Window.Show` (or `ShowWebView` if desired). Do not enqueue jobs or manipulate ready flags beyond returning true/false based on Go error.
   - `webui.inject`, `webui.send`, `webui.close`, `webui.ready?` map straight to `Window.Run`, dispatch scripts, `Window.Close`, and `Window.IsShown`. Remove custom content-type enforcement and rely on go-webui errors instead.
   - `webui.on`: ensure selector parameter is string(s). For each selector, call `window.Bind(selector, func(event ui.Event) any { ... })` and run the handler block immediately. Use the evaluator from call time, capture the frame index, and synchronize with a mutex to avoid concurrent interpreter entry. Return logic true on success.
   - `webui.poll`: when called with `none`, run `ui.Wait()` so the process blocks until all open windows close; otherwise, return `none` immediately (documented as a no-op maintained for compatibility). This maps directly to go-webui's blocking wait call.
   - `webui.start` remains a no-op (possibly call `ui.SetConfig` for defaults) but becomes an alias for `value.NewNoneVal()`.

4. **Documentation/spec/example sweep**
   - Update `docs/webui.md`, `docs/repl-usage.md`, and the WebUI contract spec to explain the new synchronous proxy behavior, removal of manager concepts, and the new meaning of `webui.poll`/`webui.on` (callbacks execute immediately, `poll` just blocks on `ui.Wait()`). Remove all references to the stub build tag.
   - Rewrite `examples/10_webui.viro` to show the simplified flow: create window, render markup, call `webui.on` (if still supported) and end by calling `webui.poll none` once to keep the process alive instead of looping.
   - Add short release note or README blurb clarifying that WebUI now requires the `webui` build tag with CGO; there is no headless fallback and no automated tests.

5. **Remove WebUI tests and adjust CI expectations**
   - Delete `test/contract/webui_test.go` and any related helper scripts, as well as `internal/webui/manager_real_async_test.go` (no longer valid without a manager/stub).
   - Ensure CI or Makefile targets that previously ran `-tags webui_stub` are pruned so the default test suite no longer references those files.

6. **Manual verification guidance**
   - Since automated coverage disappears, document the manual validation steps developers must run on a GUI-capable host (e.g., `go build -tags webui ./cmd/viro`, run `examples/10_webui.viro`, click buttons to confirm callbacks execute, and observe that `webui.poll none` blocks until window close).

## Integration Points
- `internal/native/webui.go` (all native implementations) and `internal/native/register_webui.go` (object exposure) are the primary change sites.
- `internal/bootstrap/webui_registration.go` / `_noop` plus any package that previously relied on `webui.Manager` (should only be bootstrap + natives).
- Docs/specs/examples referencing poll loops (`docs/webui.md`, `docs/repl-usage.md`, `specs/002-implement-deferred-features/contracts/webui.md`, `examples/10_webui.viro`).
- Build/runtime instructions that mentioned `webui_stub` (README, docs, Makefile comments, CI workflows) must be synced.

## Testing Strategy
- No automated WebUI tests remain once the stub manager disappears. Instead, coders must:
  - Build `./cmd/viro` with `-tags webui` on a machine with CGO GUI dependencies installed.
  - Run `examples/10_webui.viro` and manually verify that `webui.render` brings up a window, `webui.on` callbacks execute immediately, and `webui.poll none` keeps the interpreter alive until the window closes.
  - Run the standard non-WebUI `go test ./...` to ensure removing the stub/test files doesnt break unrelated packages.
- Document these manual steps so reviewers know how to validate the feature in lieu of contract/unit tests.

## Potential Challenges
- **Evaluator re-entry:** Running handler blocks directly from go-webui callbacks risks concurrent interpreter use. We must guard callbacks with a mutex or channel so only one handler executes at a time and document that handlers block the UI.
- **Loss of automated coverage:** Removing the stub/tests conflicts with the usual TDD requirement. Justify this change (per the user request) in commit messages and emphasize manual validation steps.
- **Blocking semantics changes:** Without the manager, `webui.render` may block until the browser loads, and `webui.poll` now effectively replaces the previous main loop. Examples/docs must be clear about this behavior to avoid confusion.
- **CGO-only builds:** Eliminating `webui_stub` means WebUI is entirely unavailable unless the `webui` tag is set and CGO deps exist. Ensure bootstrap errors are readable when users call `webui.*` without building with `-tags webui` (the object simply wont exist).

## Viro Guidelines Reference
- Follow `AGENTS.md`: all Go edits must go through the `viro-coder` agent, and finished patches require a pass from `viro-reviewer` even though this plan targets documentation/removal work.
- Maintain Viro coding conventions (no Go comments in code, table-driven tests elsewhere, `value.New*` constructors) when rewriting the natives.
- Recognize that dropping WebUI tests deviates from the usual TDD first rule; explicitly note in PR/commit messaging that the test removal is intentional per requirements so reviewers can sign off.
