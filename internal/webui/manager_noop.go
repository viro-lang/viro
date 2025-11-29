//go:build !webui && !webui_stub

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
	return 0, ErrFeatureUnavailable // Always fail for no-op manager
}

func (m *manager) Render(windowID uint32, markup core.Value, options map[string]core.Value) error {
	return nil
}

func (m *manager) Inject(windowID uint32, html core.Value) error {
	return nil
}

func (m *manager) Send(windowID uint32, message string, payload core.Value) error {
	return nil
}

func (m *manager) RegisterEvent(entry HandlerEntry) error {
	return nil
}

func (m *manager) Poll(windowID *uint32) ([]EventMessage, error) {
	return []EventMessage{}, nil
}

func (m *manager) Close(windowID uint32) bool {
	return false
}

func (m *manager) CloseAll() {
}

func (m *manager) Done() <-chan struct{} {
	return m.done
}

func (m *manager) Ready(windowID uint32) bool {
	return false
}
