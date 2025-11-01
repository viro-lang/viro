package integration

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/repl"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestEvaluatorDebugPause(t *testing.T) {
	evaluator := NewTestEvaluator()
	debug.GlobalDebugger.EnableStepping()

	// Parse a simple expression
	values, err := parse.Parse("x: 42")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// This should return ErrIDDebugPause, not execute
	_, err = evaluator.DoBlock(values)
	if err == nil {
		t.Fatal("expected debug pause error, got nil")
	}

	vErr, ok := err.(*verror.Error)
	if !ok {
		t.Fatalf("expected verror.Error, got %T", err)
	}
	if vErr.ID != verror.ErrIDDebugPause {
		t.Fatalf("expected ErrIDDebugPause, got %s", vErr.ID)
	}
}

func TestREPLDebugModeEntry(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out strings.Builder
	repl := repl.NewREPLForTest(evaluator, &out)

	// Enable stepping
	debug.GlobalDebugger.EnableStepping()

	// Initially not in debug mode
	if repl.IsInDebugMode() {
		t.Fatal("expected REPL not to be in debug mode initially")
	}

	// Evaluate expression that should pause
	values, _ := parse.Parse("x: 42")
	err := repl.EvalParsedValuesForTest(values)

	// Should get pause error and enter debug mode
	if err == nil {
		t.Fatal("expected debug pause error")
	}
	vErr, ok := err.(*verror.Error)
	if !ok {
		t.Fatalf("expected verror.Error, got %T", err)
	}
	if vErr.ID != verror.ErrIDDebugPause {
		t.Fatalf("expected ErrIDDebugPause, got %s", vErr.ID)
	}
	if !repl.IsInDebugMode() {
		t.Fatal("expected REPL to enter debug mode")
	}
}

func TestDebugCommandProcessing(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out strings.Builder
	repl := repl.NewREPLForTest(evaluator, &out)

	// Set up some test data
	repl.EvalLineForTest("x: 42")
	repl.EvalLineForTest("y: 24")

	// Manually enter debug mode
	err := repl.EnterDebugMode()
	if err != nil {
		t.Fatalf("failed to enter debug mode: %v", err)
	}

	// Test locals command
	continueDebug, output, err := repl.ProcessDebugCommandForTest("locals")
	if err != nil {
		t.Fatalf("locals command failed: %v", err)
	}
	if !continueDebug {
		t.Fatal("expected to continue in debug mode after locals")
	}
	if !strings.Contains(output, "Local variables") {
		t.Fatalf("expected locals output, got: %s", output)
	}
	if !strings.Contains(output, "x: 42") {
		t.Fatalf("expected x variable in locals, got: %s", output)
	}
	if !strings.Contains(output, "y: 24") {
		t.Fatalf("expected y variable in locals, got: %s", output)
	}

	// Test stack command
	continueDebug, output, err = repl.ProcessDebugCommandForTest("stack")
	if err != nil {
		t.Fatalf("stack command failed: %v", err)
	}
	if !continueDebug {
		t.Fatal("expected to continue in debug mode after stack")
	}
	if !strings.Contains(output, "Call stack") {
		t.Fatalf("expected stack output, got: %s", output)
	}

	// Test quit command
	continueDebug, output, err = repl.ProcessDebugCommandForTest("quit")
	if err != nil {
		t.Fatalf("quit command failed: %v", err)
	}
	if continueDebug {
		t.Fatal("expected to exit debug mode after quit")
	}
	if repl.IsInDebugMode() {
		t.Fatal("expected REPL to exit debug mode after quit")
	}
}

func TestDebugPrintCommand(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out strings.Builder
	repl := repl.NewREPLForTest(evaluator, &out)

	// Set up test data
	repl.EvalLineForTest("test-var: 100")

	// Enter debug mode
	repl.EnterDebugMode()

	// Test print command
	continueDebug, output, err := repl.ProcessDebugCommandForTest("print test-var")
	if err != nil {
		t.Fatalf("print command failed: %v", err)
	}
	if !continueDebug {
		t.Fatal("expected to continue in debug mode after print")
	}
	if !strings.Contains(output, "100") {
		t.Fatalf("expected print output to contain 100, got: %s", output)
	}

	// Test print with expression
	continueDebug, output, err = repl.ProcessDebugCommandForTest("p test-var + 50")
	if err != nil {
		t.Fatalf("print expression failed: %v", err)
	}
	if !continueDebug {
		t.Fatal("expected to continue in debug mode after print expression")
	}
	if !strings.Contains(output, "150") {
		t.Fatalf("expected print expression output to contain 150, got: %s", output)
	}
}

func TestDebugSessionLifecycle(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out strings.Builder
	repl := repl.NewREPLForTest(evaluator, &out)

	// Not in debug mode initially
	if repl.IsInDebugMode() {
		t.Fatal("expected REPL not to be in debug mode initially")
	}

	// Enter debug mode
	err := repl.EnterDebugMode()
	if err != nil {
		t.Fatalf("failed to enter debug mode: %v", err)
	}
	if !repl.IsInDebugMode() {
		t.Fatal("expected REPL to be in debug mode")
	}

	// Should have active debug session
	session := repl.GetDebugSessionForTest()
	if session == nil {
		t.Fatal("expected debug session to exist")
	}
	if !repl.IsDebugSessionActiveForTest() {
		t.Fatal("expected debug session to be active")
	}

	// Exit debug mode
	repl.ExitDebugMode()
	if repl.IsInDebugMode() {
		t.Fatal("expected REPL to exit debug mode")
	}
	if repl.GetDebugSessionForTest() != nil {
		t.Fatal("expected debug session to be nil after exit")
	}
}

func TestDebugStepAndContinue(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out strings.Builder
	repl := repl.NewREPLForTest(evaluator, &out)

	// Enable stepping initially
	debug.GlobalDebugger.EnableStepping()
	if !debug.GlobalDebugger.IsStepping() {
		t.Fatal("expected debugger to be stepping initially")
	}

	// Enter debug mode
	repl.EnterDebugMode()

	// Test step command (should keep stepping enabled)
	continueDebug, _, err := repl.ProcessDebugCommandForTest("step")
	if err != nil {
		t.Fatalf("step command failed: %v", err)
	}
	if continueDebug {
		t.Fatal("expected to exit debug mode after step")
	}
	if !debug.GlobalDebugger.IsStepping() {
		t.Fatal("expected debugger to still be stepping after step command")
	}

	// Re-enter debug mode
	repl.EnterDebugMode()

	// Test continue command (should disable stepping)
	continueDebug, _, err = repl.ProcessDebugCommandForTest("continue")
	if err != nil {
		t.Fatalf("continue command failed: %v", err)
	}
	if continueDebug {
		t.Fatal("expected to exit debug mode after continue")
	}
	if debug.GlobalDebugger.IsStepping() {
		t.Fatal("expected debugger to disable stepping after continue command")
	}
}

func TestInteractiveDebugFlow(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out strings.Builder
	repl := repl.NewREPLForTest(evaluator, &out)

	// Enable stepping
	debug.GlobalDebugger.EnableStepping()

	// Simulate the debug flow programmatically
	values, _ := parse.Parse("x: 42")

	// First evaluation should pause and enter debug mode
	err := repl.EvalParsedValuesForTest(values)
	if err == nil {
		t.Fatal("expected debug pause error")
	}
	vErr, ok := err.(*verror.Error)
	if !ok {
		t.Fatalf("expected verror.Error, got %T", err)
	}
	if vErr.ID != verror.ErrIDDebugPause {
		t.Fatalf("expected ErrIDDebugPause, got %s", vErr.ID)
	}
	if !repl.IsInDebugMode() {
		t.Fatal("expected REPL to enter debug mode")
	}

	// Simulate user typing "locals"
	continueDebug, output, _ := repl.ProcessDebugCommandForTest("locals")
	if !continueDebug {
		t.Fatal("expected to continue in debug mode after locals")
	}
	if !strings.Contains(output, "x: 42") {
		t.Fatalf("expected locals to show x: 42, got: %s", output)
	}

	// Simulate user typing "continue"
	continueDebug, _, _ = repl.ProcessDebugCommandForTest("continue")
	if continueDebug {
		t.Fatal("expected to exit debug mode after continue")
	}
	if repl.IsInDebugMode() {
		t.Fatal("expected REPL to exit debug mode after continue")
	}
}
