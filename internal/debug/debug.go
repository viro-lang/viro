package debug

import (
	"fmt"
	"sync"
	"time"

	"github.com/marcin-radoszewski/viro/internal/trace"
)

type Debugger struct {
	mu          sync.Mutex
	breakpoints map[string]int
	nextID      int
	mode        DebugMode
}

type DebugMode int

const (
	DebugModeOff DebugMode = iota
	DebugModeActive
)

func (m DebugMode) String() string {
	switch m {
	case DebugModeOff:
		return "off"
	case DebugModeActive:
		return "active"
	default:
		return "unknown"
	}
}

var GlobalDebugger *Debugger

func InitDebugger() {
	GlobalDebugger = &Debugger{
		breakpoints: make(map[string]int),
		nextID:      1,
		mode:        DebugModeOff,
	}
}

func (d *Debugger) SetBreakpoint(word string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	id := d.nextID
	d.nextID++
	d.breakpoints[word] = id
	d.mode = DebugModeActive
	return id
}

func (d *Debugger) RemoveBreakpoint(word string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.breakpoints[word]; exists {
		delete(d.breakpoints, word)
		if len(d.breakpoints) == 0 {
			d.mode = DebugModeOff
		}
		return true
	}
	return false
}

func (d *Debugger) HasBreakpoint(word string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, exists := d.breakpoints[word]
	return exists
}

func (d *Debugger) Mode() DebugMode {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.mode
}

func (d *Debugger) Enable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mode = DebugModeActive
}

func (d *Debugger) Disable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mode = DebugModeOff
	d.breakpoints = make(map[string]int)
}

func (d *Debugger) RemoveBreakpointByID(id int64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	for word, bpID := range d.breakpoints {
		if int64(bpID) == id {
			delete(d.breakpoints, word)
			if len(d.breakpoints) == 0 {
				d.mode = DebugModeOff
			}
			return true
		}
	}
	return false
}

// HandleBreakpoint checks for breakpoints and emits trace events.
// Called by evaluator before word evaluation to centralize breakpoint handling.
func (d *Debugger) HandleBreakpoint(word string) {
	if !d.HasBreakpoint(word) {
		return
	}

	// Emit trace event if tracing is enabled
	if trace.GlobalTraceSession != nil && trace.GlobalTraceSession.IsEnabled() {
		trace.GlobalTraceSession.Emit(trace.TraceEvent{
			Timestamp: time.Now(),
			Word:      "debug",
			Value:     fmt.Sprintf("breakpoint hit: %s", word),
			Duration:  0,
		})
	}
}
