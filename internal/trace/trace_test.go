package trace

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInitTrace(t *testing.T) {
	tests := []struct {
		name      string
		traceFile string
		maxSizeMB int
		wantErr   bool
	}{
		{"normal init", "", 50, false},
		{"with output file", "/tmp/test_trace.json", 50, false},
		{"with custom size", "", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state
			GlobalTraceSession = nil

			err := InitTrace(tt.traceFile, tt.maxSizeMB)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitTrace() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && GlobalTraceSession == nil {
				t.Error("InitTrace() should set GlobalTraceSession")
			}
		})
	}
}

func TestInitTraceSilent(t *testing.T) {
	// Reset global state
	GlobalTraceSession = nil

	err := InitTraceSilent()
	if err != nil {
		t.Errorf("InitTraceSilent() error = %v", err)
	}

	if GlobalTraceSession == nil {
		t.Error("InitTraceSilent() should set GlobalTraceSession")
	}

	if GlobalTraceSession.IsEnabled() {
		t.Error("InitTraceSilent() should start with tracing disabled")
	}
}

func TestTraceSession_EnableDisable(t *testing.T) {
	ts := &TraceSession{}
	ts.enabled.Store(false)

	// Initially disabled
	if ts.IsEnabled() {
		t.Error("Expected trace session to be initially disabled")
	}

	// Enable
	ts.Enable(TraceFilters{})
	if !ts.IsEnabled() {
		t.Error("Expected trace session to be enabled after Enable()")
	}

	// Disable
	ts.Disable()
	if ts.IsEnabled() {
		t.Error("Expected trace session to be disabled after Disable()")
	}
}

func TestTraceSession_Emit(t *testing.T) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)
	ts.atomicFilters.Store(&TraceFilters{})

	event := TraceEvent{
		Timestamp: time.Now(),
		Value:     "test",
		Word:      "test",
		EventType: "eval",
		Step:      1,
	}

	ts.Emit(event)

	output := buf.String()
	if !strings.Contains(output, `"value":"test"`) {
		t.Errorf("Emit() should write JSON containing the event value, got: %s", output)
	}

	// Test JSON parsing
	var parsed TraceEvent
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Errorf("Emit() should produce valid JSON, got error: %v", err)
	}
}

func TestTraceSession_Filters(t *testing.T) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)

	tests := []struct {
		name       string
		filters    TraceFilters
		event      TraceEvent
		shouldEmit bool
	}{
		{
			name: "include words - match",
			filters: TraceFilters{
				IncludeWords: []string{"test"},
			},
			event: TraceEvent{
				Word: "test",
			},
			shouldEmit: true,
		},
		{
			name: "include words - no match",
			filters: TraceFilters{
				IncludeWords: []string{"other"},
			},
			event: TraceEvent{
				Word: "test",
			},
			shouldEmit: false,
		},
		{
			name: "exclude words - match",
			filters: TraceFilters{
				ExcludeWords: []string{"test"},
			},
			event: TraceEvent{
				Word: "test",
			},
			shouldEmit: false,
		},
		{
			name: "min duration - below threshold",
			filters: TraceFilters{
				MinDuration: time.Second,
			},
			event: TraceEvent{
				Duration: int64(500 * time.Millisecond),
			},
			shouldEmit: false,
		},
		{
			name: "min duration - above threshold",
			filters: TraceFilters{
				MinDuration: time.Second,
			},
			event: TraceEvent{
				Duration: int64(1500 * time.Millisecond),
			},
			shouldEmit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			ts.atomicFilters.Store(&tt.filters)

			ts.Emit(tt.event)

			emitted := buf.Len() > 0
			if emitted != tt.shouldEmit {
				t.Errorf("Emit() shouldEmit = %v, got %v", tt.shouldEmit, emitted)
			}
		})
	}
}

func TestTraceSession_StepCounter(t *testing.T) {
	ts := &TraceSession{}

	// Initial state
	step1 := ts.NextStep()
	if step1 != 1 {
		t.Errorf("NextStep() first call should return 1, got %d", step1)
	}

	step2 := ts.NextStep()
	if step2 != 2 {
		t.Errorf("NextStep() second call should return 2, got %d", step2)
	}

	// Reset
	ts.ResetStepCounter()
	step3 := ts.NextStep()
	if step3 != 1 {
		t.Errorf("NextStep() after reset should return 1, got %d", step3)
	}
}

func TestTraceSession_Getters(t *testing.T) {
	ts := &TraceSession{}

	filters := TraceFilters{
		Verbose:     true,
		IncludeArgs: true,
		StepLevel:   2,
		MaxDepth:    10,
	}
	ts.atomicFilters.Store(&filters)

	if !ts.GetVerbose() {
		t.Error("GetVerbose() should return true")
	}

	if !ts.GetIncludeArgs() {
		t.Error("GetIncludeArgs() should return true")
	}

	if !ts.ShouldTraceExpression() {
		t.Error("ShouldTraceExpression() should return true for StepLevel >= 1")
	}

	if !ts.ShouldTraceAtDepth(5) {
		t.Error("ShouldTraceAtDepth(5) should return true for depth < MaxDepth")
	}

	if ts.ShouldTraceAtDepth(15) {
		t.Error("ShouldTraceAtDepth(15) should return false for depth > MaxDepth")
	}

	// Test unlimited depth (MaxDepth = 0)
	filters.MaxDepth = 0
	ts.atomicFilters.Store(&filters)
	if !ts.ShouldTraceAtDepth(1000) {
		t.Error("ShouldTraceAtDepth(1000) should return true when MaxDepth = 0")
	}
}

func TestTraceSession_SetCallback(t *testing.T) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)
	ts.atomicFilters.Store(&TraceFilters{})

	var called bool
	var capturedEvent TraceEvent

	callback := func(event TraceEvent) {
		called = true
		capturedEvent = event
	}

	ts.SetCallback(callback)

	event := TraceEvent{
		Value: "test",
		Word:  "test",
	}
	ts.Emit(event)

	if !called {
		t.Error("Callback should have been called")
	}

	if capturedEvent.Value != "test" {
		t.Error("Callback should receive the emitted event")
	}
}

func TestTraceSession_ConcurrentAccess(t *testing.T) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)
	ts.atomicFilters.Store(&TraceFilters{})

	var wg sync.WaitGroup
	numGoroutines := 10
	eventsPerGoroutine := 100

	// Start multiple goroutines emitting events
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := TraceEvent{
					Value:     "test",
					Word:      "test",
					EventType: "eval",
					Step:      int64(id*eventsPerGoroutine + j),
				}
				ts.Emit(event)
			}
		}(i)
	}

	wg.Wait()

	// Count the number of JSON lines in output
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedEvents := numGoroutines * eventsPerGoroutine
	if len(lines) != expectedEvents {
		t.Errorf("Expected %d events, got %d lines of output", expectedEvents, len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var event TraceEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Errorf("Line %d: invalid JSON: %v", i, err)
		}
	}
}

func TestTracePortLifecycle(t *testing.T) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)
	ts.atomicFilters.Store(&TraceFilters{})

	// Save original global session
	original := GlobalTraceSession
	GlobalTraceSession = ts
	defer func() { GlobalTraceSession = original }()

	// Test port operations
	TracePortOpen("file", "/tmp/test.txt")
	TracePortRead("file", "/tmp/test.txt", 100)
	TracePortWrite("file", "/tmp/test.txt", 50)
	TracePortClose("file", "/tmp/test.txt")
	TracePortError("file", "/tmp/test.txt", nil)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 5 {
		t.Errorf("Expected 5 trace events, got %d", len(lines))
	}

	// Verify event types and content
	expectedWords := []string{"open", "read", "write", "close", "error"}
	for i, line := range lines {
		var event TraceEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Errorf("Invalid JSON in line %d: %v", i, err)
			continue
		}

		if event.Word != expectedWords[i] {
			t.Errorf("Event %d: expected word %s, got %s", i, expectedWords[i], event.Word)
		}

		if !strings.Contains(event.Value, "file") {
			t.Errorf("Event %d: expected value to contain 'file', got %s", i, event.Value)
		}
	}
}

func TestTraceObjectOperations(t *testing.T) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)
	ts.atomicFilters.Store(&TraceFilters{})

	// Save original global session
	original := GlobalTraceSession
	GlobalTraceSession = ts
	defer func() { GlobalTraceSession = original }()

	// Test object operations
	TraceObjectCreate(5)
	TraceObjectFieldRead("name", true)
	TraceObjectFieldWrite("name", "test-value")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Errorf("Expected 3 trace events, got %d", len(lines))
	}

	// Verify event types
	expectedWords := []string{"object", "select", "put"}
	for i, line := range lines {
		var event TraceEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Errorf("Invalid JSON in line %d: %v", i, err)
			continue
		}

		if event.Word != expectedWords[i] {
			t.Errorf("Event %d: expected word %s, got %s", i, expectedWords[i], event.Word)
		}
	}
}

func TestTraceEvent_JSONSerialization(t *testing.T) {
	event := TraceEvent{
		Timestamp:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Value:      "test value",
		Word:       "test",
		Duration:   1000000, // 1ms in nanoseconds
		EventType:  "eval",
		Step:       42,
		Depth:      3,
		Position:   10,
		Expression: "test-expression",
		Args: map[string]string{
			"arg1": "value1",
			"arg2": "value2",
		},
		Frame: map[string]string{
			"var1": "val1",
			"var2": "val2",
		},
		Error: "test error",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal TraceEvent: %v", err)
	}

	var unmarshaled TraceEvent
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TraceEvent: %v", err)
	}

	// Compare fields (ignoring timestamp which may lose precision)
	if unmarshaled.Value != event.Value {
		t.Errorf("Value mismatch: got %s, want %s", unmarshaled.Value, event.Value)
	}
	if unmarshaled.Word != event.Word {
		t.Errorf("Word mismatch: got %s, want %s", unmarshaled.Word, event.Word)
	}
	if unmarshaled.Duration != event.Duration {
		t.Errorf("Duration mismatch: got %d, want %d", unmarshaled.Duration, event.Duration)
	}
	if unmarshaled.EventType != event.EventType {
		t.Errorf("EventType mismatch: got %s, want %s", unmarshaled.EventType, event.EventType)
	}
	if unmarshaled.Step != event.Step {
		t.Errorf("Step mismatch: got %d, want %d", unmarshaled.Step, event.Step)
	}
	if unmarshaled.Depth != event.Depth {
		t.Errorf("Depth mismatch: got %d, want %d", unmarshaled.Depth, event.Depth)
	}
	if unmarshaled.Position != event.Position {
		t.Errorf("Position mismatch: got %d, want %d", unmarshaled.Position, event.Position)
	}
	if unmarshaled.Expression != event.Expression {
		t.Errorf("Expression mismatch: got %s, want %s", unmarshaled.Expression, event.Expression)
	}
	if unmarshaled.Error != event.Error {
		t.Errorf("Error mismatch: got %s, want %s", unmarshaled.Error, event.Error)
	}

	// Check maps
	for k, v := range event.Args {
		if unmarshaled.Args[k] != v {
			t.Errorf("Args[%s] mismatch: got %s, want %s", k, unmarshaled.Args[k], v)
		}
	}
	for k, v := range event.Frame {
		if unmarshaled.Frame[k] != v {
			t.Errorf("Frame[%s] mismatch: got %s, want %s", k, unmarshaled.Frame[k], v)
		}
	}
}

func TestTraceSession_Close(t *testing.T) {
	// Test with nil logger
	ts := &TraceSession{
		logger: nil,
	}

	err := ts.Close()
	if err != nil {
		t.Errorf("Close() with nil logger should not error, got %v", err)
	}

	// Test with logger (we can't easily test lumberjack without creating files)
	// This would require integration testing
}

// Benchmark tests
func BenchmarkTraceSession_Emit(b *testing.B) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)
	ts.atomicFilters.Store(&TraceFilters{})

	event := TraceEvent{
		Timestamp: time.Now(),
		Value:     "benchmark test",
		Word:      "bench",
		EventType: "eval",
		Step:      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.Emit(event)
	}
}

func BenchmarkTraceSession_EmitWithFilters(b *testing.B) {
	var buf bytes.Buffer
	ts := &TraceSession{
		sink: &buf,
	}
	ts.enabled.Store(true)
	ts.atomicFilters.Store(&TraceFilters{
		IncludeWords: []string{"bench"},
		Verbose:      true,
	})

	event := TraceEvent{
		Timestamp: time.Now(),
		Value:     "benchmark test",
		Word:      "bench",
		EventType: "eval",
		Step:      1,
		Args: map[string]string{
			"arg1": "value1",
			"arg2": "value2",
		},
		Frame: map[string]string{
			"var1": "val1",
			"var2": "val2",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.Emit(event)
	}
}
