package webui

import (
	"errors"
	"sync"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

var (
	ErrManagerClosed      = errors.New("manager closed")
	ErrWindowNotFound     = errors.New("window not found")
	ErrWindowClosed       = errors.New("window closed")
	ErrFeatureUnavailable = errors.New("feature unavailable")
)

type WindowSpec struct {
	Title     string
	Width     int
	Height    int
	Debug     bool
	Resizable bool
	Icon      string
	Source    string
	HTML      string
}

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
	Raw          string
	Timestamp    int64
}

type Manager interface {
	CreateWindow(spec *WindowSpec) (uint32, error)
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
	id              uint32
	window          interface{} // ui.Window in real build, interface{} in stub
	ready           bool
	closed          bool
	handlers        map[string][]HandlerEntry
	handlerBindings map[string]handlerBinding // real build only
	bridgeLoaded    bool                      // real build only
}

type handlerBinding struct {
	event    string
	selector string
	fnName   string
}

type manager struct {
	mu      sync.RWMutex
	windows map[uint32]*windowState
	nextID  uint32
	events  chan EventMessage
	done    chan struct{}
	closed  bool
}
