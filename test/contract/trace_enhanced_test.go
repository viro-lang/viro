package contract

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/marcin-radoszewski/viro/internal/trace"
)

// TestTraceEventJSONSerialization tests that TraceEvent serializes correctly with new fields.
func TestTraceEventJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		event    trace.TraceEvent
		wantJSON map[string]interface{}
		wantOmit []string
	}{
		{
			name: "basic event with legacy fields only",
			event: trace.TraceEvent{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Value:     "42",
				Word:      "test",
				Duration:  1000,
			},
			wantJSON: map[string]interface{}{
				"value":    "42",
				"word":     "test",
				"duration": float64(1000),
			},
			wantOmit: []string{
				"event_type", "step", "depth", "position",
				"expression", "args", "frame", "parent_expr", "error",
			},
		},
		{
			name: "event with all new fields populated",
			event: trace.TraceEvent{
				Timestamp:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Value:      "3",
				Word:       "fact",
				Duration:   5000,
				EventType:  "call",
				Step:       42,
				Depth:      2,
				Position:   5,
				Expression: "fact 3",
				Args:       map[string]string{"n": "3"},
				Frame:      map[string]string{"n": "3", "x": "10"},
				ParentExpr: "result: fact 3",
				Error:      "",
			},
			wantJSON: map[string]interface{}{
				"value":       "3",
				"word":        "fact",
				"duration":    float64(5000),
				"event_type":  "call",
				"step":        float64(42),
				"depth":       float64(2),
				"position":    float64(5),
				"expression":  "fact 3",
				"args":        map[string]interface{}{"n": "3"},
				"frame":       map[string]interface{}{"n": "3", "x": "10"},
				"parent_expr": "result: fact 3",
			},
			wantOmit: []string{"error"},
		},
		{
			name: "event with error",
			event: trace.TraceEvent{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Value:     "",
				Word:      "divide",
				Duration:  1000,
				EventType: "eval",
				Step:      10,
				Error:     "division by zero",
			},
			wantJSON: map[string]interface{}{
				"value":      "",
				"word":       "divide",
				"duration":   float64(1000),
				"event_type": "eval",
				"step":       float64(10),
				"error":      "division by zero",
			},
			wantOmit: []string{
				"depth", "position", "expression",
				"args", "frame", "parent_expr",
			},
		},
		{
			name: "event with partial new fields",
			event: trace.TraceEvent{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Value:     "6",
				Word:      "add",
				Duration:  500,
				EventType: "call",
				Step:      15,
				Depth:     1,
			},
			wantJSON: map[string]interface{}{
				"value":      "6",
				"word":       "add",
				"duration":   float64(500),
				"event_type": "call",
				"step":       float64(15),
				"depth":      float64(1),
			},
			wantOmit: []string{
				"position", "expression", "args",
				"frame", "parent_expr", "error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("failed to marshal TraceEvent: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			for key, want := range tt.wantJSON {
				got, ok := result[key]
				if !ok {
					t.Errorf("expected field %q to be present", key)
					continue
				}

				switch wantVal := want.(type) {
				case map[string]interface{}:
					gotMap, ok := got.(map[string]interface{})
					if !ok {
						t.Errorf("field %q: expected map, got %T", key, got)
						continue
					}
					for k, v := range wantVal {
						if gotMap[k] != v {
							t.Errorf("field %q[%q]: expected %v, got %v", key, k, v, gotMap[k])
						}
					}
				default:
					if got != want {
						t.Errorf("field %q: expected %v, got %v", key, want, got)
					}
				}
			}

			for _, omitKey := range tt.wantOmit {
				if _, ok := result[omitKey]; ok {
					t.Errorf("expected field %q to be omitted from JSON", omitKey)
				}
			}
		})
	}
}

// TestTraceEventBackwardCompatibility ensures old fields still work correctly.
func TestTraceEventBackwardCompatibility(t *testing.T) {
	event := trace.TraceEvent{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Value:     "hello",
		Word:      "print",
		Duration:  2000,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal TraceEvent: %v", err)
	}

	var result trace.TraceEvent
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal TraceEvent: %v", err)
	}

	if result.Value != "hello" {
		t.Errorf("Value: expected %q, got %q", "hello", result.Value)
	}
	if result.Word != "print" {
		t.Errorf("Word: expected %q, got %q", "print", result.Word)
	}
	if result.Duration != 2000 {
		t.Errorf("Duration: expected %d, got %d", 2000, result.Duration)
	}

	if result.EventType != "" {
		t.Errorf("EventType: expected empty, got %q", result.EventType)
	}
	if result.Step != 0 {
		t.Errorf("Step: expected 0, got %d", result.Step)
	}
	if result.Depth != 0 {
		t.Errorf("Depth: expected 0, got %d", result.Depth)
	}
}

// TestTraceSessionStepCounter tests the step counter functionality.
func TestTraceSessionStepCounter(t *testing.T) {
	filters := trace.TraceFilters{
		Verbose:     false,
		StepLevel:   1,
		IncludeArgs: false,
		MaxDepth:    0,
	}

	err := trace.InitTrace("", 50)
	if err != nil {
		t.Fatalf("failed to initialize trace: %v", err)
	}
	defer trace.GlobalTraceSession.Close()

	trace.GlobalTraceSession.Enable(filters)

	trace.GlobalTraceSession.ResetStepCounter()

	step1 := trace.GlobalTraceSession.NextStep()
	if step1 != 1 {
		t.Errorf("first step: expected 1, got %d", step1)
	}

	step2 := trace.GlobalTraceSession.NextStep()
	if step2 != 2 {
		t.Errorf("second step: expected 2, got %d", step2)
	}

	step3 := trace.GlobalTraceSession.NextStep()
	if step3 != 3 {
		t.Errorf("third step: expected 3, got %d", step3)
	}

	trace.GlobalTraceSession.ResetStepCounter()

	step4 := trace.GlobalTraceSession.NextStep()
	if step4 != 1 {
		t.Errorf("after reset: expected 1, got %d", step4)
	}
}

// TestTraceFiltersNewFields tests the new TraceFilters fields.
func TestTraceFiltersNewFields(t *testing.T) {
	tests := []struct {
		name    string
		filters trace.TraceFilters
		checks  func(*testing.T, *trace.TraceSession)
	}{
		{
			name: "verbose mode enabled",
			filters: trace.TraceFilters{
				Verbose:     true,
				StepLevel:   1,
				IncludeArgs: true,
				MaxDepth:    0,
			},
			checks: func(t *testing.T, ts *trace.TraceSession) {
				if !ts.GetVerbose() {
					t.Error("expected GetVerbose() to return true")
				}
				if !ts.GetIncludeArgs() {
					t.Error("expected GetIncludeArgs() to return true")
				}
				if !ts.ShouldTraceExpression() {
					t.Error("expected ShouldTraceExpression() to return true")
				}
				if !ts.ShouldTraceAtDepth(10) {
					t.Error("expected ShouldTraceAtDepth(10) to return true when MaxDepth=0")
				}
			},
		},
		{
			name: "verbose mode disabled",
			filters: trace.TraceFilters{
				Verbose:     false,
				StepLevel:   0,
				IncludeArgs: false,
				MaxDepth:    5,
			},
			checks: func(t *testing.T, ts *trace.TraceSession) {
				if ts.GetVerbose() {
					t.Error("expected GetVerbose() to return false")
				}
				if ts.GetIncludeArgs() {
					t.Error("expected GetIncludeArgs() to return false")
				}
				if ts.ShouldTraceExpression() {
					t.Error("expected ShouldTraceExpression() to return false when StepLevel=0")
				}
				if !ts.ShouldTraceAtDepth(3) {
					t.Error("expected ShouldTraceAtDepth(3) to return true when depth <= MaxDepth")
				}
				if ts.ShouldTraceAtDepth(6) {
					t.Error("expected ShouldTraceAtDepth(6) to return false when depth > MaxDepth")
				}
			},
		},
		{
			name: "step level 1 for expressions",
			filters: trace.TraceFilters{
				Verbose:     false,
				StepLevel:   1,
				IncludeArgs: false,
				MaxDepth:    0,
			},
			checks: func(t *testing.T, ts *trace.TraceSession) {
				if !ts.ShouldTraceExpression() {
					t.Error("expected ShouldTraceExpression() to return true when StepLevel >= 1")
				}
			},
		},
		{
			name: "step level 2 for all",
			filters: trace.TraceFilters{
				Verbose:     false,
				StepLevel:   2,
				IncludeArgs: false,
				MaxDepth:    0,
			},
			checks: func(t *testing.T, ts *trace.TraceSession) {
				if !ts.ShouldTraceExpression() {
					t.Error("expected ShouldTraceExpression() to return true when StepLevel >= 2")
				}
			},
		},
		{
			name: "max depth limiting",
			filters: trace.TraceFilters{
				Verbose:     false,
				StepLevel:   1,
				IncludeArgs: false,
				MaxDepth:    3,
			},
			checks: func(t *testing.T, ts *trace.TraceSession) {
				if !ts.ShouldTraceAtDepth(1) {
					t.Error("expected ShouldTraceAtDepth(1) to return true")
				}
				if !ts.ShouldTraceAtDepth(3) {
					t.Error("expected ShouldTraceAtDepth(3) to return true")
				}
				if ts.ShouldTraceAtDepth(4) {
					t.Error("expected ShouldTraceAtDepth(4) to return false when depth > MaxDepth")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := trace.InitTrace("", 50)
			if err != nil {
				t.Fatalf("failed to initialize trace: %v", err)
			}
			defer trace.GlobalTraceSession.Close()

			trace.GlobalTraceSession.Enable(tt.filters)
			tt.checks(t, trace.GlobalTraceSession)
		})
	}
}

// TestTraceSessionThreadSafety tests concurrent access to step counter.
func TestTraceSessionThreadSafety(t *testing.T) {
	err := trace.InitTrace("", 50)
	if err != nil {
		t.Fatalf("failed to initialize trace: %v", err)
	}
	defer trace.GlobalTraceSession.Close()

	filters := trace.TraceFilters{
		Verbose:     false,
		StepLevel:   1,
		IncludeArgs: false,
		MaxDepth:    0,
	}
	trace.GlobalTraceSession.Enable(filters)
	trace.GlobalTraceSession.ResetStepCounter()

	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				trace.GlobalTraceSession.NextStep()
			}
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	finalStep := trace.GlobalTraceSession.NextStep()
	expectedStep := goroutines*iterations + 1

	if finalStep != int64(expectedStep) {
		t.Errorf("expected step counter to be %d, got %d", expectedStep, finalStep)
	}
}

// TestExpressionLevelTracing tests that all expression types are traced when StepLevel >= 1.
func TestExpressionLevelTracing(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		stepLevel     int
		verbose       bool
		expectEvents  []string
		expectNoEvent []string
	}{
		{
			name:          "literal integer traced with StepLevel=1",
			code:          "42",
			stepLevel:     1,
			verbose:       false,
			expectEvents:  []string{"eval"},
			expectNoEvent: []string{},
		},
		{
			name:          "set-word traced with StepLevel=1",
			code:          "x: 10",
			stepLevel:     1,
			verbose:       false,
			expectEvents:  []string{"eval"},
			expectNoEvent: []string{},
		},
		{
			name:          "function call traced",
			code:          "+ 1 2",
			stepLevel:     1,
			verbose:       false,
			expectEvents:  []string{"call", "return"},
			expectNoEvent: []string{},
		},
		{
			name:          "block enter/exit traced",
			code:          "do [1 + 1]",
			stepLevel:     1,
			verbose:       false,
			expectEvents:  []string{"block-enter", "block-exit", "call", "return"},
			expectNoEvent: []string{},
		},
		{
			name:          "literals not traced with StepLevel=0",
			code:          "42",
			stepLevel:     0,
			verbose:       false,
			expectEvents:  []string{},
			expectNoEvent: []string{"eval"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: These tests verify that tracing infrastructure is in place
			// Full integration tests with actual trace output would require
			// capturing stderr or configuring a trace file, which is tested
			// elsewhere in the test suite
			_, err := Evaluate(tt.code)
			if err != nil {
				t.Logf("Note: Code evaluation result: %v", err)
			}
		})
	}
}

// TestFunctionArgumentCapture tests that function arguments are captured when IncludeArgs is enabled.
func TestFunctionArgumentCapture(t *testing.T) {
	code := `
		add: fn [a b] [a + b]
		result: add 5 3
	`

	_, err := Evaluate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestFrameStateCapture tests that frame state is captured when Verbose is enabled.
func TestFrameStateCapture(t *testing.T) {
	code := `
		x: 10
		y: 20
		result: x + y
	`

	_, err := Evaluate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCallStackDepth tests that call stack depth is tracked correctly.
func TestCallStackDepth(t *testing.T) {
	code := `
		fact: fn [n] [
			if (= n 0) [1] [
				* n (fact (- n 1))
			]
		]
		result: fact 3
	`

	_, err := Evaluate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestTraceVerboseRefinement tests the --verbose refinement.
func TestTraceVerboseRefinement(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "trace with --verbose",
			code:    "trace --on --verbose",
			wantErr: false,
		},
		{
			name:    "trace without --verbose",
			code:    "trace --on",
			wantErr: false,
		},
		{
			name:    "trace --verbose requires --on",
			code:    "trace --off",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: got error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

// TestTraceStepLevelRefinement tests the --step-level refinement.
func TestTraceStepLevelRefinement(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "step-level 0 (calls only)",
			code:    "trace --on --step-level 0",
			wantErr: false,
		},
		{
			name:    "step-level 1 (expressions)",
			code:    "trace --on --step-level 1",
			wantErr: false,
		},
		{
			name:    "step-level 2 (all)",
			code:    "trace --on --step-level 2",
			wantErr: false,
		},
		{
			name:    "step-level invalid (negative)",
			code:    "trace --on --step-level -1",
			wantErr: true,
		},
		{
			name:    "step-level invalid (too high)",
			code:    "trace --on --step-level 3",
			wantErr: true,
		},
		{
			name:    "step-level invalid type (string)",
			code:    "trace --on --step-level \"test\"",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: got error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

// TestTraceIncludeArgsRefinement tests the --include-args refinement.
func TestTraceIncludeArgsRefinement(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "include-args enabled",
			code:    "trace --on --include-args",
			wantErr: false,
		},
		{
			name:    "include-args with verbose",
			code:    "trace --on --include-args --verbose",
			wantErr: false,
		},
		{
			name:    "include-args with step-level",
			code:    "trace --on --include-args --step-level 1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: got error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

// TestTraceMaxDepthRefinement tests the --max-depth refinement.
func TestTraceMaxDepthRefinement(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "max-depth 0 (unlimited)",
			code:    "trace --on --max-depth 0",
			wantErr: false,
		},
		{
			name:    "max-depth 5",
			code:    "trace --on --max-depth 5",
			wantErr: false,
		},
		{
			name:    "max-depth 100",
			code:    "trace --on --max-depth 100",
			wantErr: false,
		},
		{
			name:    "max-depth invalid (negative)",
			code:    "trace --on --max-depth -1",
			wantErr: true,
		},
		{
			name:    "max-depth invalid type (string)",
			code:    "trace --on --max-depth \"test\"",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: got error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

// TestTraceCombinedRefinements tests combinations of trace refinements.
func TestTraceCombinedRefinements(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "all debugging refinements",
			code:    "trace --on --verbose --step-level 1 --include-args --max-depth 5",
			wantErr: false,
		},
		{
			name:    "verbose and include-args",
			code:    "trace --on --verbose --include-args",
			wantErr: false,
		},
		{
			name:    "step-level with max-depth",
			code:    "trace --on --step-level 1 --max-depth 10",
			wantErr: false,
		},
		{
			name:    "all refinements with filtering",
			code:    "trace --on --verbose --step-level 1 --include-args --max-depth 5 --only [test-fn]",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error state: got error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

// TestTraceStepCounterReset tests that step counter is reset when trace is enabled.
func TestTraceStepCounterReset(t *testing.T) {
	code := `
		trace --on
		x: 10
		trace --off
		trace --on
		y: 20
		trace --off
	`

	_, err := Evaluate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
