package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestREPLOptions_CustomPrompt(t *testing.T) {
	opts := &repl.Options{
		Prompt: "λ> ",
		Args:   []string{},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with options: %v", err)
	}

	// The custom prompt is used internally - we can't test it directly in the test REPL
	// but we can verify the REPL was created successfully with the option
	if loop == nil {
		t.Fatal("expected REPL to be created")
	}
}

func TestREPLOptions_NoWelcome(t *testing.T) {
	opts := &repl.Options{
		NoWelcome: true,
		Args:      []string{},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with options: %v", err)
	}

	// Test that REPL was created successfully
	if loop == nil {
		t.Fatal("expected REPL to be created")
	}

	// Note: The welcome message is shown in Run(), not in NewREPLWithOptions()
	// so we can't directly test it here, but we verified the option is accepted
}

func TestREPLOptions_NoHistory(t *testing.T) {
	opts := &repl.Options{
		NoHistory: true,
		Args:      []string{},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with options: %v", err)
	}

	if loop == nil {
		t.Fatal("expected REPL to be created")
	}

	// Verify no history is recorded when NoHistory is true
	loop.EvalLineForTest("x: 42")
	loop.EvalLineForTest("y: 100")

	history := loop.HistoryEntries()
	if len(history) != 0 {
		t.Errorf("expected empty history with NoHistory=true, got %d entries", len(history))
	}
}

func TestREPLOptions_WithHistory(t *testing.T) {
	// Use a custom history file that doesn't exist so we start with clean slate
	tmpFile := "/tmp/viro_test_with_history.txt"

	// Clean up any existing file from previous test runs
	_ = os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	opts := &repl.Options{
		NoHistory:   false,
		HistoryFile: tmpFile,
		Args:        []string{},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with options: %v", err)
	}

	if loop == nil {
		t.Fatal("expected REPL to be created")
	}

	// Verify history is recorded when NoHistory is false
	loop.EvalLineForTest("x: 42")
	loop.EvalLineForTest("y: 100")

	history := loop.HistoryEntries()
	if len(history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(history))
	}

	if len(history) >= 2 {
		if !strings.Contains(history[0], "x: 42") {
			t.Errorf("expected first history entry to contain 'x: 42', got %q", history[0])
		}
		if !strings.Contains(history[1], "y: 100") {
			t.Errorf("expected second history entry to contain 'y: 100', got %q", history[1])
		}
	}
}

func TestREPLOptions_CustomHistoryFile(t *testing.T) {
	tmpFile := "/tmp/viro_test_history.txt"

	opts := &repl.Options{
		HistoryFile: tmpFile,
		Args:        []string{},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with options: %v", err)
	}

	if loop == nil {
		t.Fatal("expected REPL to be created")
	}

	// Note: Testing actual file persistence would require mocking or filesystem operations
	// Here we just verify the option is accepted without error
}

func TestREPLOptions_Args(t *testing.T) {
	opts := &repl.Options{
		Args: []string{"arg1", "arg2", "arg3"},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with options: %v", err)
	}

	if loop == nil {
		t.Fatal("expected REPL to be created")
	}

	// Test that args are accessible via system.args
	loop.EvalLineForTest("length? system.args")
	// The fact that we don't get an error means system.args exists
}

func TestREPLOptions_TraceOn(t *testing.T) {
	opts := &repl.Options{
		TraceOn: true,
		Args:    []string{},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with options: %v", err)
	}

	if loop == nil {
		t.Fatal("expected REPL to be created")
	}

	// Note: Testing trace output would require capturing trace events
	// Here we just verify the option is accepted without error
}

func TestREPLOptions_AllOptions(t *testing.T) {
	opts := &repl.Options{
		Prompt:      "test> ",
		NoWelcome:   true,
		NoHistory:   true,
		HistoryFile: "/tmp/test_history.txt",
		TraceOn:     false,
		Args:        []string{"arg1", "arg2"},
	}

	loop, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		t.Fatalf("failed to create REPL with all options: %v", err)
	}

	if loop == nil {
		t.Fatal("expected REPL to be created")
	}

	// Verify all options were accepted and REPL is functional
	loop.EvalLineForTest("x: 123")

	// With NoHistory=true, there should be no history
	history := loop.HistoryEntries()
	if len(history) != 0 {
		t.Errorf("expected empty history with NoHistory=true, got %d entries", len(history))
	}
}
