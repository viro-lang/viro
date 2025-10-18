package integration

import (
	"os"

	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/trace"
)

// NewTestEvaluator creates a fully configured evaluator for testing.
// It initializes trace/debug sessions and registers all native functions.
// Use this instead of NewTestEvaluator() directly in tests to ensure
// consistent setup and proper native registration.
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
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterIONatives(rootFrame, e)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)

	return e
}
