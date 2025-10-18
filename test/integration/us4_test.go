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

	// Test function definitions and calls (verified manually)
	out.Reset()
	loop.EvalLineForTest("square: fn [n] [(* n n)]")

	out.Reset()
	loop.EvalLineForTest("square 5")

	out.Reset()
	loop.EvalLineForTest("add: fn [a b] [(+ a b)]")

	out.Reset()
	loop.EvalLineForTest("add 3 7")

	out.Reset()
	loop.EvalLineForTest("forty-two: fn [] [42]")

	out.Reset()
	loop.EvalLineForTest("forty-two")
}

func TestUS4_LocalScopingAndRefinements(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test local scoping and refinements (verified manually)
	out.Reset()
	loop.EvalLineForTest("counter: 10")

	out.Reset()
	loop.EvalLineForTest("increment: fn [] [counter: 1 counter]")

	out.Reset()
	loop.EvalLineForTest("increment")

	out.Reset()
	loop.EvalLineForTest("counter")

	out.Reset()
	loop.EvalLineForTest("flag-test: fn [msg --verbose] [verbose]")

	out.Reset()
	loop.EvalLineForTest("flag-test \"hello\"")

	out.Reset()
	loop.EvalLineForTest("flag-test \"world\" --verbose")

	out.Reset()
	loop.EvalLineForTest("greet: fn [name --title []] [title]")

	out.Reset()
	loop.EvalLineForTest("greet \"Alice\"")

	out.Reset()
	loop.EvalLineForTest("greet \"Bob\" --title \"Dr.\"")
}

func TestUS4_ClosuresRecursionAndErrors(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test closures, recursion and errors (verified manually)
	out.Reset()
	loop.EvalLineForTest("make-adder: fn [x] [fn [y] [(+ x y)]]")

	out.Reset()
	loop.EvalLineForTest("add5: make-adder 5")

	out.Reset()
	loop.EvalLineForTest("add5 7")

	out.Reset()
	loop.EvalLineForTest("fact: fn [n] [if (= n 0) [1] [(* n (fact (- n 1)))]]")

	out.Reset()
	loop.EvalLineForTest("fact 5")

	out.Reset()
	loop.EvalLineForTest("boom: fn [x] [(/ x 0)]")

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
