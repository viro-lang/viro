//go:build webui && !webui_stub

package webui

import (
	"fmt"
	"strings"
	"time"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	ui "github.com/webui-dev/go-webui/v2"
)

func NewManager() Manager {
	return newManager()
}

func newManager() Manager {
	return &manager{
		windows: make(map[uint32]*windowState),
		events:  make(chan EventMessage, 100),
		done:    make(chan struct{}),
	}
}

func (m *manager) CreateWindow(spec *WindowSpec) (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, ErrManagerClosed
	}

	m.nextID++
	id := m.nextID
	window := ui.NewWindow()
	state := &windowState{
		id:              id,
		window:          window,
		ready:           false,
		closed:          false,
		handlers:        make(map[string][]HandlerEntry),
		handlerBindings: make(map[string]handlerBinding),
		bridgeLoaded:    false,
	}
	m.windows[id] = state

	if spec != nil {
		if spec.Width > 0 && spec.Height > 0 {
			window.SetSize(uint(spec.Width), uint(spec.Height))
		}
		if spec.Resizable {
			window.SetResizable(true)
		}
		if spec.Icon != "" {
			window.SetIcon(spec.Icon, ui.GetMimeType(spec.Icon))
		}
		// Note: title is set via DOM script in Render, not supported directly
		// Note: debug? is not supported by go-webui
	}

	// Bind a special handler for ready signal
	ui.Bind(window, "__viro_ready", func(e ui.Event) any {
		m.mu.Lock()
		state.ready = true
		m.mu.Unlock()
		return nil
	})

	return id, nil
}

func (m *manager) Render(windowID uint32, markup core.Value, options map[string]core.Value) error {
	m.mu.Lock()
	state, exists := m.windows[windowID]
	if !exists {
		m.mu.Unlock()
		return ErrWindowNotFound
	}
	if state.closed {
		m.mu.Unlock()
		return ErrWindowClosed
	}

	state.ready = false
	m.mu.Unlock()

	content := ""
	if str, ok := markup.(*value.StringValue); ok {
		content = str.String()
	} else if bin, ok := markup.(*value.BinaryValue); ok {
		content = string(bin.Bytes())
	}

	window := state.window.(ui.Window)

	// Inject title if specified in options
	if options != nil {
		if titleVal, hasTitle := options["title"]; hasTitle {
			if titleStr, ok := titleVal.(*value.StringValue); ok {
				titleScript := "document.title = " + jsQuote(titleStr.String()) + ";"
				content = strings.Replace(content, "<head>", "<head><script>"+titleScript+"</script>", 1)
			}
		}
	}

	// Show the content
	err := window.Show(content)
	if err != nil {
		return err
	}

	// Inject the JS bridge and re-bind handlers
	m.injectBridgeAndRebind(windowID)

	return nil
}

func (m *manager) Inject(windowID uint32, html core.Value) error {
	m.mu.Lock()
	state, exists := m.windows[windowID]
	m.mu.Unlock()

	if !exists {
		return ErrWindowNotFound
	}
	if state.closed {
		return ErrWindowClosed
	}

	content := ""
	if str, ok := html.(*value.StringValue); ok {
		content = str.String()
	} else if bin, ok := html.(*value.BinaryValue); ok {
		content = string(bin.Bytes())
	}

	window := state.window.(ui.Window)
	js := jsQuote(content)
	window.Run("document.body.innerHTML += " + js)

	return nil
}

func (m *manager) Send(windowID uint32, message string, payload core.Value) error {
	m.mu.Lock()
	state, exists := m.windows[windowID]
	m.mu.Unlock()

	if !exists {
		return ErrWindowNotFound
	}
	if state.closed {
		return ErrWindowClosed
	}

	jsonPayload, err := value.ToJSON(payload)
	if err != nil {
		return err
	}

	window := state.window.(ui.Window)
	js := "window.dispatchEvent(new CustomEvent(" + jsQuote(message) + ", {detail: " + jsonPayload + "}));"
	window.Run(js)

	return nil
}

func (m *manager) RegisterEvent(entry HandlerEntry) error {
	m.mu.Lock()
	state, exists := m.windows[entry.WindowID]
	if !exists {
		m.mu.Unlock()
		return ErrWindowNotFound
	}
	if state.closed {
		m.mu.Unlock()
		return ErrWindowClosed
	}

	// Generate unique binding name
	bindingName := fmt.Sprintf("viro_evt_%d_%d", entry.WindowID, len(state.handlerBindings))
	binding := handlerBinding{
		event:    entry.Event,
		selector: entry.Selector,
		fnName:   bindingName,
	}
	state.handlerBindings[bindingName] = binding
	state.handlers[entry.Event] = append(state.handlers[entry.Event], entry)
	m.mu.Unlock()

	window := state.window.(ui.Window)

	// Bind the Go callback
	ui.Bind(window, bindingName, func(e ui.Event) any {
		var payload core.Value = value.NewNoneVal()
		var raw string
		if e.GetCount() > 0 {
			if arg, err := ui.GetArg[string](e); err == nil {
				raw = arg
				if parsed, err := value.FromJSON(arg); err == nil {
					payload = parsed
				} else {
					payload = value.NewStrVal(arg)
				}
			}
		}

		eventMsg := EventMessage{
			HandlerEntry: entry,
			Payload:      payload,
			Raw:          raw,
			Timestamp:    time.Now().Unix(),
		}
		select {
		case m.events <- eventMsg:
		default:
		}
		return nil
	})

	// If bridge is loaded, immediately bind the DOM listener
	if state.bridgeLoaded {
		bindScript := fmt.Sprintf("if (window._viroBind) { window._viroBind(%s, %s, %s); }",
			jsQuote(entry.Event),
			jsQuote(entry.Selector),
			jsQuote(bindingName))
		window.Run(bindScript)
	}

	return nil
}

func (m *manager) Poll(windowID *uint32) ([]EventMessage, error) {
	select {
	case <-m.done:
		return nil, nil
	default:
	}

	var events []EventMessage
	for {
		select {
		case event := <-m.events:
			if windowID == nil || event.HandlerEntry.WindowID == *windowID {
				events = append(events, event)
			}
		default:
			return events, nil
		}
	}
}

func (m *manager) Close(windowID uint32) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.windows[windowID]
	if !exists || state.closed {
		return false
	}

	state.closed = true
	window := state.window.(ui.Window)
	window.Close()

	allClosed := true
	for _, w := range m.windows {
		if !w.closed {
			allClosed = false
			break
		}
	}

	if allClosed && !m.closed {
		m.closed = true
		close(m.done)
	}

	return true
}

func (m *manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, state := range m.windows {
		if !state.closed {
			window := state.window.(ui.Window)
			window.Close()
			state.closed = true
		}
	}

	if !m.closed {
		m.closed = true
		close(m.done)
	}
}

func (m *manager) Done() <-chan struct{} {
	return m.done
}

func (m *manager) Ready(windowID uint32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.windows[windowID]
	return exists && state.ready && !state.closed
}

// injectBridgeAndRebind injects the JavaScript bridge and re-registers all event handlers
func (m *manager) injectBridgeAndRebind(windowID uint32) {
	m.mu.Lock()
	state, exists := m.windows[windowID]
	if !exists || state.closed {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()

	window := state.window.(ui.Window)

	// Inject the bridge script
	bridgeScript := `
if (!window._viroBridgeLoaded) {
	window._viroBridgeLoaded = true;

	window._viroBind = function(event, selector, fnName) {
		var elements = document.querySelectorAll(selector);
		elements.forEach(function(el) {
			el.addEventListener(event, function(e) {
				var detail = {
					selector: selector,
					event: event,
					target: e.target ? e.target.outerHTML : null,
					data: e.detail || null
				};
				window.webui[fnName](JSON.stringify(detail));
			});
		});
	};

	// Signal ready
	window.webui['__viro_ready']();
}
`
	window.Run(bridgeScript)

	// Re-bind all existing handlers
	m.mu.Lock()
	for _, binding := range state.handlerBindings {
		bindScript := fmt.Sprintf("window._viroBind(%s, %s, %s);",
			jsQuote(binding.event),
			jsQuote(binding.selector),
			jsQuote(binding.fnName))
		window.Run(bindScript)
	}
	m.mu.Unlock()
}

func jsQuote(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return "\"" + s + "\""
}
