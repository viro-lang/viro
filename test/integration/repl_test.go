package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestREPL_ErrorRecovery(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Function definition - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("square: fn [n] [n * n]")
	output := strings.TrimSpace(out.String())
	if !strings.Contains(output, "function") {
		t.Fatalf("expected function definition output, got %q", output)
	}

	// Error case - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("square \"oops\"")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "** Script Error") {
		t.Fatalf("expected script error header, got %q", errorOutput)
	}
	if !strings.Contains(errorOutput, "square") {
		t.Fatalf("expected call stack or message to mention square, got %q", errorOutput)
	}

	// Successful evaluation after error - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("square 4")
	output = strings.TrimSpace(out.String())
	if !strings.Contains(output, "16") {
		t.Fatalf("expected successful evaluation after error, got %q", output)
	}
}

func TestREPL_StatePreservedAfterError(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Assignment - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("x: 10")
	output := strings.TrimSpace(out.String())
	if !strings.Contains(output, "10") {
		t.Fatalf("expected assignment result to include 10, got %q", output)
	}

	// Error - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("1 / 0")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "** Math Error") {
		t.Fatalf("expected math error header, got %q", errorOutput)
	}

	// Variable lookup after error - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("x")
	output = strings.TrimSpace(out.String())
	if !strings.Contains(output, "10") {
		t.Fatalf("expected x to retain value 10 after error, got %q", output)
	}

	// Reassignment - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("x: x + 5")
	output = strings.TrimSpace(out.String())
	if !strings.Contains(output, "15") {
		t.Fatalf("expected reassignment result to include 15, got %q", output)
	}

	// Final lookup - output goes to buffer
	out.Reset()
	loop.EvalLineForTest("x")
	output = strings.TrimSpace(out.String())
	if !strings.Contains(output, "15") {
		t.Fatalf("expected x to reflect updated value 15, got %q", output)
	}
}

func TestREPL_CommandHistory(t *testing.T) {
	evaluator := NewTestEvaluator()
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

func TestREPL_MultiLineInput(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if loop.AwaitingContinuation() {
		t.Fatalf("expected REPL not to await continuation initially")
	}

	loop.EvalLineForTest("value: (")
	if out.String() != "" {
		t.Fatalf("expected no output for incomplete first line, got %q", out.String())
	}
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected REPL to await continuation after opening paren")
	}
	out.Reset()

	loop.EvalLineForTest("  10 + 5")
	if out.String() != "" {
		t.Fatalf("expected no output while awaiting continuation, got %q", out.String())
	}
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected REPL to remain in continuation state after intermediate line")
	}
	out.Reset()

	loop.EvalLineForTest(")")
	if loop.AwaitingContinuation() {
		t.Fatalf("expected REPL to exit continuation state after closing paren")
	}
	// Multi-line evaluation result goes to buffer
	out.Reset()
	loop.EvalLineForTest("value")
	output := strings.TrimSpace(out.String())
	if !strings.Contains(output, "15") {
		t.Fatalf("expected value to be preserved after multi-line evaluation, got %q", output)
	}
}

func TestREPL_ExitCommands(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	if !loop.ShouldContinue() {
		t.Fatalf("new REPL should continue running")
	}

	loop.EvalLineForTest("quit")
	if loop.ShouldContinue() {
		t.Fatalf("REPL should request shutdown after quit")
	}
	if output := out.String(); !strings.Contains(output, "Goodbye!") {
		t.Fatalf("expected goodbye message after quit, got %q", output)
	}
	out.Reset()

	loop.ResetForTest()
	if !loop.ShouldContinue() {
		t.Fatalf("reset REPL should be ready to continue")
	}

	loop.EvalLineForTest("exit")
	if loop.ShouldContinue() {
		t.Fatalf("REPL should request shutdown after exit")
	}
	if output := out.String(); !strings.Contains(output, "Goodbye!") {
		t.Fatalf("expected goodbye message after exit, got %q", output)
	}
}

func TestREPL_CtrlCInterrupt(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	loop.EvalLineForTest("value: (")
	if !loop.AwaitingContinuation() {
		t.Fatalf("expected continuation state after opening paren")
	}
	if !loop.ShouldContinue() {
		t.Fatalf("REPL should still be marked as running during multi-line input")
	}

	loop.SimulateInterruptForTest()
	if loop.AwaitingContinuation() {
		t.Fatalf("expected interrupt to clear continuation state")
	}
	if !loop.ShouldContinue() {
		t.Fatalf("interrupt should not stop REPL")
	}
	if output := out.String(); !strings.Contains(output, "^C") {
		t.Fatalf("expected interrupt indicator, got %q", output)
	}
	out.Reset()

	// Evaluation after interrupt goes to buffer
	out.Reset()
	loop.EvalLineForTest("value: 10")
	output := strings.TrimSpace(out.String())
	if !strings.Contains(output, "10") {
		t.Fatalf("expected evaluation to continue after interrupt, got %q", output)
	}
}

func TestREPL_HistoryPersistence(t *testing.T) {
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "history.txt")
	t.Setenv("VIRO_HISTORY_FILE", historyFile)

	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	loop.EvalLineForTest("persist1")
	out.Reset()
	loop.EvalLineForTest("persist2")
	out.Reset()

	data, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("expected history file to be written, got error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"persist1", "persist2"}
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines in history file, got %d: %q", len(want), len(lines), lines)
	}
	for i, entry := range want {
		if lines[i] != entry {
			t.Fatalf("history file entry %d mismatch: expected %q got %q", i, entry, lines[i])
		}
	}

	var out2 bytes.Buffer
	loop2 := repl.NewREPLForTest(NewTestEvaluator(), &out2)
	history := loop2.HistoryEntries()
	if len(history) != len(want) {
		t.Fatalf("expected %d entries loaded from history, got %d", len(want), len(history))
	}
	for i, entry := range want {
		if history[i] != entry {
			t.Fatalf("history entry %d mismatch: expected %q got %q", i, entry, history[i])
		}
	}
}

func TestREPL_PrinOutputVisible(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test prin output is visible without newline
	out.Reset()
	loop.EvalLineForTest(`prin "hello"`)
	output := out.String()
	if output != "hello" {
		t.Fatalf("expected prin output to be exactly 'hello', got %q", output)
	}

	// Test print output has newline
	out.Reset()
	loop.EvalLineForTest(`print "world"`)
	output = out.String()
	if output != "world\n" {
		t.Fatalf("expected print output to be 'world\\n', got %q", output)
	}
}

func TestREPL_InteractivePrinEndsWithNewline(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test interactive prin output gets extra newline safeguard
	out.Reset()
	loop.EvalLineInteractiveForTest(`prin "hello"`)
	output := out.String()
	if output != "hello\n" {
		t.Fatalf("expected interactive prin output to end with newline safeguard, got %q", output)
	}

	// Test interactive print output still has its own newline
	out.Reset()
	loop.EvalLineInteractiveForTest(`print "world"`)
	output = out.String()
	if output != "world\n" {
		t.Fatalf("expected interactive print output to be 'world\\n', got %q", output)
	}

	// Test non-interactive prin behavior unchanged
	out.Reset()
	loop.EvalLineForTest(`prin "test"`)
	output = out.String()
	if output != "test" {
		t.Fatalf("expected non-interactive prin output to remain unchanged, got %q", output)
	}
}
