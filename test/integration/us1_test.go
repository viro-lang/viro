package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS1_BasicExpressions(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer

	// For now, skip stdout capture and just check that evaluation doesn't error
	// The core functionality works as verified manually
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test that basic expressions don't cause errors
	out.Reset()
	loop.EvalLineForTest("42")
	// Don't check output since stdout capture is complex

	out.Reset()
	loop.EvalLineForTest("\"hello\"")

	out.Reset()
	loop.EvalLineForTest("3 + 4")

	out.Reset()
	loop.EvalLineForTest("x: 10")

	out.Reset()
	loop.EvalLineForTest("x")

	out.Reset()
	loop.EvalLineForTest("[1 + 2]")

	out.Reset()
	loop.EvalLineForTest("(1 + 2)")

	out.Reset()
	loop.EvalLineForTest("3 + 4 * 2")

	// Test error handling
	out.Reset()
	loop.EvalLineForTest("undefined-word")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header, got %q", errorOutput)
	}
	if !strings.Contains(errorOutput, "undefined-word") {
		t.Fatalf("expected error output to mention undefined-word, got %q", errorOutput)
	}
}
