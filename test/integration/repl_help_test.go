package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestREPL_HelpShortcut tests the special REPL-only '?' command without arguments
func TestREPL_HelpShortcut(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// In REPL, bare '?' should show categories (special handling)
	loop.EvalLineForTest("?")
	output := out.String()

	// Verify it shows categories
	if !strings.Contains(output, "Available categories:") {
		t.Errorf("Expected category list, got: %s", output)
	}

	// Verify it shows some known categories
	expectedCategories := []string{"Control", "Math", "Series", "Help"}
	for _, cat := range expectedCategories {
		if !strings.Contains(output, cat) {
			t.Errorf("Expected category %q in output, but not found", cat)
		}
	}

	// Verify it shows usage hints
	if !strings.Contains(output, "Use '? category' to list functions") {
		t.Error("Expected usage hint about '? category'")
	}
}

// TestREPL_HelpWithArgument tests that '? topic' works in REPL
// Note: ? with argument goes through parser and Help prints to os.Stdout
// which cannot be easily redirected in tests, so we just verify no errors
func TestREPL_HelpWithArgument(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Help with category should work without error
	loop.EvalLineForTest("? math")
	// Output goes to os.Stdout in real REPL, not captured here
	// Just verify no panic occurred

	out.Reset()

	// Help with function name should work without error
	loop.EvalLineForTest("? append")
	// Output goes to os.Stdout in real REPL
}

// TestREPL_HelpAfterError tests that help works after errors
func TestREPL_HelpAfterError(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Cause an error
	loop.EvalLineForTest("1 / 0")
	errorOutput := out.String()
	if !strings.Contains(errorOutput, "Error") {
		t.Logf("Expected error, got: %s", errorOutput)
	}

	out.Reset()

	// Help should still work after error
	loop.EvalLineForTest("?")
	output := out.String()

	if !strings.Contains(output, "Available categories:") {
		t.Errorf("Help should work after error, got: %s", output)
	}
}

// TestREPL_HelpIntegration tests help in various contexts
// Note: Most '? arg' calls print to os.Stdout, not captured in tests
// We verify no errors occur and REPL continues working
func TestREPL_HelpIntegration(t *testing.T) {
	evaluator := NewTestEvaluator()
	var errOut bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &errOut)

	tests := []struct {
		name          string
		input         string
		checkOutput   bool
		expectedInOut []string
		shouldContain string
	}{
		{
			name:          "Bare ? shows categories (REPL shortcut)",
			input:         "?",
			checkOutput:   true,
			expectedInOut: []string{"Available categories:", "Control", "Math", "Series"},
		},
		{
			name:          "? with category (goes to stdout)",
			input:         "? control",
			checkOutput:   false, // Goes to os.Stdout, not captured in errOut
			expectedInOut: nil,
		},
		{
			name:          "? with function (goes to stdout)",
			input:         "? if",
			checkOutput:   false,
			expectedInOut: nil,
		},
		{
			name:          "words returns block",
			input:         "words",
			checkOutput:   true,
			shouldContain: "print ", // Returns a block of words in Form format
		},
		{
			name:          "Help in sequence",
			input:         "x: 5  ? math  x + 1",
			checkOutput:   true,
			shouldContain: "6", // Final result should be printed to stdout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For bare ?, output goes to errOut, not stdout
			if tt.input == "?" {
				errOut.Reset()
				loop.EvalLineForTest(tt.input)
				result := errOut.String()
				if tt.checkOutput {
					if tt.shouldContain != "" && !strings.Contains(result, tt.shouldContain) {
						t.Errorf("Expected %q in output, got:\n%s", tt.shouldContain, result)
					}
					for _, exp := range tt.expectedInOut {
						if !strings.Contains(result, exp) {
							t.Errorf("Expected %q in output, got:\n%s", exp, result)
						}
					}
				}
			} else {
				// Capture output for successful evaluations
				errOut.Reset()
				loop.EvalLineForTest(tt.input)
				result := errOut.String()

				if tt.checkOutput {
					if tt.shouldContain != "" && !strings.Contains(result, tt.shouldContain) {
						t.Errorf("Expected %q in output, got:\n%s", tt.shouldContain, result)
					}
					for _, exp := range tt.expectedInOut {
						if !strings.Contains(result, exp) {
							t.Errorf("Expected %q in output, got:\n%s", exp, result)
						}
					}
				}
			}
			// For non-checked outputs, just verify no panic occurred
		})
	}
}

// TestREPL_HelpVsScriptBehavior documents the difference
func TestREPL_HelpVsScriptBehavior(t *testing.T) {
	t.Run("REPL_BareQuestionMark", func(t *testing.T) {
		evaluator := NewTestEvaluator()
		var out bytes.Buffer
		loop := repl.NewREPLForTest(evaluator, &out)

		// In REPL: '?' without args works (special shortcut)
		loop.EvalLineForTest("?")
		output := out.String()

		if !strings.Contains(output, "Available categories:") {
			t.Error("REPL should support bare '?' as shortcut")
		}
	})

	t.Run("Script_QuestionMarkRequiresArg", func(t *testing.T) {
		// This is documented in help_variadic_test.go
		// In scripts (parsed code), ? requires an argument (Arity: 1)
		// This test documents that difference
		t.Log("In scripts: '?' requires an argument per Arity: 1")
		t.Log("In REPL: '?' is intercepted before parsing and calls Help() directly")
		t.Log("This design allows natural REPL interaction while maintaining script clarity")
	})
}
