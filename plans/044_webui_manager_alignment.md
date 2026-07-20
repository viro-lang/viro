# Plan 044: Align WebUI Manager With go-webui/v2

## Feature Summary
- Replace the guessed CGO integration in `internal/webui/manager_real.go` with a concrete implementation that targets the real `github.com/webui-dev/go-webui/v2` API.
- Keep the stub (`webui_stub`) and noop builds compiling by mirroring the updated structs and behaviors even when CGO is unavailable.
- Update documentation/specs to reflect capabilities that the actual go-webui runtime supports vs. what was previously assumed.

## Research Findings
- `internal/webui/manager_real.go` currently fails to compile: it stores `window interface{}` yet calls methods directly, references non-imported `value` helpers, and invokes non-existent go-webui methods such as `SetTitle` (file lines 18-138).
- Actual go-webui API surface lives in `dependencies/go-webui/v2/lib.go`. Relevant call points:
  - Window lifecycle: `ui.NewWindow`, `Window.Show`, `Window.ShowBrowser`, `Window.ShowWebView`, `Window.SetSize`, `Window.SetResizable`, `Window.SetIcon`, `Window.SetRootFolder`, `Window.Run`, `Window.Close`, `webui.Wait` (lib.go lines ~181-999).
  - Eventing: `webui.Bind`/`Window.Bind` returning `ui.Event` callbacks with `EventType` enums (Disconnected, Connected, Callback) and helpers like `GetArg`, `GetArgAt`.
  - No direct title setter exists; need to drive `<title>` via DOM script.
- Existing Viro manager contract (`internal/native/webui.go`, `specs/002-implement-deferred-features/contracts/webui.md`) expects features like `title`, `resizable?`, `icon`, `source`, `html`, ready tracking, and event selectors, so the real manager must synthesize these behaviors on top of the go-webui primitives.
- Tests (`test/contract/webui_test.go`) currently exercise stub behaviors (window creation, ready flag toggling, event registration success) and must remain green when running with `-tags webui_stub`.

## Architecture Overview
- Keep the current `Manager` interface but implement a concrete `manager` struct that stores `ui.Window` handles and metadata (ready flag, registered handlers, JS bridge state).
- Introduce a WebUI bridge helper that injects JavaScript into the DOM to:
  - Attach DOM event listeners by selector/event type that call back into Go through dynamically bound `webui.<name>` functions.
  - Signal "ready" back to Go once the DOM and `webui.js` bootstrap have finished loading.
- Maintain a per-window handler registry so that `RegisterEvent` creates both the backend callback (`webui.Bind`) and the front-end listener snippet (`window.Run`).
- `manager_stub` should share the same data layout (handler bookkeeping, ready flag toggling) but operate purely in-memory for tests.

## Implementation Roadmap
1. **Inspect and map the go-webui/v2 API**
   - Read `dependencies/go-webui/v2/lib.go` to catalog the available window methods (`Show`, `ShowWebView`, `SetSize`, `SetResizable`, `SetIcon`, `SetRootFolder`, `Navigate`, `Run`, `Script`, etc.) and cross-check README/examples for usage patterns (e.g., `examples/call_go_from_js.go`, `examples/frameless.go`).
   - Confirm which desired spec knobs are supported directly vs. require shims: e.g., `resizable?` -> `SetResizable`, `width`/`height` -> `SetSize`, `icon` -> `SetIcon` plus MIME detection, `title` -> DOM script fallback, `debug?` -> (likely unsupported) document limitation.
   - Validate event handling expectations by inspecting `webui.js` contract (examples show `webui.<bindingName>(args)`), so selectors/events must be implemented with injected scripts that call the bound function names.

2. **Update `manager_real.go` to the real API**
   - Change `windowState.window` to `ui.Window` and add fields for `bridgeLoaded bool`, `handlerBindings map[string]handlerBinding` (stores selector/event/function name), and `readyBinder string` for the special ready callback.
   - `CreateWindow`:
     - Instantiate `ui.NewWindow`, assign ID, configure width/height/resizable/icon/root folder as supported; warn or ignore unsupported spec keys (title/debug) but document behavior.
     - Pre-bind a lifecycle callback: `ui.Bind(window, viroLifecycleName, ...)` to listen for JS ready pings and to mark the `ready` flag.
     - Store state and return ID.
   - `Render`:
     - Accept `string`, `binary`, or `core.File?` values; convert to string content and call either `window.Show` (embedded markup) or `window.ShowWebView` if we detect a file path requiring WebView (document selection logic in helper).
     - Reset `state.ready=false` before showing. After `Show`, inject the JS bridge (if not already) plus any previously registered handler wiring via `window.Run` scripts. Bridge should define functions like `window._viroBind(event, selector, fnName)` that attach DOM listeners and call `webui[fnName](JSON.stringify(detail))`.
   - `Inject` and `Send` should call `window.Run` with sanitized JS strings; reuse `jsQuote` but ensure we also escape template literal/backticks.
   - `RegisterEvent`:
     - Validate window state, allocate a unique backend binding name (e.g., `viro_evt_{id}_{counter}`), call `ui.Bind(window, bindingName, callback)`.
     - Within callback, decode incoming payload JSON (use `value.FromJSON`) and enqueue `EventMessage` just like today.
     - Immediately queue a JS snippet that asks the bridge to attach DOM listeners for the provided selector/event combination so that `webui[bindingName](serializedEvent)` is invoked client-side. Handle multiple selectors by generating multiple DOM listeners.
   - `Ready` flag should flip true after receiving the JS "ready" callback. Provide a fallback that marks ready when we get a `ui.Event` of type `ui.Connected` on our lifecycle binding to avoid deadlock if the injected script fails.
   - Close/CloseAll just call `Window.Close` and keep done channel semantics (already implemented but ensure `state.window` isn't nil).

3. **Keep stub/noop managers and tests aligned**
   - Mirror any new fields added to `windowState`/`manager` so the interfaces stay in sync across build tags.
   - Enhance the stub to simulate behaviors the real manager now exposes: toggling `ready` after `Render`, tracking handler metadata, and optionally enqueueing fake events when tests call helper hooks (if needed). Ensure the stub compiles without referencing `ui.Window` by gating fields/build tags or using an interface type alias defined in stub builds.
   - Update `test/contract/webui_test.go` only if newly implemented semantics require coverage adjustments (e.g., verifying that `webui.on` rejects selectors before render, ensuring `webui.send` errors propagate). Tests should continue to run against the stub via `-tags webui_stub` to avoid GUI dependencies.

4. **Documentation and spec updates**
   - Update `docs/webui.md` and `specs/002-implement-deferred-features/contracts/webui.md` to clarify which spec keys are supported: document that `title`/`debug?` are best-effort (implemented by DOM script), `icon` expects a path plus optional inferred MIME, and `ready?` relies on the injected bridge handshake.
   - Describe the bridge-driven selector/event system so users know they must include `webui.js` and that events are dispatched via JS listeners, not native WebUI element bindings.
   - Mention any unsupported keys or future work so expectations are clear.

5. **Verification steps**
   - Build with CGO: `CGO_ENABLED=1 go build -tags webui ./cmd/viro` (or `make build WEBUI=1` if a helper target exists) to ensure the real manager compiles and links against go-webui.
   - Run the existing contract tests with the stub: `go test -tags webui_stub ./test/contract -run WebUI` plus any focused packages touched in `internal/webui` and `internal/native`.
   - Optionally run a manual smoke example (e.g., minimal WebUI script) to confirm real windows open when CGO is available.

## Integration Points
- `internal/native/webui.go` (already parsing specs) will start exercising real behaviors via the manager; ensure any new errors map to `verror` IDs.
- `internal/value` JSON helpers are used when sending/receiving payloads; confirm imports and error paths are maintained.
- Build tags: `manager_real.go` (`//go:build webui && !webui_stub`), `manager_stub.go` (`//go:build webui_stub`), `manager_noop.go` (`//go:build !webui && !webui_stub`). Maintain identical method sets.

## Testing Strategy
- Unit/contract tests continue to use the stub so CI remains headless.
- Consider adding a small integration test guarded by `webui_stub` to assert handler registration bookkeeping (e.g., `manager_stub` pushes a synthetic event into the queue and `webui.poll` processes it).
- Manual/optional: run an example script under `CGO_ENABLED=1` to ensure the DOM bridge correctly routes events.

## Potential Challenges
- No native `SetTitle` or `debug` toggles exist; need to document the workaround and ensure JS injection handles cases where HTML already defines `<title>`.
- Injected scripts must be idempotent across multiple renders/injects; track whether the bridge helper has been injected and re-bind selectors after DOM refresh.
- Event binding is asynchronous; must ensure we re-register listeners after each `Render` (DOM replaced) and treat `webui.inject` as incremental updates that may target already bound nodes.
- Need to guard against concurrent access to `windowState` (use existing mutex) when binding/unbinding handlers and when ready callbacks arrive on background goroutines.

## Viro Guidelines Reference
- Follow `AGENTS.md`: use `viro-coder` for all interpreter code edits, run tests first (TDD) in `test/contract` before touching `internal/native` or `internal/webui`.
- Maintain spec compliance from `specs/002-implement-deferred-features/contracts/webui.md` and update docs when semantics change.
- Preserve Go style rules (no inline comments, constructor helpers, error handling via `verror`).
