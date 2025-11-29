package webui

import (
	"sync"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

type HandlerEntry struct {
	WindowID      uint32
	Event         string
	Selector      string
	HandlerBlock  *value.BlockValue
	CapturedFrame int
}

type EventMessage struct {
	HandlerEntry HandlerEntry
	Payload      core.Value
	Timestamp    int64
}

type Manager interface {
	CreateWindow(spec map[string]core.Value) (uint32, error)
	Render(windowID uint32, markup core.Value, options map[string]core.Value) error
	Inject(windowID uint32, html core.Value) error
	Send(windowID uint32, message string, payload core.Value) error
	RegisterEvent(entry HandlerEntry) error
	Poll(windowID *uint32) ([]EventMessage, error)
	Close(windowID uint32) bool
	CloseAll()
	Done() <-chan struct{}
	Ready(windowID uint32) bool
}

type windowState struct {
	id     uint32
	ready  bool
	closed bool
}

type manager struct {
	mu      sync.RWMutex
	windows map[uint32]*windowState
	nextID  uint32
	events  chan EventMessage
	done    chan struct{}
	closed  bool
}
