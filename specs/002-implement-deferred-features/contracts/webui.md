# WebUI Contract

## Overview

The WebUI integration provides Viro scripts with the ability to create and control native WebUI windows (HTML/CSS/JS surfaces) while maintaining the interpreter's CLI process. This enables GUI applications written in Viro without requiring external processes or new CLI modes.

## API Surface

All WebUI functionality is exposed through dot-notation methods on the shared `webui` object.

### webui.start

Initializes the WebUI subsystem (idempotent).

**Signature:** `webui.start`

**Returns:** `none!`

**Notes:** Idempotent initializer that other webui functions auto-call.

### webui.window

Creates a new window.

**Signature:** `webui.window`

**Parameters:** None

**Returns:** `webui-window!` handle

**Notes:** Creates a new WebUI window. Window properties like title, size, etc. are not supported by the underlying go-webui library.

### webui.render

Replaces current DOM with provided markup.

**Signature:** `webui.render window markup`

**Parameters:**
- `window`: `webui-window!` handle
- `markup`: `string!` or `binary!`

**Returns:** `logic!` success

**Notes:** Replaces the current page content with the provided HTML markup.

### webui.inject

Executes JavaScript in the window context.

**Signature:** `webui.inject window script`

**Parameters:**
- `window`: `webui-window!` handle
- `script`: `string!` or `binary!`

**Returns:** `logic!` success

**Notes:** Executes JavaScript code in the browser context.

### webui.send

Pushes data to JS side.

**Signature:** `webui.send window message payload`

**Parameters:**
- `window`: `webui-window!` handle
- `message`: `text!` channel name
- `payload`: `none`, `logic!`, `integer!`, `decimal!`, `string!`, `binary!`, `map!`, `block!`, `object!`

**Returns:** `logic!` success

**Notes:** Non-string types serialized to JSON. Binary as base64.

### webui.on

Registers event callback.

**Signature:** `webui.on window event selector handler-block`

**Parameters:**
- `window`: `webui-window!` handle
- `event`: `text!` event name
- `selector`: `text!` CSS selector
- `handler-block`: `block!` to execute

**Returns:** `logic!` success

**Notes:** Captures current frame for lexical bindings. Handler receives `event-name`, `event-selector`, `event-payload`, `event-window`.

### webui.poll

Drains pending events and runs handlers.

**Signature:** `webui.poll window`

**Parameters:**
- `window`: `webui-window!` or `none` (all windows)

**Returns:** `none`

**Notes:** When called with `none`, blocks until all windows close (equivalent to `ui.Wait()`). When called with a specific window, no-op (per-window wait not available in go-webui). Handler errors are swallowed and do not propagate.

### webui.close

Closes window and removes registrations.

**Signature:** `webui.close window`

**Parameters:**
- `window`: `webui-window!` handle

**Returns:** `logic!` whether window was closed

**Notes:** When last window closes, signals `Done()`.

### webui.ready?

Checks if window is ready for rendering.

**Signature:** `webui.ready? window`

**Parameters:**
- `window`: `webui-window!` handle

**Returns:** `logic!`

**Notes:** True after first `Ready` callback. Resets to false on new `webui.render` or close.

## Event Semantics

Events are processed synchronously via go-webui callbacks. Handlers execute immediately when events occur with auto-bound locals:
- `event-name`: Event name (string!)
- `event-selector`: CSS selector (string!)
- `event-payload`: Event payload (string!)
- `event-window`: Window handle (webui-window!)

## Hot Reload Behavior

- Use `webui.render` with `source` option for full page reloads
- Use `webui.inject` for incremental updates
- Manual watchers can re-invoke render on file changes
- Future: `webui.refresh` for hook registration

## Error Handling

- Type mismatches: `verror.NewScriptError` with category "webui" and appropriate ID
- Serialization failures: "send-failed"
- Handler execution errors: Propagate to event callback

## CGO Requirements

- Requires CGO enabled (`CGO_ENABLED=1`)
- Platform dependencies:
  - Linux: WebKitGTK
  - Windows: Edge WebView2
  - macOS: WebKit
- Build with `webui` tag to enable functionality

## Implementation Notes

- Thin wrappers around go-webui functions
- No runtime bookkeeping or registries
- Window handles store ui.Window directly
- Synchronous event handling via go-webui callbacks
- `webui.poll none` blocks until all windows close (equivalent to `ui.Wait()`)
- `webui.poll window` is a no-op (per-window wait not available)
- Event handlers execute immediately when events occur
- Lifecycle: Windows closed on interpreter exit