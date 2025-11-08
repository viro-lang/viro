package contract

import (
	"os"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func NewTestEvaluator() *eval.Evaluator {
	// Initialize trace/debug sessions for test isolation
	// Use os.DevNull to avoid trace output pollution during tests
	if err := trace.InitTrace(os.DevNull, 50); err != nil {
		// Disable trace if initialization fails
		if trace.GlobalTraceSession != nil {
			trace.GlobalTraceSession.Disable()
		}
	}
	// Disable trace between tests to ensure clean state
	if trace.GlobalTraceSession != nil {
		trace.GlobalTraceSession.Disable()
	}
	// Always reset debugger to ensure clean state between tests
	debug.InitDebugger()

	// Create evaluator and register all natives
	e := eval.NewEvaluator()
	rootFrame := e.GetFrameByIndex(0)
	native.RegisterMathNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterIONatives(rootFrame, e)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)

	return e
}

// Evaluate is a helper function to evaluate Viro code in tests.
func Evaluate(src string) (core.Value, error) {
	vals, locations, err := parse.ParseWithSource(src, "(test)")
	if err != nil {
		return value.NewNoneVal(), err
	}

	e := NewTestEvaluator()
	result, err := e.DoBlock(vals, locations)
	if err != nil {
		if verr, ok := err.(*verror.Error); ok {
			if verr.Category == verror.ErrThrow {
				if verr.ID == verror.ErrIDBreak {
					return value.NewNoneVal(), verror.NewScriptError(
						verror.ErrIDBreakOutsideLoop,
						[3]string{},
					)
				}
				if verr.ID == verror.ErrIDContinue {
					return value.NewNoneVal(), verror.NewScriptError(
						verror.ErrIDContinueOutsideLoop,
						[3]string{},
					)
				}
			}
		}
	}
	return result, err
}

// RunSeriesTest is a unified test helper for series operations that handles
// common error checking patterns and result validation.
func RunSeriesTest(t *testing.T, input string, want string, wantErr bool, errID string) {
	t.Helper()

	e := NewTestEvaluator()
	tokens, locations, parseErr := parse.ParseWithSource(input, "(test)")
	if parseErr != nil {
		t.Fatalf("Parse error: %v", parseErr)
	}

	result, err := e.DoBlock(tokens, locations)

	if wantErr {
		if err == nil {
			t.Errorf("Expected error with ID %s, got nil", errID)
			return
		}
		if evalErr, ok := err.(*verror.Error); ok {
			if evalErr.ID != errID {
				t.Errorf("Expected error ID %s, got %s", errID, evalErr.ID)
			}
		} else {
			t.Errorf("Expected verror.Error, got %T", err)
		}
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	got := result.Mold()
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}
