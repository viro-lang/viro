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

	loop := repl.NewREPLForTest(evaluator, &out)

	// Test that basic expressions evaluate correctly
	out.Reset()
	loop.EvalLineForTest("42")
	result := strings.TrimSpace(out.String())
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("\"hello\"")
	result = strings.TrimSpace(out.String())
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("3 + 4")
	result = strings.TrimSpace(out.String())
	if result != "7" {
		t.Errorf("expected '7', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("x: 10")
	result = strings.TrimSpace(out.String())
	if result != "10" {
		t.Errorf("expected '10', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("x")
	result = strings.TrimSpace(out.String())
	if result != "10" {
		t.Errorf("expected '10', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("[1 + 2]")
	result = strings.TrimSpace(out.String())
	if result != "3" {
		t.Errorf("expected '3', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("(1 + 2)")
	result = strings.TrimSpace(out.String())
	if result != "3" {
		t.Errorf("expected '3', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("3 + 4 * 2")
	result = strings.TrimSpace(out.String())
	if result != "14" {
		t.Errorf("expected '14', got %q", result)
	}

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
