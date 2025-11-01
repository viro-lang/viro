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

	"github.com/marcin-radoszewski/viro/internal/core"
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
	stepState   StepState // New: state for step-by-step execution
}

// StepState manages pausing and resuming during step-by-step execution.
type StepState struct {
	Paused      bool          // Whether execution is currently paused
	WaitChan    chan struct{} // Channel to pause/resume execution
	CurrentExpr core.Value    // Current expression being evaluated
	CurrentPos  int           // Position in current block
	FrameIndex  int           // Current frame index
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
		stepState: StepState{
			Paused:   false,
			WaitChan: make(chan struct{}, 1), // Buffered to avoid blocking
		},
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

// PauseExecution pauses the evaluator at the current expression.
// Called by the evaluator when stepping or hitting a breakpoint.
// This method blocks until ResumeExecution is called.
func (d *Debugger) PauseExecution(expr core.Value, pos, frameIdx int) {
	d.mu.Lock()
	d.stepState.Paused = true
	d.stepState.CurrentExpr = expr
	d.stepState.CurrentPos = pos
	d.stepState.FrameIndex = frameIdx
	d.mu.Unlock()

	// Wait for resume signal (blocks here)
	<-d.stepState.WaitChan

	d.mu.Lock()
	d.stepState.Paused = false
	d.mu.Unlock()
}

// ResumeExecution resumes paused execution.
// Called when user issues step, continue, or other resume commands.
// Safe to call multiple times - extra signals are ignored.
func (d *Debugger) ResumeExecution() {
	select {
	case d.stepState.WaitChan <- struct{}{}:
		// Successfully sent resume signal
	default:
		// Channel already has a signal pending, ignore
	}
}

// ShouldPause returns true if the evaluator should pause before the next expression.
// This is called by the evaluator to determine if it should pause.
// Thread-safe and atomic.
func (d *Debugger) ShouldPause() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stepping
}

// IsPaused returns true if execution is currently paused.
func (d *Debugger) IsPaused() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stepState.Paused
}

// GetCurrentStepInfo returns information about the current paused state.
func (d *Debugger) GetCurrentStepInfo() (core.Value, int, int, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stepState.CurrentExpr, d.stepState.CurrentPos, d.stepState.FrameIndex, d.stepState.Paused
}

// GetFrameLocals returns the local variables in the specified frame.
func (d *Debugger) GetFrameLocals(eval core.Evaluator, frameIdx int) map[string]core.Value {
	if eval == nil {
		return make(map[string]core.Value)
	}

	frame := eval.GetFrameByIndex(frameIdx)
	if frame == nil {
		return make(map[string]core.Value)
	}

	bindings := frame.GetAll()
	result := make(map[string]core.Value, len(bindings))
	for _, binding := range bindings {
		result[binding.Symbol] = binding.Value
	}
	return result
}

// GetCallStack returns the current call stack as a slice of function names.
func (d *Debugger) GetCallStack(eval core.Evaluator) []string {
	if eval == nil {
		return []string{}
	}

	return eval.GetCallStack()
}
