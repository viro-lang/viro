package integration

import (
	"os"

	"github.com/marcin-radoszewski/viro/internal/bootstrap"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/trace"
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
