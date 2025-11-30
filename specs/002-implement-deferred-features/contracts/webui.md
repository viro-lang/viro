# WebUI Contract

## Overview

The WebUI integration provides Viro scripts with direct access to the WebUI C API functions for creating and controlling native WebUI windows (HTML/CSS/JS surfaces) while maintaining the interpreter's CLI process. This enables GUI applications written in Viro without requiring external processes or new CLI modes.

## API Surface

All WebUI functionality is exposed as direct native functions with kebab-case naming, corresponding directly to the WebUI C API.

### Window Management

#### webui-new-window

Creates a new WebUI window.

**Signature:** `webui-new-window`

**Parameters:** None

**Returns:** `integer!` Window ID

**Notes:** Creates a new WebUI window and returns its window ID as an integer.

#### webui-new-window-id

Creates a new WebUI window with a specific ID.

**Signature:** `webui-new-window-id window-id`

**Parameters:**
- `window-id`: `integer!` Window ID to use

**Returns:** `integer!` Window ID (same as input)

**Notes:** Creates a new WebUI window with the specified window ID.

### Window Display

#### webui-show

Shows a window with HTML content.

**Signature:** `webui-show window-id content`

**Parameters:**
- `window-id`: `integer!` Window ID
- `content`: `string!` HTML content

**Returns:** `logic!` Success status

**Notes:** Shows the WebUI window with the specified HTML content. Returns immediately (non-blocking) after displaying the window, even if no JavaScript bridge is present. The window stays open until `webui-wait` or `webui-close` is called.

#### webui-show-browser

Shows a window in a specific browser with HTML content.

**Signature:** `webui-show-browser window-id content browser`

**Parameters:**
- `window-id`: `integer!` Window ID
- `content`: `string!` HTML content
- `browser`: `integer!` Browser ID

**Returns:** `logic!` Success status

**Notes:** Shows the WebUI window in a specific browser with HTML content. Returns immediately (non-blocking) after displaying the window, even if no JavaScript bridge is present. The window stays open until `webui-wait` or `webui-close` is called.

### Event Handling

#### webui-bind

Binds a Viro block as an event handler for an element.

**Signature:** `webui-bind window-id element handler`

**Parameters:**
- `window-id`: `integer!` Window ID
- `element`: `string!` Element name/ID
- `handler`: `block!` Handler block

**Returns:** `logic!` Success status

**Notes:** Binds a Viro block as an event handler. The handler block receives `event-element`, `event-data`, `event-window-id`, and `event-number` bindings.

### Script Execution

#### webui-run

Executes JavaScript in a window.

**Signature:** `webui-run window-id script`

**Parameters:**
- `window-id`: `integer!` Window ID
- `script`: `string!` JavaScript code

**Returns:** `none!` None

**Notes:** Executes JavaScript code in the specified window.

### Window Lifecycle

#### webui-close

Closes a window.

**Signature:** `webui-close window-id`

**Parameters:**
- `window-id`: `integer!` Window ID

**Returns:** `none!` None

**Notes:** Closes the specified WebUI window.

#### webui-destroy

Destroys a window and frees resources.

**Signature:** `webui-destroy window-id`

**Parameters:**
- `window-id`: `integer!` Window ID

**Returns:** `none!` None

**Notes:** Destroys the specified WebUI window and frees its resources.

### Window State

#### webui-is-shown

Checks if a window is shown.

**Signature:** `webui-is-shown window-id`

**Parameters:**
- `window-id`: `integer!` Window ID

**Returns:** `logic!` Whether window is shown

**Notes:** Returns whether the specified window is currently shown.

### Configuration

#### webui-set-timeout

Sets the global timeout for WebUI operations.

**Signature:** `webui-set-timeout timeout`

**Parameters:**
- `timeout`: `integer!` Timeout in seconds

**Returns:** `none!` None

**Notes:** Sets the global timeout for WebUI operations in seconds.

#### webui-wait

Waits for all windows to close (blocking call).

**Signature:** `webui-wait`

**Parameters:** None

**Returns:** `none!` None

**Notes:** Blocks until all WebUI windows are closed. This is the main event loop for WebUI applications.

#### webui-exit

Exits the WebUI subsystem.

**Signature:** `webui-exit`

**Parameters:** None

**Returns:** `none!` None

**Notes:** Exits the WebUI subsystem.

#### webui-set-config

Sets a WebUI configuration option.

**Signature:** `webui-set-config option value`

**Parameters:**
- `option`: `integer!` Config option ID
- `value`: `logic!` Option value

**Returns:** `none!` None

**Notes:** Sets a WebUI configuration option.

### System Information

#### webui-get-parent-process-id

Gets the parent process ID for a window.

**Signature:** `webui-get-parent-process-id window-id`

**Parameters:**
- `window-id`: `integer!` Window ID

**Returns:** `integer!` Parent process ID

**Notes:** Returns the parent process ID for the specified window.

### File System

#### webui-set-root-folder

Sets the root folder for serving files in a window.

**Signature:** `webui-set-root-folder window-id path`

**Parameters:**
- `window-id`: `integer!` Window ID
- `path`: `string!` Root folder path

**Returns:** `logic!` Success status

**Notes:** Sets the root folder for serving files in the specified window.

#### webui-set-default-root-folder

Sets the default root folder for serving files.

**Signature:** `webui-set-default-root-folder path`

**Parameters:**
- `path`: `string!` Default root folder path

**Returns:** `logic!` Success status

**Notes:** Sets the default root folder for serving files.

### URL and Navigation

#### webui-open-url

Opens a URL in the default browser.

**Signature:** `webui-open-url url`

**Parameters:**
- `url`: `string!` URL to open

**Returns:** `none!` None

**Notes:** Opens the specified URL in the default web browser.

### Window Properties

#### webui-set-hide

Hides or shows a window.

**Signature:** `webui-set-hide window-id hidden`

**Parameters:**
- `window-id`: `integer!` Window ID
- `hidden`: `logic!` Whether to hide window

**Returns:** `none!` None

**Notes:** Hides or shows the specified window.

#### webui-set-size

Sets the size of a window.

**Signature:** `webui-set-size window-id width height`

**Parameters:**
- `window-id`: `integer!` Window ID
- `width`: `integer!` Window width
- `height`: `integer!` Window height

**Returns:** `none!` None

**Notes:** Sets the size of the specified window.

#### webui-set-position

Sets the position of a window.

**Signature:** `webui-set-position window-id x y`

**Parameters:**
- `window-id`: `integer!` Window ID
- `x`: `integer!` X coordinate
- `y`: `integer!` Y coordinate

**Returns:** `none!` None

**Notes:** Sets the position of the specified window.

#### webui-set-profile

Sets the browser profile for a window.

**Signature:** `webui-set-profile window-id name path`

**Parameters:**
- `window-id`: `integer!` Window ID
- `name`: `string!` Profile name
- `path`: `string!` Profile path

**Returns:** `none!` None

**Notes:** Sets the browser profile for the specified window.

#### webui-get-size

Gets the size of a window.

**Signature:** `webui-get-size window-id`

**Parameters:**
- `window-id`: `integer!` Window ID

**Returns:** `block!` [width height]

**Notes:** Returns the default size of the specified window (800x600) as a block.

#### webui-get-position

Gets the position of a window.

**Signature:** `webui-get-position window-id`

**Parameters:**
- `window-id`: `integer!` Window ID

**Returns:** `block!` [x y]

**Notes:** Returns the default position of the specified window (100x100) as a block.

#### webui-set-icon

Sets the icon for a window.

**Signature:** `webui-set-icon window-id icon icon-type`

**Parameters:**
- `window-id`: `integer!` Window ID
- `icon`: `string!` Icon data or path
- `icon-type`: `string!` Icon type

**Returns:** `none!` None

**Notes:** Sets the icon for the specified window.

### Data Transfer

#### webui-send-raw

Sends raw binary data to JavaScript.

**Signature:** `webui-send-raw window-id function raw-data`

**Parameters:**
- `window-id`: `integer!` Window ID
- `function`: `string!` JavaScript function name
- `raw-data`: `binary!` Raw data to send

**Returns:** `none!` None

**Notes:** Sends raw binary data to JavaScript in the specified window.

## Event Semantics

Events are processed via WebUI C API callbacks. Handlers execute with auto-bound locals:
- `event-element`: Element name/ID (string!)
- `event-data`: Event data (string! or none!)
- `event-window-id`: Window ID (integer!)
- `event-number`: Event number (integer!)

## Error Handling

- Type mismatches: `verror.NewScriptError` with category "webui" and appropriate ID
- WebUI is always available in standard builds (CGO required)

## CGO Requirements

- Requires CGO enabled (`CGO_ENABLED=1`)
- Platform dependencies:
  - Linux: WebKitGTK
  - Windows: Edge WebView2
  - macOS: WebKit
- Build with `webui` tag to enable functionality

## Implementation Notes

- Direct wrappers around WebUI C API functions
- Window IDs are integers (size_t from C API)
- Event handlers stored in global map for callback execution
- Synchronous event handling via WebUI callbacks with mutex protection
- webui-wait provides the main blocking event loop
- Lifecycle: Windows managed by WebUI C library