package contract

import (
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Evaluate is a helper function to evaluate Viro code in tests.
func Evaluate(src string) (value.Value, *verror.Error) {
	// Initialize trace and debug sessions (Feature 002)
	// This is needed for tests that use trace/debug natives
	// Reset state for each test to ensure isolation
	if trace.GlobalTraceSession == nil {
		_ = trace.InitTrace("", 50)
	}
	// Disable trace between tests to ensure clean state
	if trace.GlobalTraceSession != nil {
		trace.GlobalTraceSession.Disable()
	}
	// Always reset debugger to ensure clean state between tests
	debug.InitDebugger()

	vals, err := parse.Parse(src)
	if err != nil {
		return value.NoneVal(), err
	}

	e := eval.NewEvaluator()
	return e.Do_Blk(vals)
}
