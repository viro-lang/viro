package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

func evalLine(t *testing.T, loop *repl.REPL, out *bytes.Buffer, input string) string {
	t.Helper()
	out.Reset()
	loop.EvalLineForTest(input)
	return out.String()
}

func TestUS1_BasicExpressions(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if output := strings.TrimSpace(evalLine(t, loop, &out, "42")); output != "42" {
		t.Fatalf("literal evaluation expected 42, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "\"hello\"")); output != "\"hello\"" {
		t.Fatalf("string literal expected \"hello\", got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "3 + 4")); output != "7" {
		t.Fatalf("arithmetic expected 7, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "x: 10")); output != "10" {
		t.Fatalf("assignment expected 10, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "x")); output != "10" {
		t.Fatalf("word lookup expected 10, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "[1 + 2]")); output != "[(+ 1 2)]" {
		t.Fatalf("block evaluation expected [(+ 1 2)], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "(1 + 2)")); output != "3" {
		t.Fatalf("paren evaluation expected 3, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "3 + 4 * 2")); output != "11" {
		t.Fatalf("precedence expected 11, got %q", output)
	}

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
