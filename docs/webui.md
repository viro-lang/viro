# WebUI Integration

Viro provides direct WebUI integration through kebab-case native functions, allowing scripts to create and control native HTML/CSS/JS windows while running in the same CLI process.

## Dependencies

The WebUI integration requires the `github.com/webui-dev/go-webui` package and CGO support.

### Installation

The WebUI integration uses `github.com/webui-dev/go-webui/v2` which is included as a git submodule in `dependencies/go-webui/v2`.

### Build Requirements

- Set `CGO_ENABLED=1` in environment
- Platform-specific packages:

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install webkit2gtk3-devel

# Arch
sudo pacman -S webkit2gtk
```

**Windows:**
- Edge WebView2 (usually pre-installed on modern Windows)

**macOS:**
- WebKit framework (built-in)

### Build Tag

WebUI functionality is only available when built with the `webui` build tag:

```bash
go build -tags webui ./cmd/viro
```

Without this tag, `webui-*` natives are not available and will result in undefined word errors.

## API Overview

All WebUI functionality is accessed via direct kebab-case native functions:

- `webui-new-window` - Create a new window
- `webui-show window-id content` - Show window with HTML content
- `webui-run window-id script` - Execute JavaScript in window
- `webui-bind window-id element handler` - Bind event handler to element
- `webui-wait` - Wait for all windows to close
- `webui-close window-id` - Close a window
- `webui-is-shown window-id` - Check if window is shown

## Lifecycle Behavior

1. **Window Creation:** `webui-new-window` creates a native window and returns a window ID
2. **Rendering:** `webui-show` loads HTML content and shows the window
3. **Event Handling:** Scripts register handlers with `webui-bind` which execute immediately when events occur
4. **Communication:** Use `webui-run` to execute JavaScript in the window
5. **Event Processing:** `webui-wait` blocks until all windows close
6. **Cleanup:** Windows close with `webui-close` or when the interpreter exits

## Event Loop Pattern

WebUI uses synchronous event handling - events are processed immediately when they occur:

```viro
window-id: webui-new-window

webui-bind window-id "button" [
    print ["Button clicked:" event-data]
    ; Send data back to JavaScript
    webui-run window-id "updateOutput('Button was clicked!')"
]

webui-show window-id "<html><body><button onclick='webui.button()'>Click me</button><div id='output'></div></body></html>"

; Block until all windows close
webui-wait
print "All windows closed"
```

## Hot Reload

For development, load HTML content from files:

```viro
html-content: read %index.html
webui-show window-id html-content
```

Manual reload with file watchers (using external tools like `entr`):

```bash
echo "index.html" | entr -r ./viro script.viro
```

## JavaScript Integration

JavaScript can communicate back to Viro via custom events:

```javascript
// Send data to Viro
window.dispatchEvent(new CustomEvent('viro-event', {
    detail: { message: 'Hello from JS', value: 42 }
}));
```

Viro receives this as an event that can be handled with `webui-bind`.

## Error Handling

WebUI operations may fail due to:
- Missing GUI libraries
- Invalid window handles
- Content-type mismatches
- Serialization errors

Check return values and handle errors appropriately in scripts.

## Running GUI Scripts

Execute WebUI scripts normally:

```bash
./viro gui-script.viro
```

The interpreter will remain running until all windows are closed or the script exits.

## Limitations

- Requires GUI environment (not suitable for headless servers)
- CGO dependency may complicate cross-compilation
- Event handlers execute synchronously (may block UI)
- No built-in file watching (use external tools for hot reload)
- **Title setting**: Window titles are set via DOM script injection, not native window title bars
- **Debug mode**: The `debug?` spec key is not supported by the underlying go-webui library
- **Event handling**: Uses direct WebUI C API callbacks; JavaScript calls bound functions directly