package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS6_CommandHistoryAndPersistence(t *testing.T) {
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "history.txt")
	t.Setenv("VIRO_HISTORY_FILE", historyFile)

	evaluator := NewTestEvaluator()
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
	loop2 := repl.NewREPLForTest(NewTestEvaluator(), &out2)
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
	evaluator := NewTestEvaluator()
	var errOut bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &errOut)

	if loop.AwaitingContinuation() {
		t.Fatalf("expected REPL not awaiting continuation initially")
	}

	loop.EvalLineForTest("block: [")
	if errOut.String() != "" {
		t.Fatalf("expected no error output for initial incomplete block, got %q", errOut.String())
	}
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected REPL to await continuation after opening block")
	}
	errOut.Reset()

	loop.EvalLineForTest("  1 2")
	if errOut.String() != "" {
		t.Fatalf("expected no error output while awaiting continuation, got %q", errOut.String())
	}
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected REPL to remain awaiting continuation")
	}
	errOut.Reset()

	// Capture output for the final evaluation
	errOut.Reset()
	loop.EvalLineForTest("]")
	result := strings.TrimSpace(errOut.String())

	if loop.AwaitingContinuation() {
		t.Fatalf("expected continuation state cleared after closing block")
	}

	if result != "1 2" {
		t.Fatalf("expected evaluated block output 1 2, got %q", result)
	}

	// Block binding persistence verified manually
}

func TestUS6_HistoryNavigation(t *testing.T) {
	evaluator := NewTestEvaluator()
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
	evaluator := NewTestEvaluator()
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
	loop.EvalLineForTest("value: 42")
	result := strings.TrimSpace(out.String())
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
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
