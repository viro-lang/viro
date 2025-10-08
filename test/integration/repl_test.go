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

func TestREPL_StatePreservedAfterError(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	loop.EvalLineForTest("x: 10")
	if output := out.String(); !strings.Contains(output, "10") {
		t.Fatalf("expected assignment result to include 10, got %q", output)
	}
	out.Reset()

	loop.EvalLineForTest("1 / 0")
	if output := out.String(); !strings.Contains(output, "** Math Error") {
		t.Fatalf("expected math error header, got %q", output)
	}
	out.Reset()

	loop.EvalLineForTest("x")
	if output := out.String(); !strings.Contains(output, "10") {
		t.Fatalf("expected x to retain value 10 after error, got %q", output)
	}
	out.Reset()

	loop.EvalLineForTest("x: x + 5")
	if output := out.String(); !strings.Contains(output, "15") {
		t.Fatalf("expected reassignment result to include 15, got %q", output)
	}
	out.Reset()

	loop.EvalLineForTest("x")
	if output := out.String(); !strings.Contains(output, "15") {
		t.Fatalf("expected x to reflect updated value 15, got %q", output)
	}
}

func TestREPL_CommandHistory(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	commands := []string{
		"alpha: 1",
		"beta: alpha + 2",
		"beta",
	}

	for _, cmd := range commands {
		loop.EvalLineForTest(cmd)
		out.Reset()
	}

	history := loop.HistoryEntries()
	if len(history) != len(commands) {
		t.Fatalf("expected %d history entries, got %d", len(commands), len(history))
	}

	for i, cmd := range commands {
		if history[i] != cmd {
			t.Fatalf("history entry %d mismatch: expected %q got %q", i, cmd, history[i])
		}
	}

	entry, ok := loop.HistoryUp()
	if !ok || entry != commands[2] {
		t.Fatalf("first history up should return %q, got %q (ok=%v)", commands[2], entry, ok)
	}

	entry, ok = loop.HistoryUp()
	if !ok || entry != commands[1] {
		t.Fatalf("second history up should return %q, got %q (ok=%v)", commands[1], entry, ok)
	}

	entry, ok = loop.HistoryUp()
	if !ok || entry != commands[0] {
		t.Fatalf("third history up should return %q, got %q (ok=%v)", commands[0], entry, ok)
	}

	entry, ok = loop.HistoryUp()
	if !ok || entry != commands[0] {
		t.Fatalf("additional history up should stay on first command %q, got %q (ok=%v)", commands[0], entry, ok)
	}

	entry, ok = loop.HistoryDown()
	if !ok || entry != commands[1] {
		t.Fatalf("history down should return %q, got %q (ok=%v)", commands[1], entry, ok)
	}

	entry, ok = loop.HistoryDown()
	if !ok || entry != commands[2] {
		t.Fatalf("history down should return %q, got %q (ok=%v)", commands[2], entry, ok)
	}

	entry, ok = loop.HistoryDown()
	if ok || entry != "" {
		t.Fatalf("history down at end should signal empty input, got %q (ok=%v)", entry, ok)
	}

	loop.EvalLineForTest("gamma: beta + 5")
	out.Reset()
	entry, ok = loop.HistoryUp()
	if !ok || entry != "gamma: beta + 5" {
		t.Fatalf("after new command, history up should return latest entry, got %q (ok=%v)", entry, ok)
	}
}
