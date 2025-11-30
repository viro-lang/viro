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

Creates or reuses a window.

**Signature:** `webui.window spec-block`

**Parameters:**
- `spec-block`: Block accepting keys:
  - `title text!`: Window title (not supported by go-webui library)
  - `width integer!`: Window width (not supported by go-webui library)
  - `height integer!`: Window height (not supported by go-webui library)
  - `resizable? logic!`: Allow resizing (not supported by go-webui library)
  - `icon file!`: Window icon (not supported by go-webui library)
  - `source file!|url!`: Initial content source (not supported)
  - `html string!|binary!`: Initial HTML content

**Returns:** `webui-window!` handle

**Notes:** Most spec keys are not supported by the underlying go-webui library and are ignored. Only `html` is currently supported for initial content.

### webui.render

Replaces current DOM with provided markup.

**Signature:** `webui.render window markup options?`

**Parameters:**
- `window`: `webui-window!` handle
- `markup`: `string!`, `binary!`, or `file!`
- `options?`: Optional block with keys:
  - `content-type word!|string!`: One of `text/html`, `text/plain`, `application/xhtml+xml`, `application/json`, `text/javascript`

**Returns:** `logic!` success

**Notes:** Resets page and replays handlers. Binary markup requires `content-type`. Default content-type is `text/html` for strings. Invalid content-type raises spec error. Other options are not currently supported.

### webui.inject

Evaluates markup snippets in existing page context.

**Signature:** `webui.inject window html`

**Parameters:**
- `window`: `webui-window!` handle
- `html`: `string!` or `binary!`

**Returns:** `logic!` success

**Notes:** Does not reset handlers. For incremental updates.

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
- `selector`: `text!` or `block!` of strings (CSS selector)
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

- Invalid spec blocks: `verror.NewScriptError` with category "webui" and ID "invalid-spec"
- Binary without content-type: "content-type-required"
- Invalid content-type: "invalid-content-type"
- Serialization failures: "serialization-error"
- Window not found: "window-not-found"
- Handler execution errors: Propagate to `webui.poll` caller

## CGO Requirements

- Requires CGO enabled (`CGO_ENABLED=1`)
- Platform dependencies:
  - Linux: WebKitGTK
  - Windows: Edge WebView2
  - macOS: WebKit
- Build with `webui` tag to enable functionality

## Implementation Notes

- Direct go-webui integration without manager abstraction
- Synchronous event handling via go-webui callbacks
- `webui.poll none` blocks until all windows close (equivalent to `ui.Wait()`)
- `webui.poll window` is a no-op (per-window wait not available)
- Window handles are numeric IDs
- Event handlers execute immediately when events occur
- Lifecycle: Windows closed on interpreter exit