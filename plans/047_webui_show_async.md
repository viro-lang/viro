# 047 - Stabilize `webui-show` Semantics

## Feature Summary
- `webui-show` (and `webui-show-browser`) should return immediately with `true` when a window is shown and stay open until `webui-wait` runs. Right now the window flashes, the native returns `false`, and go-webui logs a WebKit signal error because it waits for a JavaScript handshake that never arrives.
- Provide a predictable experience for Viro scripts that want to stage HTML and call `webui-wait` later without including the `webui.js` bridge script when they do not need callbacks.

## Research Findings
- The current native implementation is a thin wrapper that calls `ui.Window.Show` and converts any error into the `false` logic value (`internal/native/webui.go:41-90`). There is no retry, no config adjustment, and no logging that would explain *why* `Show` failed.
- Documentation promises that `webui-show` simply displays content and that `webui-wait` blocks later (`docs/webui.md:48-85`), so the current behavior contradicts published semantics.
- go-webui defaults to `Config.ShowWaitConnection = true`, meaning `webui_show` blocks until a client fully connects; if no `webui.js` script runs, the call returns `false` and prints a fatal signal log (`dependencies/go-webui/v2/lib.go:94-121,362-381`). We recently exposed the direct API without also switching that config, which explains the reported failure when the user supplies a bare HTML string.

## Architecture Overview
- Introduce a thin "runtime" wrapper in `internal/native` that centralizes go-webui calls. The wrapper will:
  - Set `ui.SetConfig(ui.ShowWaitConnection, false)` once at startup (lazy-init guard) so `Show` no longer waits for a bridge connection before returning.
  - Optionally expose a `Show` helper that can still opt into waiting when the caller explicitly asks (e.g., via a new refinement or by reusing `webui-set-config`).
  - Capture errors from go-webui and map them to descriptive Viro script errors or logs so users can diagnose platform issues.
- Keep `webui-show`, `webui-show-browser`, and any future display-related natives routed through that wrapper, ensuring consistent non-blocking behavior and simplifying future adjustments (timeouts, logging, goroutine management).

## Implementation Roadmap
1. **Reproduce & Log**
   - Build with `-tags webui` and run a minimal script (`webui-new-window`, `webui-show`, `webui-wait`) to confirm the `false` return and collect the exact log emitted by go-webui. This serves as the baseline behavior to verify the fix against.
2. **Instrument Runtime Initialization**
   - Add a `webuiRuntime` helper (new file under `internal/native`, same build tag) that exposes `initOnce sync.Once`, `ensureInitialized()`, and wrapper methods for `Show`, `ShowBrowser`, etc.
   - Inside `ensureInitialized`, call `ui.SetConfig(ui.ShowWaitConnection, false)` and optionally hook `ui.SetLogger` to route fatal messages through `verror` or Go logging so tests can assert on them without GUI dependencies.
3. **Update Natives**
   - Change `webui-show`/`webui-show-browser` to call the wrapper instead of `ui.Window` directly. The wrapper should:
     - Ensure initialization.
     - Call the underlying go-webui function.
     - Convert errors into `verror.NewScriptError` (category `verror.CategoryIO`?) or, if we want to keep the boolean, at least include detailed log entries so users know the failure reason.
     - Return `true` when `Show` succeeds immediately, even though the UI might still be loading.
4. **Optional Timeout Controls**
   - Surface a way to re-enable waiting for specific windows (maybe a refinement like `webui-show/await` or instruct users to call `webui-set-config 0 true`). The plan should document this decision so coder agent either adds a keyword or relies entirely on `webui-set-config` while updating docs to explain the new default.
5. **Testing**
   - Add build-tagged Go tests under `internal/native` (e.g., `webui_runtime_test.go` with `//go:build webui`) that stub go-webui via an interface, or
   - Abstract go-webui behind an interface that can be swapped with a fake implementation at test time (place fake inside `_test` file). The test should assert that `ensureInitialized` toggles `ShowWaitConnection` exactly once and that `WebUIShow` returns `true` when the fake `Show` succeeds and propagates `false` otherwise.
   - Add a contract or integration test script under `test/integration` guarded by `t.Skip` unless `VIRO_TEST_ENABLE_WEBUI=1`, verifying the end-to-end scenario on developer machines.
6. **Docs & Release Notes**
   - Update `docs/webui.md` to explain the new default (no wait for connection) and that users must include `webui.js` or call `webui-set-config ShowWaitConnection true` if they rely on immediate bridge availability.
   - Add an entry to `RELEASE_NOTES.md` describing the behavioral fix for `webui-show`.

## Integration Points
- `internal/native/webui.go`: replace direct go-webui calls for show-related natives with the new runtime helper.
- `internal/native` (new helper file) with `//go:build webui` plus a `*_noop` version under `//go:build !webui` to keep headless builds compiling.
- Optional updates to `internal/bootstrap/webui_registration*.go` if init ordering matters.
- Documentation in `docs/webui.md` and possibly `examples/10_webui.viro` to demonstrate the corrected flow.

## Testing Strategy
- **Unit Tests (build-tagged)**: Use a fake implementation to assert initialization semantics and boolean conversions without spawning real windows.
- **Integration Test**: Small script executed via `go test ./test/integration -run TestWebUIShow` that is skipped unless GUI deps are present but can be run manually before release.
- **Manual Verification**: Follow reproduction script before/after to ensure `webui-show` now returns `true` immediately and the window stays alive until `webui-wait`.

## Potential Challenges
- Ensuring tests compile when the `webui` build tag is off; every helper must have a corresponding noop stub.
- Managing global config: disabling `ShowWaitConnection` affects every caller, so we must document and possibly expose overrides for users who *do* want to wait.
- go-webui logging is global; routing logs without spamming users requires careful integration with Viro's logging or trace facilities.
- CI environments may lack GUI libraries, so any non-skipped integration test must either run behind the `webui` tag or use fakes exclusively.

## Viro Guidelines Reference
- Follow the native implementation rules outlined in `docs/viro_architecture_knowledge.md` for value constructors and naming.
- Honor the no-comments-in-code policy and keep logic inside dedicated helpers instead of large inline blocks (see AGENTS.md code style section).
- All behavior changes require tests in `test/contract` or `test/integration` before touching the interpreter (TDD mandate).
- Update documentation whenever public behavior changes (`docs/webui.md`).
