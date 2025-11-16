package contract

import (
	"os"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/bootstrap"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func NewTestEvaluator() *eval.Evaluator {
	if err := trace.InitTrace(os.DevNull, 50); err != nil {
		if trace.GlobalTraceSession != nil {
			trace.GlobalTraceSession.Disable()
		}
	}
	if trace.GlobalTraceSession != nil {
		trace.GlobalTraceSession.Disable()
	}
	debug.InitDebugger()

	evaluator, err := bootstrap.NewEvaluatorWithNatives(nil, nil, nil, false)
	if err != nil {
		panic(err)
	}
	return evaluator
}

func Evaluate(src string) (core.Value, error) {
	vals, locations, err := parse.ParseWithSource(src, "(test)")
	if err != nil {
		return value.NewNoneVal(), err
	}

	e := NewTestEvaluator()
	result, err := e.DoBlock(vals, locations)
	if err != nil {
		if returnSig, ok := err.(*eval.ReturnSignal); ok {
			return returnSig.Value(), nil
		}

		convertedErr := verror.ConvertLoopControlSignal(err)
		if convertedErr != err {
			return value.NewNoneVal(), convertedErr
		}
	}
	return result, err
}

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
