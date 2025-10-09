package native

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// TraceSession manages trace event collection and output (Feature 002).
// Per FR-015: emits structured events to stderr by default, optional file redirection.
// Per research.md: uses lumberjack for rotating log file support.
type TraceSession struct {
	mu       sync.Mutex
	enabled  bool
	sink     io.Writer        // Output destination (stderr or file)
	logger   *lumberjack.Logger // Optional file logger
	filters  TraceFilters
}

// TraceFilters controls which events are emitted.
type TraceFilters struct {
	IncludeWords []string      // Only trace these words (empty = all)
	ExcludeWords []string      // Never trace these words
	MinDuration  time.Duration // Only trace operations taking longer than this
}

// TraceEvent represents a single trace event (per FR-015 requirements).
type TraceEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`     // String representation of evaluated value
	Word      string    `json:"word"`      // Word being evaluated (if applicable)
	Duration  int64     `json:"duration"`  // Nanoseconds spent evaluating
	Depth     int       `json:"depth"`     // Call stack depth
}

// GlobalTraceSession is the active trace session (singleton).
var GlobalTraceSession *TraceSession

// InitTrace initializes the global trace session.
// Called during REPL initialization with CLI flag values.
func InitTrace(traceFile string, maxSizeMB int) error {
	var sink io.Writer = os.Stderr // Default per FR-015

	var logger *lumberjack.Logger
	if traceFile != "" {
		// User requested file output
		logger = &lumberjack.Logger{
			Filename:   traceFile,
			MaxSize:    maxSizeMB, // Per clarification: 50MB default
			MaxBackups: 5,         // Per FR-015: retain 5 backup files
			MaxAge:     0,         // No age-based deletion
			Compress:   true,      // Per FR-015: compress backups
		}
		sink = logger
	}

	GlobalTraceSession = &TraceSession{
		enabled: false, // Disabled by default, enabled via trace --on
		sink:    sink,
		logger:  logger,
		filters: TraceFilters{},
	}

	return nil
}

// Enable activates tracing with optional filters.
func (ts *TraceSession) Enable(filters TraceFilters) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.enabled = true
	ts.filters = filters
}

// Disable stops tracing.
func (ts *TraceSession) Disable() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.enabled = false
}

// IsEnabled returns true if tracing is active.
func (ts *TraceSession) IsEnabled() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.enabled
}

// Emit writes a trace event if tracing is enabled and event passes filters.
func (ts *TraceSession) Emit(event TraceEvent) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.enabled {
		return
	}

	// Apply filters
	if len(ts.filters.IncludeWords) > 0 {
		found := false
		for _, w := range ts.filters.IncludeWords {
			if event.Word == w {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	for _, w := range ts.filters.ExcludeWords {
		if event.Word == w {
			return
		}
	}

	if ts.filters.MinDuration > 0 && time.Duration(event.Duration) < ts.filters.MinDuration {
		return
	}

	// Serialize as JSON (per FR-015: line-delimited JSON)
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "trace serialization error: %v\n", err)
		return
	}

	// Write to sink
	fmt.Fprintf(ts.sink, "%s\n", data)
}

// Close flushes and closes the trace session.
func (ts *TraceSession) Close() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.logger != nil {
		return ts.logger.Close()
	}
	return nil
}

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
