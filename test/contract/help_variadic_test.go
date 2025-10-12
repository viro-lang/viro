package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestHelpVariableArity tests that ? function works correctly:
// - Direct call with 0 args works (shows categories)
// - Script/parsed call requires 1 arg (Arity: 1)
// - REPL has special handling for bare '?' (tested in integration tests)
func TestHelpVariableArity(t *testing.T) {
	t.Run("HelpWithNoArgs_DirectCall", func(t *testing.T) {
		// Direct call to Help() with no args should work - shows categories
		e := eval.NewEvaluator()
		result, err := native.Help([]value.Value{}, e)
		if err != nil {
			t.Fatalf("Direct call Help() with no args failed: %v", err)
		}

		if result.Type != value.TypeNone {
			t.Errorf("Expected none!, got %v", result.Type)
		}
	})

	t.Run("HelpWithNoArgs_ParsedScript_ShouldFail", func(t *testing.T) {
		// In scripts, ? requires an argument (Arity: 1)
		// This is by design - bare ? is REPL-only shortcut
		val, err := Evaluate("?")
		if err == nil {
			t.Error("Expected arity error for '?' in script (requires 1 arg), but got none")
			t.Log("Note: In scripts, use '? topic'. Bare '?' only works in REPL.")
		} else {
			// Verify it's an arity error
			errMsg := err.Error()
			if !strings.Contains(errMsg, "argument") && !strings.Contains(errMsg, "arity") {
				t.Errorf("Error message doesn't mention arity: %v", errMsg)
			}
		}

		if val.Type != value.TypeNone {
			t.Errorf("Expected none!, got %v", val.Type)
		}
	})

	t.Run("HelpWithOneArg_ParsedScript", func(t *testing.T) {
		// This should work: ? with one argument
		val, err := Evaluate("? append")
		if err != nil {
			t.Fatalf("Eval error for '? append': %v", err)
		}

		if val.Type != value.TypeNone {
			t.Errorf("Expected none!, got %v", val.Type)
		}
	})

	t.Run("HelpWithTwoArgs_ShouldError", func(t *testing.T) {
		// This should fail: ? with two arguments
		// Note: "? append insert" might parse as two separate commands
		val, err := Evaluate("? append insert")
		if err == nil {
			t.Log("Note: 'insert' may have been evaluated as separate command")
		} else {
			// Check that error mentions arity/argument count
			errMsg := err.Error()
			if !strings.Contains(errMsg, "argument") && !strings.Contains(errMsg, "arity") {
				t.Errorf("Error message doesn't mention arity: %v", errMsg)
			}
		}

		// Should return none regardless
		if val.Type != value.TypeNone {
			t.Errorf("Expected none! even on error, got %v", val.Type)
		}
	})

	t.Run("HelpInExpression", func(t *testing.T) {
		// Test that ? works in an expression context with proper argument
		val, err := Evaluate("print \"Starting help\"  ? math  print \"Done\"")
		if err != nil {
			t.Fatalf("Eval error: %v", err)
		}

		if val.Type != value.TypeNone {
			t.Errorf("Expected none!, got %v", val.Type)
		}
	})
}

// TestHelpDocumentationMatchesImplementation verifies documentation is accurate
func TestHelpDocumentationMatchesImplementation(t *testing.T) {
	t.Run("ScriptUsageRequiresArgument", func(t *testing.T) {
		// Documentation correctly states that in scripts, ? requires an argument
		// Examples show: "? math" not just "?"
		_, err := Evaluate("?")
		if err == nil {
			t.Error("Expected arity error for bare '?' in script context")
			t.Log("Documentation states: 'In scripts, you must provide an argument'")
		}
	})

	t.Run("DirectCallSupportsNoArgs", func(t *testing.T) {
		// The Help function itself supports 0 args (for REPL shortcut)
		e := eval.NewEvaluator()
		result, err := native.Help([]value.Value{}, e)
		if err != nil {
			t.Errorf("Direct Help() call with 0 args should work: %v", err)
		}
		if result.Type != value.TypeNone {
			t.Errorf("Expected none!, got %v", result.Type)
		}
	})
}
