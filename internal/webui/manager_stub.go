//go:build webui_stub

package webui

import (
	"github.com/marcin-radoszewski/viro/internal/core"
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
	m.windows[id] = &windowState{
		id:              id,
		ready:           false,
		closed:          false,
		handlers:        make(map[string][]HandlerEntry),
		handlerBindings: make(map[string]handlerBinding),
		bridgeLoaded:    false,
	}

	return id, nil
}

func (m *manager) Render(windowID uint32, markup core.Value, options map[string]core.Value) error {
	m.mu.Lock()
	window, exists := m.windows[windowID]
	if !exists {
		m.mu.Unlock()
		return ErrWindowNotFound
	}
	if window.closed {
		m.mu.Unlock()
		return ErrWindowClosed
	}

	window.ready = false
	m.mu.Unlock()

	m.mu.Lock()
	window.ready = true
	m.mu.Unlock()

	return nil
}

func (m *manager) Inject(windowID uint32, html core.Value) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	window, exists := m.windows[windowID]
	if !exists {
		return ErrWindowNotFound
	}
	if window.closed {
		return ErrWindowClosed
	}

	return nil
}

func (m *manager) Send(windowID uint32, message string, payload core.Value) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	window, exists := m.windows[windowID]
	if !exists {
		return ErrWindowNotFound
	}
	if window.closed {
		return ErrWindowClosed
	}

	return nil
}

func (m *manager) RegisterEvent(entry HandlerEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	window, exists := m.windows[entry.WindowID]
	if !exists {
		return ErrWindowNotFound
	}
	if window.closed {
		return ErrWindowClosed
	}

	window.handlers[entry.Event] = append(window.handlers[entry.Event], entry)

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

	window, exists := m.windows[windowID]
	if !exists || window.closed {
		return false
	}

	window.closed = true

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

	for _, window := range m.windows {
		window.closed = true
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

	window, exists := m.windows[windowID]
	return exists && window.ready && !window.closed
}
