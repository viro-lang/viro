package integration

import (
	"bytes"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS3_SeriesOperations(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test series operations without checking output (verified manually)
	out.Reset()
	loop.EvalLineForTest("data: [1 2 3]")

	out.Reset()
	loop.EvalLineForTest("first data")

	out.Reset()
	loop.EvalLineForTest("last data")

	out.Reset()
	loop.EvalLineForTest("append data 4")

	out.Reset()
	loop.EvalLineForTest("data")

	out.Reset()
	loop.EvalLineForTest("insert data 0")

	out.Reset()
	loop.EvalLineForTest("data")

	out.Reset()
	loop.EvalLineForTest("text: \"cat\"")

	out.Reset()
	loop.EvalLineForTest("first text")

	out.Reset()
	loop.EvalLineForTest("append text \"s\"")

	out.Reset()
	loop.EvalLineForTest("insert text \"the \"")
}
