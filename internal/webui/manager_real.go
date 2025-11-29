//go:build webui && !webui_stub

package webui

import (
	"strings"
	"sync"
	"time"

	"github.com/marcin-radoszewski/viro/internal/core"
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
		id:       id,
		window:   window,
		ready:    false,
		closed:   false,
		handlers: make(map[string][]HandlerEntry),
	}
	m.windows[id] = state

	// Apply window spec
	if spec != nil {
		if spec.Title != "" {
			window.SetTitle(spec.Title)
		}
		if spec.Width > 0 && spec.Height > 0 {
			window.SetSize(spec.Width, spec.Height)
		}
		// TODO: Apply other spec values (debug, resizable, icon, etc.)
	}

	// Bind ready event
	ui.Bind(window, "__ready", func(e ui.Event) any {
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

	// Reset ready state
	state.ready = false
	m.mu.Unlock()

	content := ""
	if str, ok := markup.(core.StringValue); ok {
		content = str.String()
	} else if bin, ok := markup.(core.BinaryValue); ok {
		content = string(bin.Bytes())
	}

	// Show the window
	state.window.Show(content)

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
	if str, ok := html.(core.StringValue); ok {
		content = str.String()
	} else if bin, ok := html.(core.BinaryValue); ok {
		content = string(bin.Bytes())
	}

	// Inject HTML - this might need a JS call
	js := jsQuote(content)
	state.window.Run("document.body.innerHTML += " + js)

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

	// Serialize payload to JSON
	jsonPayload, err := value.ToJSON(payload)
	if err != nil {
		return err
	}

	// Send data to JS - this might need a JS call or custom event
	js := "window.dispatchEvent(new CustomEvent(" + jsQuote(message) + ", {detail: " + jsonPayload + "}));"
	state.window.Run(js)

	return nil
}

func (m *manager) RegisterEvent(entry HandlerEntry) error {
	m.mu.Lock()
	state, exists := m.windows[entry.WindowID]
	m.mu.Unlock()

	if !exists {
		return ErrWindowNotFound
	}
	if state.closed {
		return ErrWindowClosed
	}

	// Allow multiple handlers per event/selector combination
	state.handlers[entry.Event] = append(state.handlers[entry.Event], entry)

	// Bind the event handler
	ui.Bind(state.window, entry.Event, func(e ui.Event) any {
		// Extract payload from event arguments
		var payload core.Value = value.NewNoneVal()
		var raw string
		if e.GetCount() > 0 {
			if arg, err := ui.GetArg[string](e); err == nil {
				raw = arg
				// Try to parse as JSON
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
			// Queue full, drop event (TODO: consider overflow handling)
		}
		return nil
	})

	return nil
}

func (m *manager) Poll(windowID *uint32) ([]EventMessage, error) {
	select {
	case <-m.done:
		return nil, nil
	default:
	}

	var events []EventMessage
	// Drain all queued events
	for {
		select {
		case event := <-m.events:
			if windowID == nil || event.HandlerEntry.WindowID == *windowID {
				events = append(events, event)
			}
		default:
			// No more events in queue
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
	state.window.Close()

	// Check if all windows closed
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
			state.window.Close()
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

// jsQuote escapes a string for safe inclusion in JavaScript code
func jsQuote(s string) string {
	// Simple implementation - replace backslash and quote
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return "\"" + s + "\""
}
