package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS4_FunctionDefinitionsAndCalls(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if output := strings.TrimSpace(evalLine(t, loop, &out, "square: fn [n] [(* n n)]")); output != "function[square]" {
		t.Fatalf("function definition should yield function[square], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "square 5")); output != "25" {
		t.Fatalf("square 5 expected 25, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "add: fn [a b] [(+ a b)]")); output != "function[add]" {
		t.Fatalf("add definition should yield function[add], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "add 3 7")); output != "10" {
		t.Fatalf("add 3 7 expected 10, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "forty-two: fn [] [42]")); output != "function[forty-two]" {
		t.Fatalf("zero-arg function definition should yield function[forty-two], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "forty-two")); output != "42" {
		t.Fatalf("calling zero-arg function expected 42, got %q", output)
	}
}

func TestUS4_LocalScopingAndRefinements(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if output := strings.TrimSpace(evalLine(t, loop, &out, "counter: 10")); output != "10" {
		t.Fatalf("global counter initialization expected 10, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "increment: fn [] [counter: 1 counter]")); output != "function[increment]" {
		t.Fatalf("increment definition should yield function[increment], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "increment")); output != "1" {
		t.Fatalf("increment call expected local value 1, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "counter")); output != "10" {
		t.Fatalf("global counter should remain 10, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "flag-test: fn [msg --verbose] [verbose]")); output != "function[flag-test]" {
		t.Fatalf("flag-test definition should yield function[flag-test], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "flag-test \"hello\"")); output != "false" {
		t.Fatalf("flag refinement default expected false, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "flag-test \"world\" --verbose")); output != "true" {
		t.Fatalf("flag refinement set expected true, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "greet: fn [name --title []] [title]")); output != "function[greet]" {
		t.Fatalf("greet definition should yield function[greet], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "greet \"Alice\"")); output != "" {
		t.Fatalf("value refinement default expected none (suppressed output), got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "greet \"Bob\" --title \"Dr.\"")); output != "\"Dr.\"" {
		t.Fatalf("value refinement with title expected \"Dr.\", got %q", output)
	}
}

func TestUS4_ClosuresRecursionAndErrors(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if output := strings.TrimSpace(evalLine(t, loop, &out, "make-adder: fn [x] [fn [y] [(+ x y)]]")); output != "function[make-adder]" {
		t.Fatalf("make-adder definition should yield function[make-adder], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "add5: make-adder 5")); output != "function[add5]" {
		t.Fatalf("closure creation expected function[add5], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "add5 7")); output != "12" {
		t.Fatalf("closure invocation expected 12, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "fact: fn [n] [if (= n 0) [1] [(* n (fact (- n 1)))]]")); output != "function[fact]" {
		t.Fatalf("recursive function definition should yield function[fact], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "fact 5")); output != "120" {
		t.Fatalf("fact 5 expected 120, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "boom: fn [x] [(/ x 0)]")); output != "function[boom]" {
		t.Fatalf("boom definition should yield function[boom], got %q", output)
	}

	out.Reset()
	loop.EvalLineForTest("boom 5")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "** Math Error") {
		t.Fatalf("expected math error header, got %q", errorOutput)
	}
	if !strings.Contains(errorOutput, "Division by zero") {
		t.Fatalf("expected division by zero message, got %q", errorOutput)
	}
	if !strings.Contains(strings.ToLower(errorOutput), "boom") {
		t.Fatalf("expected error output to mention function context boom, got %q", errorOutput)
	}
}
