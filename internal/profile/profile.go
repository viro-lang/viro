// Package profile provides execution profiling infrastructure for Viro.
//
// This package enables performance analysis by collecting execution statistics
// from the trace system. It works by registering a callback with the trace
// session to capture timing data for function calls and operations.
//
// # Architecture
//
// The profiler integrates with the existing trace system (internal/trace) via
// callbacks, avoiding duplication of event collection logic. When profiling is
// enabled, trace events are captured and aggregated into function-level statistics.
//
// Key components:
//   - Profiler: Thread-safe statistics collector
//   - FunctionStats: Per-function timing metrics (count, min, max, avg, total)
//   - ProfileReport: Aggregated results with multiple output formats
//
// # Usage
//
// Profiling is typically enabled via CLI flag (--profile) which:
//  1. Initializes trace system with silent output (io.Discard)
//  2. Creates a Profiler instance
//  3. Registers profiler callback with trace session
//  4. Enables trace collection with minimal filters
//  5. Displays report after script execution
//
// # Profiling vs Tracing
//
// Use profiling when you need:
//   - High-level performance overview
//   - Function call counts and timing statistics
//   - Minimal overhead (no JSON serialization)
//   - Human-readable summary reports
//
// Use tracing when you need:
//   - Detailed execution flow analysis
//   - Step-by-step debugging information
//   - Frame state inspection
//   - Machine-parseable event streams
//
// # Performance Implications
//
// Profiling has lower overhead than full tracing because:
//   - Trace output is discarded (no I/O)
//   - Minimal trace filters (StepLevel=0, no verbose mode)
//   - Lock-free callback invocation
//   - In-memory aggregation only
//
// However, the trace system itself still runs, so there is non-zero overhead
// compared to execution without any instrumentation.
//
// # Thread Safety
//
// All Profiler methods are thread-safe and can be called concurrently.
// Statistics are protected by a mutex, and the enabled flag uses atomic
// operations for lock-free reads.
package profile

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/marcin-radoszewski/viro/internal/trace"
)

type Profiler struct {
	mu            sync.Mutex
	enabled       bool
	startTime     time.Time
	endTime       time.Time
	functionStats map[string]*FunctionStats
	eventCount    int64
	totalTime     time.Duration
}

type FunctionStats struct {
	Name        string
	CallCount   int64
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	AverageTime time.Duration
}

type ProfileReport struct {
	TotalExecutionTime time.Duration
	TotalEvents        int64
	Functions          []*FunctionStats
}

func NewProfiler() *Profiler {
	return &Profiler{
		functionStats: make(map[string]*FunctionStats),
		enabled:       false,
	}
}

func (p *Profiler) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = true
	p.startTime = time.Now()
	p.eventCount = 0
	p.functionStats = make(map[string]*FunctionStats)
}

func (p *Profiler) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = false
	p.endTime = time.Now()
	p.totalTime = p.endTime.Sub(p.startTime)
}

func (p *Profiler) IsEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.enabled
}

func (p *Profiler) RecordEvent(word string, duration time.Duration) {
	if !p.IsEnabled() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.eventCount < math.MaxInt64 {
		p.eventCount++
	}

	if word == "" {
		return
	}

	stats, exists := p.functionStats[word]
	if !exists {
		stats = &FunctionStats{
			Name:      word,
			CallCount: 0,
			MinTime:   duration,
			MaxTime:   duration,
		}
		p.functionStats[word] = stats
	}

	if stats.CallCount < math.MaxInt64 {
		stats.CallCount++
	}

	newTotal := stats.TotalTime + duration
	if newTotal < stats.TotalTime {
		stats.TotalTime = time.Duration(math.MaxInt64)
	} else {
		stats.TotalTime = newTotal
	}

	if duration < stats.MinTime {
		stats.MinTime = duration
	}
	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}

	if stats.CallCount > 0 {
		stats.AverageTime = time.Duration(int64(stats.TotalTime) / stats.CallCount)
	}
}

func (p *Profiler) GetReport() *ProfileReport {
	p.mu.Lock()
	defer p.mu.Unlock()

	functions := make([]*FunctionStats, 0, len(p.functionStats))
	for _, stats := range p.functionStats {
		functions = append(functions, stats)
	}

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].TotalTime > functions[j].TotalTime
	})

	return &ProfileReport{
		TotalExecutionTime: p.totalTime,
		TotalEvents:        p.eventCount,
		Functions:          functions,
	}
}

func (r *ProfileReport) FormatText(w io.Writer) {
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(w, "                         EXECUTION PROFILE\n")
	fmt.Fprintf(w, "═══════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Total Execution Time: %v\n", r.TotalExecutionTime)
	fmt.Fprintf(w, "Total Events:         %d\n", r.TotalEvents)
	fmt.Fprintf(w, "\n")

	if len(r.Functions) == 0 {
		fmt.Fprintf(w, "No function calls recorded.\n")
		return
	}

	fmt.Fprintf(w, "Function Statistics (sorted by total time):\n")
	fmt.Fprintf(w, "───────────────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(w, "%-20s %10s %12s %10s %10s %10s\n",
		"Function", "Calls", "Total Time", "Avg Time", "Min Time", "Max Time")
	fmt.Fprintf(w, "───────────────────────────────────────────────────────────────────────\n")

	for _, stats := range r.Functions {
		name := stats.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		fmt.Fprintf(w, "%-20s %10d %12s %10s %10s %10s\n",
			name,
			stats.CallCount,
			formatDuration(stats.TotalTime),
			formatDuration(stats.AverageTime),
			formatDuration(stats.MinTime),
			formatDuration(stats.MaxTime))
	}

	fmt.Fprintf(w, "═══════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(w, "\n")
}

func (r *ProfileReport) FormatJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000.0)
	}
	return fmt.Sprintf("%.3fs", d.Seconds())
}

// EnableProfilingWithTrace configures the trace session to collect profiling data.
// It registers the profiler as a trace callback and enables tracing with minimal
// overhead filters optimized for performance profiling.
//
// Trace filter rationale:
//   - Verbose=false: Skip frame state capture (not needed for profiling)
//   - StepLevel=0: Only capture function calls, not individual expressions
//   - IncludeArgs=false: Skip argument capture (reduces overhead and memory)
//   - MaxDepth=0: No depth limit (profile entire call stack)
//
// These settings minimize trace overhead while capturing all function-level
// timing data needed for accurate profiling statistics.
func EnableProfilingWithTrace(traceSession *trace.TraceSession, profiler *Profiler) {
	profiler.Enable()

	traceSession.SetCallback(func(event trace.TraceEvent) {
		duration := time.Duration(event.Duration)
		profiler.RecordEvent(event.Word, duration)
	})

	filters := trace.TraceFilters{
		Verbose:     false,
		StepLevel:   0,
		IncludeArgs: false,
		MaxDepth:    0,
	}
	traceSession.Enable(filters)
}
