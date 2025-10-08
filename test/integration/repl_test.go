package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestREPL_ErrorRecovery(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	loop.EvalLineForTest("square: fn [n] [n * n]")
	output := out.String()
	if !strings.Contains(output, "function") {
		t.Fatalf("expected function definition output, got %q", output)
	}

	out.Reset()

	loop.EvalLineForTest("square \"oops\"")
	output = out.String()
	if !strings.Contains(output, "** Script Error") {
		t.Fatalf("expected script error header, got %q", output)
	}
	if !strings.Contains(output, "square") {
		t.Fatalf("expected call stack or message to mention square, got %q", output)
	}

	out.Reset()

	loop.EvalLineForTest("square 4")
	output = out.String()
	if !strings.Contains(output, "16") {
		t.Fatalf("expected successful evaluation after error, got %q", output)
	}
}
