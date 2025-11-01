// Package trace provides tracing and observability infrastructure for Viro.
//
// This package manages trace event collection and output, supporting:
// - Structured event emission (JSON format)
// - Filtering by word patterns
// - File and stderr output with rotation
// - Port lifecycle tracing
// - Object operation tracing
//
// Per Feature 002 FR-015: Tracing system with structured events.
package trace

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// TraceSession manages trace event collection and output (Feature 002).
// Per FR-015: emits structured events to stderr by default, optional file redirection.
// Per research.md: uses lumberjack for rotating log file support.
type TraceSession struct {
	mu            sync.Mutex
	enabled       atomic.Bool
	sink          io.Writer          // Output destination (stderr or file)
	logger        *lumberjack.Logger // Optional file logger
	atomicFilters atomic.Value       // Stores *TraceFilters for lock-free reads
	stepCounter   int64              // Monotonic step counter
	callback      atomic.Value       // Stores func(TraceEvent) for lock-free reads
}

// TraceFilters controls which events are emitted.
type TraceFilters struct {
	IncludeWords []string      // Only trace these words (empty = all)
	ExcludeWords []string      // Never trace these words
	MinDuration  time.Duration // Only trace operations taking longer than this

	Verbose     bool // Include frame state
	StepLevel   int  // 0=calls only, 1=expressions, 2=all
	IncludeArgs bool // Include function arguments
	MaxDepth    int  // Only trace up to this call depth (0=unlimited)
}

// TraceEvent represents a single trace event (per FR-015 requirements).
type TraceEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`    // String representation of evaluated value
	Word      string    `json:"word"`     // Word being evaluated (if applicable)
	Duration  int64     `json:"duration"` // Nanoseconds spent evaluating

	EventType  string            `json:"event_type,omitempty"`  // "eval", "call", "return", "block-enter", "block-exit"
	Step       int64             `json:"step,omitempty"`        // Execution step counter
	Depth      int               `json:"depth,omitempty"`       // Call stack depth
	Position   int               `json:"position,omitempty"`    // Position in current block
	Expression string            `json:"expression,omitempty"`  // Mold of expression being evaluated
	Args       map[string]string `json:"args,omitempty"`        // Function arguments (name -> value)
	Frame      map[string]string `json:"frame,omitempty"`       // Local variables (only in verbose mode)
	ParentExpr string            `json:"parent_expr,omitempty"` // Context expression
	Error      string            `json:"error,omitempty"`       // Error message if evaluation failed
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

	ts := &TraceSession{
		sink:   sink,
		logger: logger,
	}
	ts.enabled.Store(false)
	ts.atomicFilters.Store(&TraceFilters{})
	GlobalTraceSession = ts

	return nil
}

// InitTraceSilent initializes the global trace session with output suppressed.
// Used for profiling where only callbacks are needed, not JSON output.
// Uses io.Discard for cross-platform compatibility (works on Windows, Unix, etc).
func InitTraceSilent() error {
	ts := &TraceSession{
		sink:   io.Discard,
		logger: nil,
	}
	ts.enabled.Store(false)
	ts.atomicFilters.Store(&TraceFilters{})
	GlobalTraceSession = ts

	return nil
}

// Enable activates tracing with optional filters.
func (ts *TraceSession) Enable(filters TraceFilters) {
	ts.atomicFilters.Store(&filters)
	ts.enabled.Store(true)
}

// Disable stops tracing.
func (ts *TraceSession) Disable() {
	ts.enabled.Store(false)
}

// IsEnabled returns true if tracing is active.
func (ts *TraceSession) IsEnabled() bool {
	return ts.enabled.Load()
}

// Emit writes a trace event if tracing is enabled and event passes filters.
func (ts *TraceSession) Emit(event TraceEvent) {
	if !ts.enabled.Load() {
		return
	}

	filters := ts.atomicFilters.Load().(*TraceFilters)

	// Apply filters
	if len(filters.IncludeWords) > 0 {
		found := slices.Contains(filters.IncludeWords, event.Word)
		if !found {
			return
		}
	}

	if slices.Contains(filters.ExcludeWords, event.Word) {
		return
	}

	if filters.MinDuration > 0 && time.Duration(event.Duration) < filters.MinDuration {
		return
	}

	callback := ts.callback.Load()
	if callback != nil {
		callback.(func(TraceEvent))(event)
	}

	// Serialize as JSON (per FR-015: line-delimited JSON)
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "trace serialization error: %v\n", err)
		return
	}

	// Write to sink (mutex-protected for safe concurrent writes)
	ts.mu.Lock()
	fmt.Fprintf(ts.sink, "%s\n", data)
	ts.mu.Unlock()
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

// NextStep increments and returns the next step counter value.
func (ts *TraceSession) NextStep() int64 {
	return atomic.AddInt64(&ts.stepCounter, 1)
}

// ResetStepCounter resets the step counter to zero.
func (ts *TraceSession) ResetStepCounter() {
	atomic.StoreInt64(&ts.stepCounter, 0)
}

// GetVerbose returns whether verbose mode is enabled.
func (ts *TraceSession) GetVerbose() bool {
	filters := ts.atomicFilters.Load().(*TraceFilters)
	return filters.Verbose
}

// GetIncludeArgs returns whether function arguments should be included.
func (ts *TraceSession) GetIncludeArgs() bool {
	filters := ts.atomicFilters.Load().(*TraceFilters)
	return filters.IncludeArgs
}

// ShouldTraceExpression returns whether expressions should be traced based on StepLevel.
func (ts *TraceSession) ShouldTraceExpression() bool {
	filters := ts.atomicFilters.Load().(*TraceFilters)
	return filters.StepLevel >= 1
}

// ShouldTraceAtDepth returns whether tracing should occur at the given depth.
func (ts *TraceSession) ShouldTraceAtDepth(depth int) bool {
	filters := ts.atomicFilters.Load().(*TraceFilters)
	if filters.MaxDepth == 0 {
		return true
	}
	return depth <= filters.MaxDepth
}

// SetCallback registers a callback function that will be invoked for each emitted trace event.
// This is useful for profiling and other real-time analysis.
// The callback is invoked lock-free before the event is serialized to JSON.
func (ts *TraceSession) SetCallback(callback func(TraceEvent)) {
	ts.callback.Store(callback)
}

// Port lifecycle trace event helpers (T076)

// TracePortOpen emits a trace event for port open operation.
func TracePortOpen(scheme, spec string) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "open",
		Value:     fmt.Sprintf("port opened: %s (%s)", spec, scheme),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}

// TracePortRead emits a trace event for port read operation.
func TracePortRead(scheme, spec string, bytes int) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "read",
		Value:     fmt.Sprintf("port read: %s (%s) %d bytes", spec, scheme, bytes),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}

// TracePortWrite emits a trace event for port write operation.
func TracePortWrite(scheme, spec string, bytes int) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "write",
		Value:     fmt.Sprintf("port write: %s (%s) %d bytes", spec, scheme, bytes),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}

// TracePortClose emits a trace event for port close operation.
func TracePortClose(scheme, spec string) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "close",
		Value:     fmt.Sprintf("port closed: %s (%s)", spec, scheme),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}

// TracePortError emits a trace event for port error.
func TracePortError(scheme, spec string, err error) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "error",
		Value:     fmt.Sprintf("port error: %s (%s) - %v", spec, scheme, err),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}

// TraceObjectCreate emits a trace event for object creation (Feature 002, US3).
func TraceObjectCreate(fieldCount int) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "object",
		Value:     fmt.Sprintf("object created: fields:%d", fieldCount),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}

// TraceObjectFieldRead emits a trace event for object field access (Feature 002, US3).
func TraceObjectFieldRead(field string, found bool) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	status := "found"
	if !found {
		status = "not-found"
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "select",
		Value:     fmt.Sprintf("object field read: field:%s (%s)", field, status),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}

// TraceObjectFieldWrite emits a trace event for object field mutation (Feature 002, US3).
func TraceObjectFieldWrite(field string, newValue string) {
	if GlobalTraceSession == nil || !GlobalTraceSession.IsEnabled() {
		return
	}

	event := TraceEvent{
		Timestamp: time.Now(),
		Word:      "put",
		Value:     fmt.Sprintf("object field write: field:%s value:%s", field, newValue),
		Duration:  0,
	}
	GlobalTraceSession.Emit(event)
}
