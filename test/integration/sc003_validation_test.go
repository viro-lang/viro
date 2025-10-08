package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC003_ErrorMessageUsability validates success criterion SC-003:
// Error messages include sufficient context (error category, expression location,
// relevant values) that users can diagnose and fix issues in under 2 minutes
// for common errors
//
// This is a structural validation - we verify error messages contain required
// information. Actual user testing would require manual validation.
func TestSC003_ErrorMessageUsability(t *testing.T) {
	evaluator := eval.NewEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	errorTests := []struct {
		name             string
		input            string
		expectedCategory string
		expectedElements []string
	}{
		{
			name:             "Undefined word error",
			input:            "undefined-variable",
			expectedCategory: "Script Error",
			expectedElements: []string{"undefined-variable", "No value"},
		},
		{
			name:             "Division by zero",
			input:            "10 / 0",
			expectedCategory: "Math Error",
			expectedElements: []string{"Division by zero", "/", "10", "0"},
		},
		{
			name:             "Type mismatch in arithmetic",
			input:            `5 + "hello"`,
			expectedCategory: "Script Error",
			expectedElements: []string{"type", "integer", "string"},
		},
		{
			name:             "Wrong number of arguments",
			input:            "+ 5",
			expectedCategory: "Script Error",
			expectedElements: []string{"argument", "+"},
		},
		{
			name:             "Empty series access",
			input:            "first []",
			expectedCategory: "Script Error",
			expectedElements: []string{"empty", "first"},
		},
	}

	passedTests := 0
	
	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			out.Reset()
			loop.EvalLineForTest(tt.input)
			errorOutput := out.String()

			// Verify error category is present
			if !strings.Contains(errorOutput, "**") || !strings.Contains(errorOutput, "Error") {
				t.Errorf("Error should have category header with **, got: %s", errorOutput)
				return
			}

			if !strings.Contains(errorOutput, tt.expectedCategory) {
				t.Errorf("Expected category %q in error, got: %s", tt.expectedCategory, errorOutput)
				return
			}

			// Verify required context elements are present
			missingElements := []string{}
			for _, elem := range tt.expectedElements {
				if !strings.Contains(errorOutput, elem) {
					missingElements = append(missingElements, elem)
				}
			}

			if len(missingElements) > 0 {
				t.Errorf("Error missing context elements %v, got: %s", missingElements, errorOutput)
				return
			}

			t.Logf("Error message includes: category, context, relevant values ✓")
			passedTests++
		})
	}

	t.Logf("SC-003 VALIDATION: %d/%d error messages have sufficient context", passedTests, len(errorTests))

	if passedTests < len(errorTests) {
		t.Errorf("SC-003: Some error messages lack sufficient context")
	} else {
		t.Logf("SC-003 SUCCESS: Error messages provide diagnostic context")
	}

	// Additional checks for error structure
	t.Run("Error format consistency", func(t *testing.T) {
		out.Reset()
		loop.EvalLineForTest("10 / 0")
		errorOutput := out.String()

		// Should have clear structure
		checks := map[string]bool{
			"Has error marker (*)":       strings.Contains(errorOutput, "*"),
			"Has category":                strings.Contains(errorOutput, "Error"),
			"Has descriptive message":     len(errorOutput) > 20,
			"Not just error code":         !strings.Contains(errorOutput, "Error 300") && strings.Contains(errorOutput, "Division"),
		}

		allPassed := true
		for check, passed := range checks {
			if !passed {
				t.Errorf("%s: FAILED", check)
				allPassed = false
			} else {
				t.Logf("%s: PASSED", check)
			}
		}

		if allPassed {
			t.Logf("Error format is consistent and user-friendly")
		}
	})
}
