package contract

import (
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Test suite for Feature 002: Trace and debug capabilities
// Contract tests validate FR-020 and FR-021 requirements

// T131: trace --on/--off/trace?
func TestTraceControls(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectType core.ValueType
		checkFunc  func(*testing.T, core.Value)
		wantErr    bool
	}{
		{
			name:       "enable tracing with trace --on",
			code:       "trace --on",
			expectType: value.TypeNone, // trace --on returns none
			wantErr:    false,
		},
		{
			name:       "disable tracing with trace --off",
			code:       "trace --on\ntrace --off",
			expectType: value.TypeNone,
			wantErr:    false,
		},
		{
			name:       "query trace status with trace?",
			code:       "trace --on\ntrace?",
			expectType: value.TypeLogic, // Returns boolean indicating trace state
			checkFunc: func(t *testing.T, v core.Value) {
				enabled, ok := value.AsLogicValue(v)
				if !ok {
					t.Fatal("expected trace? to return boolean!")
				}
				if !enabled {
					t.Error("expected trace? to return true when enabled")
				}
			},
			wantErr: false,
		},
		{
			name:       "trace? when disabled",
			code:       "trace?",
			expectType: value.TypeLogic,
			checkFunc: func(t *testing.T, v core.Value) {
				// Should return false when disabled
				enabled, ok := value.AsLogicValue(v)
				if !ok {
					t.Fatal("expected trace? to return boolean!")
				}
				if enabled {
					t.Error("expected trace? to return false when disabled")
				}
			},
			wantErr: false,
		},
		{
			name:       "enable and re-enable trace",
			code:       "trace --on\ntrace --on",
			expectType: value.TypeNone,
			wantErr:    false, // Re-enabling should be idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != tt.expectType {
				t.Errorf("expected type %v, got %v", tt.expectType, result.GetType())
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// T132: trace filtering (--only, --exclude)
func TestTraceFiltering(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
		errCat  verror.ErrorCategory
	}{
		{
			name:    "trace with --only filter",
			code:    "trace --on --only [calculate-interest]",
			wantErr: false,
		},
		{
			name:    "trace with --exclude filter",
			code:    "trace --on --exclude [debug-helper]",
			wantErr: false,
		},
		{
			name:    "trace with both filters",
			code:    "trace --on --only [func1 func2] --exclude [helper1]",
			wantErr: false,
		},
		{
			name:    "empty --only filter (include all)",
			code:    "trace --on --only []",
			wantErr: false,
		},
		{
			name:    "invalid --only value (not a block)",
			code:    "trace --on --only 123",
			wantErr: true,
			errCat:  verror.ErrScript,
		},
		{
			name:    "invalid --exclude value (not a block)",
			code:    "trace --on --exclude \"test\"",
			wantErr: true,
			errCat:  verror.ErrScript,
		},
		{
			name:    "--only block with non-word entries",
			code:    "trace --on --only [func1 123 func2]",
			wantErr: true,
			errCat:  verror.ErrScript,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if vErr, ok := err.(*verror.Error); ok {
					if vErr.Category != tt.errCat {
						t.Errorf("expected error category %v, got %v", tt.errCat, vErr.Category)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// T133: trace sink configuration (--file, --append)
// NOTE: File type (%) is not yet implemented, so file-related tests are skipped
func TestTraceSinkConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
		errCat  verror.ErrorCategory
	}{
		// File-related tests removed until file type (%) is implemented
		{
			name:    "trace with custom file path",
			code:    "trace --on --file \"trace-custom.log\"",
			wantErr: false,
		},
		{
			name:    "trace with append mode",
			code:    "trace --on --file \"trace-custom.log\" --append",
			wantErr: false,
		},
		{
			name:    "trace without file path (default)",
			code:    "trace --on",
			wantErr: false,
		},
		{
			name:    "trace with path outside sandbox",
			code:    "trace --on --file \"../../etc/passwd\"",
			wantErr: true,
			errCat:  verror.ErrAccess, // Sandbox violation
		},
		{
			name:    "trace with absolute path outside sandbox",
			code:    "trace --on --file \"/tmp/evil.log\"",
			wantErr: true,
			errCat:  verror.ErrAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if vErr, ok := err.(*verror.Error); ok {
					if vErr.Category != tt.errCat {
						t.Errorf("expected error category %v, got %v", tt.errCat, vErr.Category)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// T134: debug --breakpoint/--remove
func TestDebugBreakpoints(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "enable debugger",
			code:    "debug --on",
			wantErr: false,
		},
		{
			name:    "disable debugger",
			code:    "debug --on\ndebug --off",
			wantErr: false,
		},
		{
			name:    "set breakpoint on word",
			code:    "debug --on\nsquare: fn [x] [x * x]\ndebug --breakpoint 'square",
			wantErr: false,
		},
		{
			name:    "set breakpoint with index",
			code:    "debug --on\nfunc: fn [x] [print x print x + 1]\ndebug --breakpoint 'func 2",
			wantErr: false,
		},
		{
			name:    "remove breakpoint by ID",
			code:    "debug --on\nsquare: fn [x] [x * x]\nid: debug --breakpoint 'square\ndebug --remove id",
			wantErr: false,
		},
		{
			name:    "add breakpoint when debugger disabled",
			code:    "square: fn [x] [x * x]\ndebug --breakpoint 'square",
			wantErr: true,
			errMsg:  "debugger", // Error should mention debugger not enabled
		},
		{
			name:    "breakpoint on unknown word",
			code:    "debug --on\ndebug --breakpoint 'nonexistent",
			wantErr: true,
			errMsg:  "unknown", // Error should mention unknown symbol
		},
		{
			name:    "remove nonexistent breakpoint",
			code:    "debug --on\ndebug --remove 99999",
			wantErr: true,
			errMsg:  "breakpoint", // Error should mention no such breakpoint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if vErr, ok := err.(*verror.Error); ok {
					if vErr.Category != verror.ErrScript {
						t.Errorf("expected Script error, got %v", vErr.Category)
					}
					if tt.errMsg != "" && !strings.Contains(strings.ToLower(vErr.Message), tt.errMsg) {
						t.Errorf("expected error message to contain %q, got %q", tt.errMsg, vErr.Message)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// captureTraceEvents captures trace events emitted during test execution.
func captureTraceEvents(t *testing.T, action func()) []trace.TraceEvent {
	var events []trace.TraceEvent
	var mu sync.Mutex

	err := trace.InitTraceSilent()
	if err != nil {
		t.Fatalf("failed to initialize trace: %v", err)
	}
	defer trace.GlobalTraceSession.Close()

	trace.GlobalTraceSession.SetCallback(func(event trace.TraceEvent) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
	})

	// Enable the trace session for capturing events
	trace.GlobalTraceSession.Enable(trace.TraceFilters{})

	action()

	return events
}

// TestProbeNative tests the probe debug native function.
func TestProbeNative(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectOutput string
		expectResult string
		quiet        bool
	}{
		{
			name:         "probe with integer",
			input:        "42",
			expectOutput: "== 42\n",
			expectResult: "42",
			quiet:        false,
		},
		{
			name:         "probe with string",
			input:        `"hello world"`,
			expectOutput: `== "hello world"` + "\n",
			expectResult: `"hello world"`,
			quiet:        false,
		},
		{
			name:         "probe with block",
			input:        "[1 2 3]",
			expectOutput: "== [1 2 3]\n",
			expectResult: "[1 2 3]",
			quiet:        false,
		},
		{
			name:         "probe with integer in quiet mode",
			input:        "42",
			expectOutput: "", // No stdout output in quiet mode
			expectResult: "42",
			quiet:        true,
		},
		{
			name:         "probe with string in quiet mode",
			input:        `"hello world"`,
			expectOutput: "",
			expectResult: `"hello world"`,
			quiet:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the probe expression
			script := "probe " + tt.input
			vals, locations, perr := parse.ParseWithSource(script, "(test)")
			if perr != nil {
				t.Fatalf("Parse failed: %v", perr)
			}

			// Create evaluator
			e := NewTestEvaluator()

			var captured string
			var result core.Value
			var err error

			if tt.quiet {
				// In quiet mode, set output writer to io.Discard and don't capture output
				e.SetOutputWriter(io.Discard)
				result, err = e.DoBlock(vals, locations)
				captured = "" // Expect no output
			} else {
				// In normal mode, capture output
				captured, result, err = captureOutput(t, e, func() (core.Value, error) {
					val, derr := e.DoBlock(vals, locations)
					if derr != nil {
						return value.NewNoneVal(), derr
					}
					return val, nil
				})
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check result value
			if result.Mold() != tt.expectResult {
				t.Errorf("expected result %s, got %s", tt.expectResult, result.Mold())
			}

			// Check stdout output
			if captured != tt.expectOutput {
				t.Errorf("expected stdout %q, got %q", tt.expectOutput, captured)
			}
		})
	}
}

// TestProbeFallback tests that probe falls back to error writer when trace is unavailable.
func TestProbeFallback(t *testing.T) {
	// Parse the probe expression
	script := "probe 42"
	vals, locations, perr := parse.ParseWithSource(script, "(test)")
	if perr != nil {
		t.Fatalf("Parse failed: %v", perr)
	}

	e := NewTestEvaluator()
	e.SetOutputWriter(io.Discard) // Quiet mode

	// Save and restore global trace session to avoid leaking state
	oldSession := trace.GlobalTraceSession
	defer func() {
		trace.GlobalTraceSession = oldSession
	}()

	// Ensure no trace session is available
	trace.GlobalTraceSession = nil

	// Capture error output
	var errorOutput strings.Builder
	e.SetErrorWriter(&errorOutput)

	_, err := e.DoBlock(vals, locations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output went to error writer
	expected := "== 42\n"
	if errorOutput.String() != expected {
		t.Errorf("expected error output %q, got %q", expected, errorOutput.String())
	}
}

// TestProbeTraceFallback tests that probe emits a trace event when stdout is suppressed but trace session is active.
func TestProbeTraceFallback(t *testing.T) {
	// Parse the probe expression
	script := "probe 42"
	vals, locations, perr := parse.ParseWithSource(script, "(test)")
	if perr != nil {
		t.Fatalf("Parse failed: %v", perr)
	}

	// Create evaluator first (this will initialize and disable trace)
	e := NewTestEvaluator()
	e.SetOutputWriter(io.Discard) // Suppress stdout

	// Now set up our trace session (this overrides the one from NewTestEvaluator)
	var events []trace.TraceEvent
	var mu sync.Mutex

	traceInitErr := trace.InitTraceSilent()
	if traceInitErr != nil {
		t.Fatalf("failed to initialize trace: %v", traceInitErr)
	}
	defer trace.GlobalTraceSession.Close()

	trace.GlobalTraceSession.SetCallback(func(event trace.TraceEvent) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
	})

	// Enable the trace session
	trace.GlobalTraceSession.Enable(trace.TraceFilters{})

	// Execute the probe
	result, err := e.DoBlock(vals, locations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the result is correct
	if result.Mold() != "42" {
		t.Errorf("expected result '42', got %q", result.Mold())
	}

	// Verify trace event was emitted
	if len(events) != 1 {
		t.Fatalf("expected 1 trace event, got %d", len(events))
	}

	event := events[0]
	if event.Word != "probe" {
		t.Errorf("expected event.Word to be 'probe', got %q", event.Word)
	}
	if event.EventType != "debug" {
		t.Errorf("expected event.EventType to be 'debug', got %q", event.EventType)
	}
	if event.Value != "42" {
		t.Errorf("expected event.Value to be '42', got %q", event.Value)
	}
}

// TestProbeArityError tests that probe returns arity error with no arguments.
func TestProbeArityError(t *testing.T) {
	script := "probe"
	vals, locations, perr := parse.ParseWithSource(script, "(test)")
	if perr != nil {
		t.Fatalf("Parse failed: %v", perr)
	}

	e := NewTestEvaluator()
	_, err := e.DoBlock(vals, locations)

	if err == nil {
		t.Fatal("expected arity error, got nil")
	}

	verr, ok := err.(*verror.Error)
	if !ok {
		t.Fatalf("expected *verror.Error, got %T", err)
	}

	if verr.Category != verror.ErrScript {
		t.Errorf("expected ErrScript category, got %v", verr.Category)
	}

	if verr.ID != verror.ErrIDArgCount {
		t.Errorf("expected %q ID, got %q", verror.ErrIDArgCount, verr.ID)
	}
}

// TestProbeIdentityPreservation tests that probe returns the same object reference.
func TestProbeIdentityPreservation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"block identity", "[1 2 3]"},
		{"object identity", "make object! [a: 1 b: 2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and probe the value in one script
			script := "original: " + tt.input + " probe original"
			vals, locations, perr := parse.ParseWithSource(script, "(test)")
			if perr != nil {
				t.Fatalf("Parse failed: %v", perr)
			}

			e := NewTestEvaluator()
			probed, err := e.DoBlock(vals, locations)
			if err != nil {
				t.Fatalf("failed to probe value: %v", err)
			}

			// Get the original value from the frame
			rootFrame := e.GetFrameByIndex(0)
			original, found := rootFrame.Get("original")
			if !found {
				t.Fatal("original value not found in frame")
			}

			// Verify they are the same reference
			if original != probed {
				t.Error("expected probed value to be the same reference as original")
			}

			// Mutate the probed value and verify the original is also mutated
			// (This tests that they share the same underlying reference)
			switch original.GetType() {
			case value.TypeBlock:
				// For blocks, append an element
				block, ok := original.(*value.BlockValue)
				if !ok {
					t.Fatalf("expected *BlockValue, got %T", original)
				}
				block.Append(value.NewIntVal(999))
				if len(block.Elements) != 4 {
					t.Errorf("expected block length 4 after mutation, got %d", len(block.Elements))
				}
			case value.TypeObject:
				// For objects, set a new field
				obj, ok := value.AsObject(original)
				if !ok {
					t.Fatalf("expected ObjectInstance, got %T", original)
				}
				obj.SetField("mutated", value.NewStrVal("yes"))
				if val, exists := obj.GetField("mutated"); !exists || val.Mold() != `"yes"` {
					t.Error("expected object to have mutated field after setting on probed value")
				}
			}
		})
	}
}
