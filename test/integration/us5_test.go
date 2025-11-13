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
	var errorOutputLower string

	out.Reset()
	loop.EvalLineForTest("undefined-word")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header, got %q", errorOutput)
	}
	errorOutputLower = strings.ToLower(errorOutput)
	if !strings.Contains(errorOutputLower, "undefined-word") {
		t.Fatalf("expected undefined word mentioned, got %q", errorOutput)
	}
	if !strings.Contains(errorOutputLower, "no value for word") {
		t.Fatalf("expected no value message, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("10 / 0")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Math Error") {
		t.Fatalf("expected math error header, got %q", errorOutput)
	}
	errorOutputLower = strings.ToLower(errorOutput)
	if !strings.Contains(errorOutputLower, "division by zero") {
		t.Fatalf("expected division by zero message, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("+ \"str\" 5")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header for type mismatch, got %q", errorOutput)
	}
	errorOutputLower = strings.ToLower(errorOutput)
	if !strings.Contains(errorOutputLower, "type mismatch") {
		t.Fatalf("expected type mismatch message, got %q", errorOutput)
	}
	if !strings.Contains(errorOutputLower, "string") || !strings.Contains(errorOutput, "str") {
		t.Fatalf("expected offending value information in output, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("@")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Syntax Error") {
		t.Fatalf("expected syntax error header for malformed input, got %q", errorOutput)
	}
	errorOutputLower = strings.ToLower(errorOutput)
	if !strings.Contains(errorOutputLower, "invalid syntax") &&
		!strings.Contains(errorOutputLower, "unexpected") &&
		!strings.Contains(errorOutputLower, "invalid character") {
		t.Fatalf("expected syntax error details, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("square: fn [n] [n * n]")
	loop.EvalLineForTest("square")
	errorOutput = out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header for arg count, got %q", errorOutput)
	}
	errorOutputLower = strings.ToLower(errorOutput)
	if !strings.Contains(errorOutputLower, "expected 1, got 0") {
		t.Fatalf("expected argument count details, got %q", errorOutput)
	}
	if !strings.Contains(errorOutputLower, "square") {
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
	errorOutputLower = strings.ToLower(errorOutput)
	if !strings.Contains(errorOutputLower, "missing") {
		t.Fatalf("expected missing word mentioned, got %q", errorOutput)
	}
	if !strings.Contains(errorOutput, "inner") || !strings.Contains(errorOutput, "outer") {
		t.Fatalf("expected call stack frames inner and outer, got %q", errorOutput)
	}

	out.Reset()
	loop.EvalLineForTest("1 + 1")
	result := strings.TrimSpace(out.String())
	if result != "2" {
		t.Errorf("expected '2', got %q", result)
	}
}
