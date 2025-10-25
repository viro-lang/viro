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
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// TraceSession manages trace event collection and output (Feature 002).
// Per FR-015: emits structured events to stderr by default, optional file redirection.
// Per research.md: uses lumberjack for rotating log file support.
type TraceSession struct {
	mu      sync.Mutex
	enabled bool
	sink    io.Writer          // Output destination (stderr or file)
	logger  *lumberjack.Logger // Optional file logger
	filters TraceFilters
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
	Value     string    `json:"value"`    // String representation of evaluated value
	Word      string    `json:"word"`     // Word being evaluated (if applicable)
	Duration  int64     `json:"duration"` // Nanoseconds spent evaluating
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
		found := slices.Contains(ts.filters.IncludeWords, event.Word)
		if !found {
			return
		}
	}

	if slices.Contains(ts.filters.ExcludeWords, event.Word) {
		return
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
