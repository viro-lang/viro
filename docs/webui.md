# WebUI Integration

Viro provides WebUI integration through the `webui` object, allowing scripts to create and control native HTML/CSS/JS windows while running in the same CLI process.

## Dependencies

The WebUI integration requires the `github.com/webui-dev/go-webui` package and CGO support.

### Installation

```bash
go get github.com/webui-dev/go-webui
```

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

### Stub Build Tag

For headless environments (CI, testing without GUI), use the `webui_stub` build tag:

```bash
go build -tags webui_stub ./cmd/viro
go test -tags webui_stub ./...
```

This provides a stub implementation that satisfies the API without requiring CGO or GUI libraries.

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

1. **Initialization:** `webui.start` initializes the subsystem (auto-called by other methods)
2. **Window Creation:** `webui.window` creates a native window and returns a handle
3. **Rendering:** `webui.render` loads HTML content and waits for DOM ready
4. **Event Handling:** Scripts register handlers with `webui.on`, poll events with `webui.poll`
5. **Communication:** Use `webui.send` to push data to JavaScript, receive via events
6. **Cleanup:** Windows auto-close on interpreter exit, or manually with `webui.close`

## Event Loop Pattern

Typical GUI application structure:

```viro
window: webui.window [title: "My App" width: 800 height: 600]

webui.on window "click" "#button" [
    print ["Button clicked:" event-payload]
]

webui.render window "<html><body><button id='button'>Click me</button></body></html>"

forever [
    webui.poll window
    wait 0.1  ; Prevent busy loop
]
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