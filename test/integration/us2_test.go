package integration

import (
	"io"
	"os"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func captureEvalOutput(t *testing.T, e *eval.Evaluator, script string) (string, value.Value, *verror.Error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w

	vals, parseErr := parse.Parse(script)
	if parseErr != nil {
		t.Fatalf("Parse failed for %q: %v", script, parseErr)
	}

	result, evalErr := e.Do_Blk(vals)

	if err := w.Close(); err != nil {
		t.Fatalf("closing stdout writer failed: %v", err)
	}
	os.Stdout = oldStdout

	data, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("reading captured stdout failed: %v", readErr)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("closing stdout reader failed: %v", err)
	}

	return string(data), result, evalErr
}

func runScript(t *testing.T, e *eval.Evaluator, script string) (value.Value, *verror.Error) {
	t.Helper()
	vals, parseErr := parse.Parse(script)
	if parseErr != nil {
		t.Fatalf("Parse failed for %q: %v", script, parseErr)
	}
	return e.Do_Blk(vals)
}

func TestUS2_ControlFlowScenarios(t *testing.T) {
	e := NewTestEvaluator()

	output, result, err := captureEvalOutput(t, e, "when true [print \"yes\"]")
	if err != nil {
		t.Fatalf("when true should succeed, got error: %v", err)
	}
	if output != "yes\n" {
		t.Fatalf("expected print output 'yes', got %q", output)
	}
	if !result.Equals(value.NoneVal()) {
		t.Fatalf("when should return none, got %v", result)
	}

	output, result, err = captureEvalOutput(t, e, "when false [print \"yes\"]")
	if err != nil {
		t.Fatalf("when false should succeed, got error: %v", err)
	}
	if output != "" {
		t.Fatalf("when false should not emit output, got %q", output)
	}
	if !result.Equals(value.NoneVal()) {
		t.Fatalf("when false should return none, got %v", result)
	}

	val, err := runScript(t, e, "if 1 < 2 [\"less\"] [\"greater\"]")
	if err != nil {
		t.Fatalf("if true branch failed: %v", err)
	}
	if !val.Equals(value.StrVal("less")) {
		t.Fatalf("expected if true branch to return \"less\", got %v", val)
	}

	val, err = runScript(t, e, "if false [\"yes\"] [\"no\"]")
	if err != nil {
		t.Fatalf("if false branch failed: %v", err)
	}
	if !val.Equals(value.StrVal("no")) {
		t.Fatalf("expected if false branch to return \"no\", got %v", val)
	}

	output, result, err = captureEvalOutput(t, e, "loop 3 [print \"hi\"]")
	if err != nil {
		t.Fatalf("loop should succeed, got error: %v", err)
	}
	if output != "hi\nhi\nhi\n" {
		t.Fatalf("loop print expected three lines of hi, got %q", output)
	}
	if !result.Equals(value.NoneVal()) {
		t.Fatalf("loop with print should return none, got %v", result)
	}

	val, err = runScript(t, e, "when true [if true [\"nested\"] [\"no\"]]")
	if err != nil {
		t.Fatalf("nested control flow failed: %v", err)
	}
	if !val.Equals(value.StrVal("nested")) {
		t.Fatalf("expected nested control flow to return \"nested\", got %v", val)
	}

	val, err = runScript(t, e, "when false [if true [\"should-not-run\"] [\"x\"]]")
	if err != nil {
		t.Fatalf("when false with nested block should succeed: %v", err)
	}
	if !val.Equals(value.NoneVal()) {
		t.Fatalf("when false should return none even with nested block, got %v", val)
	}
}
