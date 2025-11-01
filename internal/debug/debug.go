// Package debug provides debugging infrastructure for Viro.
//
// This package manages breakpoint state and stepping control, supporting:
// - Breakpoint management (set, remove, check)
// - Stepping modes (single-step, continue)
// - Debug mode states (off, active, stepping)
//
// Per Feature 002 FR-016: Debugger with breakpoints and stepping control.
package debug

import (
	"fmt"
	"sync"
	"time"

	"github.com/marcin-radoszewski/viro/internal/trace"
)

// Debugger manages breakpoint state and stepping control (Feature 002).
// Per FR-016: supports breakpoint, remove, step, continue, stack, locals commands.
type Debugger struct {
	mu          sync.Mutex
	breakpoints map[string]int // word -> breakpoint ID
	nextID      int
	mode        DebugMode
	stepping    bool
}

// DebugMode controls debugger behavior.
type DebugMode int

const (
	DebugModeOff      DebugMode = iota // Debugger disabled
	DebugModeActive                    // Breakpoints active
	DebugModeStepping                  // Single-stepping mode
)

func (m DebugMode) String() string {
	switch m {
	case DebugModeOff:
		return "off"
	case DebugModeActive:
		return "active"
	case DebugModeStepping:
		return "stepping"
	default:
		return "unknown"
	}
}

// GlobalDebugger is the active debugger instance (singleton).
var GlobalDebugger *Debugger

// InitDebugger initializes the global debugger.
func InitDebugger() {
	GlobalDebugger = &Debugger{
		breakpoints: make(map[string]int),
		nextID:      1,
		mode:        DebugModeOff,
		stepping:    false,
	}
}

// SetBreakpoint adds a breakpoint on the given word.
func (d *Debugger) SetBreakpoint(word string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	id := d.nextID
	d.nextID++
	d.breakpoints[word] = id
	d.mode = DebugModeActive
	return id
}

// RemoveBreakpoint removes a breakpoint by word.
func (d *Debugger) RemoveBreakpoint(word string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.breakpoints[word]; exists {
		delete(d.breakpoints, word)
		if len(d.breakpoints) == 0 && !d.stepping {
			d.mode = DebugModeOff
		}
		return true
	}
	return false
}

// HasBreakpoint returns true if a breakpoint is set on the word.
func (d *Debugger) HasBreakpoint(word string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, exists := d.breakpoints[word]
	return exists
}

// EnableStepping activates single-step mode.
func (d *Debugger) EnableStepping() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mode = DebugModeStepping
	d.stepping = true
}

// DisableStepping deactivates single-step mode.
func (d *Debugger) DisableStepping() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.stepping = false
	if len(d.breakpoints) == 0 {
		d.mode = DebugModeOff
	} else {
		d.mode = DebugModeActive
	}
}

// IsStepping returns true if single-step mode is active.
func (d *Debugger) IsStepping() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.stepping
}

// Mode returns the current debugger mode.
func (d *Debugger) Mode() DebugMode {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.mode
}

// Enable activates the debugger in active mode.
func (d *Debugger) Enable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mode = DebugModeActive
}

// Disable deactivates the debugger and clears all breakpoints.
func (d *Debugger) Disable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mode = DebugModeOff
	d.breakpoints = make(map[string]int)
	d.stepping = false
}

// RemoveBreakpointByID removes a breakpoint by its ID.
// Returns true if a breakpoint was found and removed.
func (d *Debugger) RemoveBreakpointByID(id int64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	for word, bpID := range d.breakpoints {
		if int64(bpID) == id {
			delete(d.breakpoints, word)
			if len(d.breakpoints) == 0 && !d.stepping {
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

	// Future: Add interactive debugging logic here
}
