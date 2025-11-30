# WebUI Integration

Viro provides WebUI integration through the `webui` object, allowing scripts to create and control native HTML/CSS/JS windows while running in the same CLI process.

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

Without this tag, `webui.*` natives are not available and will result in undefined word errors.

## API Overview

All WebUI functionality is accessed via dot-notation on the `webui` object:

- `webui.window spec-block` - Create a window
- `webui.render window markup options?` - Render HTML content
- `webui.inject window html` - Inject HTML snippet
- `webui.send window message payload` - Send data to JavaScript
- `webui.on window event selector handler-block` - Register event handler
- `webui.poll window` - Process pending events
- `webui.close window` - Close window
- `webui.ready? window` - Check if window is ready

## Lifecycle Behavior

1. **Window Creation:** `webui.window` creates a native window and returns a handle
2. **Rendering:** `webui.render` loads HTML content and shows the window
3. **Event Handling:** Scripts register handlers with `webui.on` which bind directly to go-webui events
4. **Communication:** Use `webui.send` to push data to JavaScript via injected scripts
5. **Event Processing:** `webui.poll none` blocks until all windows close; `webui.poll window` is a no-op
6. **Cleanup:** Windows close with `webui.close` or when the interpreter exits

## Event Loop Pattern

WebUI uses synchronous event handling - events are processed immediately when they occur:

```viro
window: webui.window [title: "My App" width: 800 height: 600]

webui.on window "click" "#button" [
    print ["Button clicked:" event-payload]
    ; Send data back to JavaScript
    webui.send window "update" [message: "Button was clicked!"]
]

webui.render window "<html><body><button id='button'>Click me</button></body></html>"

; Block until all windows close
webui.poll none
print "All windows closed"
```

## Hot Reload

For development, use `webui.render` with file sources:

```viro
webui.render window %index.html [source: %index.html]
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

Viro receives this as an event that can be handled with `webui.on`.

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
- **Event handling**: Uses a JavaScript bridge to attach DOM event listeners; requires `webui.js` to be included in HTML