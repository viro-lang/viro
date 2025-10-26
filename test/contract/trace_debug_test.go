package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
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

// T135: debug stepping (--step, --next, --finish, --continue)
func TestDebugStepping(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "debug --continue",
			code:    "debug --on\ndebug --continue",
			wantErr: false,
		},
		{
			name:    "debug --step",
			code:    "debug --on\ndebug --step",
			wantErr: false,
		},
		{
			name:    "debug --next",
			code:    "debug --on\ndebug --next",
			wantErr: false,
		},
		{
			name:    "debug --finish",
			code:    "debug --on\ndebug --finish",
			wantErr: false,
		},
		{
			name:    "stepping when debugger disabled",
			code:    "debug --step",
			wantErr: true, // Should fail if debugger not enabled
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
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// T136: debug --locals/--stack
func TestDebugInspection(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectType core.ValueType
		wantErr    bool
	}{
		{
			name:       "debug --locals returns object",
			code:       "debug --on\nx: 10\ny: 20\ndebug --locals",
			expectType: value.TypeObject, // Should return object with local bindings
			wantErr:    false,
		},
		{
			name:       "debug --stack returns block",
			code:       "debug --on\nsquare: fn [x] [x * x]\ndebug --stack",
			expectType: value.TypeBlock, // Should return block of stack frames
			wantErr:    false,
		},
		{
			name:       "debug --locals in function context",
			code:       "debug --on\ntest: fn [a b] [c: a + b debug --locals]\ntest 5 10",
			expectType: value.TypeObject,
			wantErr:    false,
		},
		{
			name:    "debug --locals when disabled",
			code:    "debug --locals",
			wantErr: true, // Should fail if debugger not enabled
		},
		{
			name:    "debug --stack when disabled",
			code:    "debug --stack",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if vErr, ok := err.(*verror.Error); ok {
					if vErr.Category != verror.ErrScript {
						t.Errorf("expected Script error, got %v", vErr.Category)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != tt.expectType {
				t.Errorf("expected type %v, got %v", tt.expectType, result.GetType())
			}
		})
	}
}
