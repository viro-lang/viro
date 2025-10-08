package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS6_CommandHistoryAndPersistence(t *testing.T) {
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "history.txt")
	t.Setenv("VIRO_HISTORY_FILE", historyFile)

	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	commands := []string{
		"first-cmd",
		"second-cmd",
		"third-cmd",
	}

	for _, cmd := range commands {
		loop.EvalLineForTest(cmd)
		out.Reset()
	}

	hist := loop.HistoryEntries()
	if len(hist) != len(commands) {
		t.Fatalf("expected %d history entries, got %d", len(commands), len(hist))
	}

	for i, cmd := range commands {
		if hist[i] != cmd {
			t.Fatalf("history entry mismatch at %d: expected %q got %q", i, cmd, hist[i])
		}
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("expected history file to be written, got error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(commands) {
		t.Fatalf("expected %d history lines, got %d", len(commands), len(lines))
	}
	for i, cmd := range commands {
		if lines[i] != cmd {
			t.Fatalf("history file entry %d mismatch: expected %q got %q", i, cmd, lines[i])
		}
	}

	var out2 bytes.Buffer
	loop2 := repl.NewREPLForTest(eval.NewEvaluator(), &out2)
	hist2 := loop2.HistoryEntries()
	if len(hist2) != len(commands) {
		t.Fatalf("expected %d entries loaded into new REPL, got %d", len(commands), len(hist2))
	}
	for i, cmd := range commands {
		if hist2[i] != cmd {
			t.Fatalf("persisted history mismatch at %d: expected %q got %q", i, cmd, hist2[i])
		}
	}
}

func TestUS6_MultiLineContinuation(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if loop.AwaitingContinuation() {
		t.Fatalf("expected REPL not awaiting continuation initially")
	}

	loop.EvalLineForTest("block: [")
	if out.String() != "" {
		t.Fatalf("expected no output for initial incomplete block, got %q", out.String())
	}
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected REPL to await continuation after opening block")
	}
	out.Reset()

	loop.EvalLineForTest("  1 2")
	if out.String() != "" {
		t.Fatalf("expected no output while awaiting continuation, got %q", out.String())
	}
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected REPL to remain awaiting continuation")
	}
	out.Reset()

	loop.EvalLineForTest("]")
	if loop.AwaitingContinuation() {
		t.Fatalf("expected continuation state cleared after closing block")
	}
	if result := strings.TrimSpace(out.String()); result != "[1 2]" {
		t.Fatalf("expected evaluated block output [1 2], got %q", result)
	}
	out.Reset()

	if output := strings.TrimSpace(evalLine(t, loop, &out, "block")); output != "[1 2]" {
		t.Fatalf("expected block binding persisted, got %q", output)
	}
}

func TestUS6_HistoryNavigation(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	commands := []string{"alpha: 1", "beta: alpha + 2", "beta"}
	for _, cmd := range commands {
		loop.EvalLineForTest(cmd)
		out.Reset()
	}

	entry, ok := loop.HistoryUp()
	if !ok || entry != commands[2] {
		t.Fatalf("first history up expected %q, got %q (ok=%v)", commands[2], entry, ok)
	}
	entry, ok = loop.HistoryUp()
	if !ok || entry != commands[1] {
		t.Fatalf("second history up expected %q, got %q (ok=%v)", commands[1], entry, ok)
	}
	entry, ok = loop.HistoryUp()
	if !ok || entry != commands[0] {
		t.Fatalf("third history up expected %q, got %q (ok=%v)", commands[0], entry, ok)
	}

	entry, ok = loop.HistoryDown()
	if !ok || entry != commands[1] {
		t.Fatalf("history down expected %q, got %q (ok=%v)", commands[1], entry, ok)
	}
	entry, ok = loop.HistoryDown()
	if !ok || entry != commands[2] {
		t.Fatalf("history down expected %q, got %q (ok=%v)", commands[2], entry, ok)
	}
	entry, ok = loop.HistoryDown()
	if ok || entry != "" {
		t.Fatalf("history down at end should be empty, got %q (ok=%v)", entry, ok)
	}
}

func TestUS6_ExitAndInterrupt(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if !loop.ShouldContinue() {
		t.Fatalf("expected REPL should continue at start")
	}

	loop.EvalLineForTest("quit")
	if loop.ShouldContinue() {
		t.Fatalf("quit should signal REPL stop")
	}
	if !strings.Contains(out.String(), "Goodbye!") {
		t.Fatalf("expected goodbye message after quit, got %q", out.String())
	}

	loop.ResetForTest()
	out.Reset()

	loop.EvalLineForTest("exit")
	if loop.ShouldContinue() {
		t.Fatalf("exit should signal REPL stop")
	}
	if !strings.Contains(out.String(), "Goodbye!") {
		t.Fatalf("expected goodbye message after exit, got %q", out.String())
	}

	loop.ResetForTest()
	out.Reset()

	loop.EvalLineForTest("value: (")
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected awaiting continuation after opening paren")
	}

	loop.SimulateInterruptForTest()
	if loop.AwaitingContinuation() {
		t.Fatalf("interrupt should clear continuation state")
	}
	if !strings.Contains(out.String(), "^C") {
		t.Fatalf("expected caret indicator after interrupt, got %q", out.String())
	}

	out.Reset()
	if output := strings.TrimSpace(evalLine(t, loop, &out, "value: 42")); output != "42" {
		t.Fatalf("expected evaluation to proceed after interrupt, got %q", output)
	}
}

func TestUS6_WelcomeMessage(t *testing.T) {
	welcome := repl.WelcomeMessage()
	if !strings.Contains(welcome, "Viro 0.1.0") {
		t.Fatalf("expected welcome message to include version, got %q", welcome)
	}
	if !strings.Contains(welcome, "Type 'exit' or 'quit' to leave") {
		t.Fatalf("expected welcome instructions, got %q", welcome)
	}
}
