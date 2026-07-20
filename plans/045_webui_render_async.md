# Plan 045: WebUI render must be non-blocking

## Feature Summary
- Ensure `webui.render` kicks off a WebUI window refresh without stalling the evaluator so scripts can immediately enter their event loops (e.g., `while [webui.poll ...]`).
- Diagnose why the real manager currently returns `render-failed`/"window.render function blocks" when the user script from `test.viro` runs and fix the runtime so `webui.render` returns once the new DOM load is scheduled, not after the WebUI event loop completes.
- Keep `webui.poll` responsible only for draining queued events; users should not need to poll to unblock `render`.

## Research Findings
- `internal/native/webui.go:125-184` simply forwards `webui.render` to `Manager.Render` and surfaces errors as `render-failed`, so the blocking originates inside the manager implementation.
- `internal/webui/manager_real.go:27-118` calls `ui.NewWindow().Show(content)` synchronously and only then injects the JS bridge. If `Show` blocks (e.g., because go-webui expects its own event loop to pump before returning), any failure is reported as `render-failed` to the caller.
- `internal/webui/manager_real.go:314-360` injects a bridge that toggles `state.ready` by invoking the bound `__viro_ready` handler; however, the guard is never reached if `Show` never returns.
- `docs/webui.md:62-86` and `specs/002-implement-deferred-features/contracts/webui.md:42-118` describe `webui.render` as a DOM-resetting operation that should return promptly so users can call `webui.poll` in a loop, matching the expectation from `test.viro`.
- `examples/10_webui.viro` loops forever and polls the window, implying `webui.render` is expected to finish immediately, otherwise the examples would hang before reaching the polling loop.
- go-webui’s native usage (see `dependencies/go-webui/v2/webui/README.md:168-195`) shows `webui_show()` followed by a separate `webui_wait()` call, suggesting we can run `window.Show` in a dedicated worker goroutine while keeping the interpreter responsive.

## Architecture Overview
- Introduce an internal WebUI render worker per manager (or per window) that owns the actual go-webui calls. `Manager.Render` enqueues render jobs and returns as soon as they are scheduled, while the worker drives `window.Show`, bridge injection, and ready flag updates.
- Maintain a background goroutine started on the first real window that calls `ui.Wait()` (or equivalent) so the go-webui runtime keeps processing OS/UI events independently of the evaluator thread. Closing all windows should call `ui.Exit()` (or `window.Close`) and unblock the waiter.
- Guard state transitions with per-window mutexes: `Render` should reset `state.ready`, record the pending markup/options, and signal the worker to refresh the window without holding the global lock during the actual `Show` call.
- Keep the existing `Manager` interface untouched so `internal/native/webui.go` does not change, but extend the hidden `manager` struct (shared by real and stub builds) with queues/channels powering asynchronous operations.

## Implementation Roadmap
1. **Reproduce & guard with tests**
   - Extend `test/contract/webui_test.go` with a script-level test that emulates `test.viro` (a render followed by a polling loop) under the stub manager to codify the expectation that `webui.render` returns truthy immediately and `webui.poll` can be called afterwards without throwing.
   - Add Go unit tests under `internal/webui` (likely `manager_real_async_test.go` gated by `//go:build webui_stub` to reuse a fake window) that verify `Manager.Render` enqueues quickly and does not block the caller even when a fake window’s `Show` is artificially delayed.
2. **Instrument the manager for diagnostics**
   - Augment `windowState` with fields like `renderQueue chan renderRequest`, `rendering bool`, and timestamps so we can detect blocked renders and log/return clearer errors.
   - Add tracing hooks guarded by `--verbose` so developers can see when renders start/finish.
3. **Introduce asynchronous render execution**
   - Refactor `manager.Render` (real build) to push a `renderRequest{windowID, markup, options}` onto a per-window queue and immediately return. The worker goroutine drains that queue, calls `window.Show`, injects the bridge, and updates `state.ready`/`bridgeLoaded` once the injected JS posts the `__viro_ready` signal.
   - Ensure errors from `window.Show` are captured and propagated: the worker should send the failure back over a channel so synchronous callers can still learn about setup errors (e.g., invalid HTML) without waiting for the entire render lifecycle.
4. **Start/stop the go-webui loop explicitly**
   - When the first real window is created, spawn a goroutine calling `ui.Wait()` so the underlying C library runs as designed; when all windows close (or the interpreter shuts down), call `ui.Exit()` and close the goroutine, mirroring the `CloseAll` logic in `manager_stub.go`.
   - Tie `manager.Done()` to the go-webui loop so `webui.poll` can observe shutdown without racing.
5. **Decouple ready state from polling**
   - Ensure `state.ready` flips to true solely when the injected bridge calls `__viro_ready`; it must not depend on `webui.poll`. Add a timeout or watchdog that reports a more actionable error if ready is never signaled within a reasonable duration.
6. **Documentation & examples**
   - Update `docs/webui.md` to clarify that `webui.render` is asynchronous but `webui.poll` must be called in a loop (preferably with `wait`/`sleep`) to keep the app responsive, and mention the background `ui.Wait()` worker requirement.
   - Refresh `examples/10_webui.viro` (and `test.viro` if necessary) to include a small delay (`wait 0.05`) between polls so users avoid busy loops, matching the new behavior.
7. **Regression sweep**
   - Run `go test -tags webui_stub ./...` to ensure headless builds still pass, plus manual `./viro test.viro` on a GUI-enabled machine to confirm the window stays up and logs no blocking errors.

## Integration Points
- `internal/webui/manager_real.go` (asynchronous render pipeline, go-webui loop management) and its shared types in `internal/webui/manager.go`.
- `internal/native/webui.go` (still the only consumer of `Manager.Render`, but add tests to guarantee the API contract if new fields appear).
- `internal/bootstrap/webui_registration.go` (ensures manager is instantiated once and closes all windows on interpreter exit).
- `test/contract/webui_test.go` plus any helper scripts under `examples/` or `test/scripts` showcasing the expected looping pattern.

## Testing Strategy
- **Contract tests**: add table-driven cases ensuring `webui.render` returns `logic! true` even when followed immediately by `webui.poll`, and that repeated renders reset readiness without throwing.
- **Manager unit tests**: craft fake window implementations (with controllable `Show` delays and injected errors) and run them under `go test ./internal/webui -tags webui_stub` to confirm the asynchronous job queue drains correctly, ready flags flip as expected, and error paths propagate to callers.
- **Manual GUI test**: run `./viro test.viro` (and `examples/10_webui.viro`) on a machine with the `webui` tag build to verify the window stays open while the poll loop runs and no `render-failed` error appears.
- **Race detection**: run `go test -race ./internal/webui` under the stub build to ensure the new goroutines and channels are synchronized properly.

## Potential Challenges
- **CGO main-thread requirements**: go-webui may require certain calls (e.g., `ui.Wait`) to run on the main OS thread. We must verify whether starting the worker in `init()` or `main` is necessary and document any constraints in `cmd/viro`.
- **Error propagation timing**: once renders are queued asynchronously, we need a deterministic way to surface immediate failures without blocking forever. Use bounded channels/timeouts so `webui.render` either reports the failure promptly or returns success.
- **Resource cleanup**: goroutines associated with closed windows must exit to avoid leaks. Ensure queues are drained and `CloseAll` shuts down workers before the interpreter exits.
- **Ready handshake reliability**: if the injected bridge fails to execute (e.g., invalid HTML missing `<head>`), the ready signal may never fire. Implement timeouts and helpful error messages rather than hanging indefinitely.

## Viro Guidelines Reference
- Follow `AGENTS.md`: add/adjust tests first (`test/contract` + Go unit tests), run all edits through the `viro-coder` agent, and request a mandatory pass from `viro-reviewer` afterward.
- Maintain spec alignment with `specs/002-implement-deferred-features/contracts/webui.md`, ensure no inline Go comments, and keep table-driven tests for any new Go unit suites.
- Run `go test -tags webui_stub ./...` plus targeted packages before submission, and document GUI prerequisites in `docs/webui.md` per the repo’s documentation practices.
