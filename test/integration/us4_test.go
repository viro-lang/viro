package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS4_FunctionDefinitionsAndCalls(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test function definitions and calls
	out.Reset()
	loop.EvalLineForTest("square: fn [n] [(* n n)]")
	result := strings.TrimSpace(out.String())
	if result != "function[square]" {
		t.Errorf("expected 'function[square]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("square 5")
	result = strings.TrimSpace(out.String())
	if result != "25" {
		t.Errorf("expected '25', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("add: fn [a b] [(+ a b)]")
	result = strings.TrimSpace(out.String())
	if result != "function[add]" {
		t.Errorf("expected 'function[add]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("add 3 7")
	result = strings.TrimSpace(out.String())
	if result != "10" {
		t.Errorf("expected '10', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("forty-two: fn [] [42]")
	result = strings.TrimSpace(out.String())
	if result != "function[forty-two]" {
		t.Errorf("expected 'function[forty-two]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("forty-two")
	result = strings.TrimSpace(out.String())
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
	}
}

func TestUS4_LocalScopingAndRefinements(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test local scoping and refinements
	out.Reset()
	loop.EvalLineForTest("counter: 10")
	result := strings.TrimSpace(out.String())
	if result != "10" {
		t.Errorf("expected '10', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("increment: fn [] [counter: 1 counter]")
	result = strings.TrimSpace(out.String())
	if result != "function[increment]" {
		t.Errorf("expected 'function[increment]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("increment")
	result = strings.TrimSpace(out.String())
	if result != "1" {
		t.Errorf("expected '1', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("counter")
	result = strings.TrimSpace(out.String())
	if result != "10" {
		t.Errorf("expected '10', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("flag-test: fn [msg --verbose] [verbose]")
	result = strings.TrimSpace(out.String())
	if result != "function[flag-test]" {
		t.Errorf("expected 'function[flag-test]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("flag-test \"hello\"")
	result = strings.TrimSpace(out.String())
	if result != "false" {
		t.Errorf("expected 'false', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("flag-test \"world\" --verbose")
	result = strings.TrimSpace(out.String())
	if result != "true" {
		t.Errorf("expected 'true', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("greet: fn [name --title []] [title]")
	result = strings.TrimSpace(out.String())
	if result != "function[greet]" {
		t.Errorf("expected 'function[greet]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("greet \"Alice\"")
	result = strings.TrimSpace(out.String())
	if result != "" {
		t.Errorf("expected '', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("greet \"Bob\" --title \"Dr.\"")
	result = strings.TrimSpace(out.String())
	if result != "Dr." {
		t.Errorf("expected 'Dr.', got %q", result)
	}
}

func TestUS4_ClosuresRecursionAndErrors(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test closures, recursion and errors
	out.Reset()
	loop.EvalLineForTest("make-adder: fn [x] [fn [y] [(+ x y)]]")
	result := strings.TrimSpace(out.String())
	if result != "function[make-adder]" {
		t.Errorf("expected 'function[make-adder]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("add5: make-adder 5")
	result = strings.TrimSpace(out.String())
	if result != "function[add5]" {
		t.Errorf("expected 'function[add5]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("add5 7")
	result = strings.TrimSpace(out.String())
	if result != "12" {
		t.Errorf("expected '12', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("fact: fn [n] [if (= n 0) [1] [(* n (fact (- n 1)))]]")
	result = strings.TrimSpace(out.String())
	if result != "function[fact]" {
		t.Errorf("expected 'function[fact]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("fact 5")
	result = strings.TrimSpace(out.String())
	if result != "120" {
		t.Errorf("expected '120', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("boom: fn [x] [(/ x 0)]")
	result = strings.TrimSpace(out.String())
	if result != "function[boom]" {
		t.Errorf("expected 'function[boom]', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("boom 5")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "** Math Error") {
		t.Fatalf("expected math error header, got %q", errorOutput)
	}
	if !strings.Contains(errorOutput, "div-zero") {
		t.Fatalf("expected div-zero error, got %q", errorOutput)
	}
}
