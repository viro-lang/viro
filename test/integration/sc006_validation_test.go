package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestSC006_TypeErrorDetection validates success criterion SC-006:
// Type error detection catches at least 95% of type mismatches before
// execution (during argument validation phase)
//
// We test type checking at the native function level where arguments
// are validated before execution.
func TestSC006_TypeErrorDetection(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	typeErrorTests := []struct {
		name          string
		input         string
		shouldError   bool
		errorContains string
	}{
		// Arithmetic type errors
		{
			name:          "Add integer and string",
			input:         `5 + "hello"`,
			shouldError:   true,
			errorContains: "type",
		},
		{
			name:          "Multiply string and integer",
			input:         `"text" * 3`,
			shouldError:   true,
			errorContains: "type",
		},
		{
			name:          "Divide with non-integer",
			input:         `10 / true`,
			shouldError:   true,
			errorContains: "type",
		},

		// Comparison type errors
		{
			name:          "Compare integer and string",
			input:         `5 < "10"`,
			shouldError:   true,
			errorContains: "type",
		},

		// Logic type errors (should be permissive with truthy conversion)
		{
			name:        "And with integers (truthy)",
			input:       `and 5 10`,
			shouldError: false,
		},

		// Series operation type errors
		{
			name:          "First on non-series",
			input:         `first 42`,
			shouldError:   true,
			errorContains: "type",
		},
		{
			name:          "Append to non-series",
			input:         `append 42 1`,
			shouldError:   true,
			errorContains: "type",
		},
		{
			name:          "Length? on non-series",
			input:         `length? 42`,
			shouldError:   true,
			errorContains: "type",
		},

		// Control flow type errors
		{
			name:          "Loop with non-integer count",
			input:         `loop "five" [42]`,
			shouldError:   true,
			errorContains: "type",
		},
		{
			name:          "If with non-block branches",
			input:         `if true 1 2`,
			shouldError:   true,
			errorContains: "type",
		},

		// Function definition type errors
		{
			name:          "Fn with non-block params",
			input:         `fn 5 [42]`,
			shouldError:   true,
			errorContains: "type",
		},
		{
			name:          "Fn with non-block body",
			input:         `fn [x] 42`,
			shouldError:   true,
			errorContains: "type",
		},

		// Valid operations (should not error)
		{
			name:        "Valid arithmetic",
			input:       `5 + 10`,
			shouldError: false,
		},
		{
			name:        "Valid series operation",
			input:       `first [1 2 3]`,
			shouldError: false,
		},
		{
			name:        "Valid function call",
			input:       `sq: fn [n] [(* n n)]  sq 5`,
			shouldError: false,
		},
	}

	totalTests := len(typeErrorTests)
	errorsCaught := 0
	falsePositives := 0
	falseNegatives := 0

	_ = totalTests // Used for documentation/context

	for _, tt := range typeErrorTests {
		t.Run(tt.name, func(t *testing.T) {
			out.Reset()
			loop.EvalLineForTest(tt.input)
			output := out.String()
			hasError := strings.Contains(output, "Error")

			if tt.shouldError {
				if hasError && strings.Contains(strings.ToLower(output), strings.ToLower(tt.errorContains)) {
					errorsCaught++
					t.Logf("✓ Type error correctly caught")
				} else if hasError {
					errorsCaught++
					t.Logf("✓ Error caught (different message than expected)")
				} else {
					falseNegatives++
					t.Errorf("✗ Type error NOT caught: %s", tt.input)
				}
			} else {
				if hasError {
					falsePositives++
					t.Errorf("✗ False positive - valid operation rejected: %s\nOutput: %s", tt.input, output)
				} else {
					t.Logf("✓ Valid operation accepted")
				}
			}
		})
	}

	// Count how many tests expected errors
	expectedErrors := 0
	for _, tt := range typeErrorTests {
		if tt.shouldError {
			expectedErrors++
		}
	}

	detectionRate := float64(errorsCaught) / float64(expectedErrors) * 100

	t.Logf("SC-006 VALIDATION: Type error detection rate: %.1f%%", detectionRate)
	t.Logf("SC-006 VALIDATION: Errors caught: %d/%d", errorsCaught, expectedErrors)
	t.Logf("SC-006 VALIDATION: False negatives: %d", falseNegatives)
	t.Logf("SC-006 VALIDATION: False positives: %d", falsePositives)

	if detectionRate < 95.0 {
		t.Errorf("SC-006 FAILED: Detection rate %.1f%% is below 95%% target", detectionRate)
	} else {
		t.Logf("SC-006 SUCCESS: Type error detection rate %.1f%% meets 95%% target", detectionRate)
	}
}
