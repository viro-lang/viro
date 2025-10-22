package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS5_ErrorScenarios(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	out.Reset()
	loop.EvalLineForTest("undefined-word")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "undefined-word") {
		t.Fatalf("expected undefined word mentioned, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "no value for word") {
		t.Fatalf("expected no value message, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("10 / 0")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Math Error") {
		t.Fatalf("expected math error header, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "division by zero") {
		t.Fatalf("expected division by zero message, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("+ \"str\" 5")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header for type mismatch, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "type mismatch") {
		t.Fatalf("expected type mismatch message, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "string") || !strings.Contains(errorOutput, "str") {
		t.Fatalf("expected offending value information in output, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("@")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Syntax Error") {
		t.Fatalf("expected syntax error header for malformed input, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "invalid syntax") && !strings.Contains(strings.ToLower(errorOutput), "unexpected") {
		t.Fatalf("expected syntax error details, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("square: fn [n] [n * n]")
	loop.EvalLineForTest("square")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header for arg count, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "expected 1, got 0") {
		t.Fatalf("expected argument count details, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "square") {
		t.Fatalf("expected function name in call stack, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("inner: fn [y] [y + missing]")
	loop.EvalLineForTest("outer: fn [n] [inner n]")
	loop.EvalLineForTest("outer 5")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header for call stack propagation, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "missing") {
		t.Fatalf("expected missing word mentioned, got %q", errorOutput)
	}
	if !strings.Contains(errorOutput, "inner") || !strings.Contains(errorOutput, "outer") {
		t.Fatalf("expected call stack frames inner and outer, got %q", errorOutput)
	}

	// Test that REPL still works after errors
	out.Reset()
	loop.EvalLineForTest("1 + 1")
	result := strings.TrimSpace(out.String())
	if result != "2" {
		t.Errorf("expected '2', got %q", result)
	}
}
