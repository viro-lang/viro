package contract

import (
	"strings"
	"testing"
)

// TestRefinementPlacement verifies that refinements can appear in any position
// between function arguments, matching the behavior of the Lua implementation.
//
// Contract: Refinements should be readable before each positional argument
// and after all positional arguments, allowing flexible placement like:
// fn arg1 --ref1 arg2 --ref2 value arg3 --ref3
func TestRefinementPlacement(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "refinement before first argument",
			script: `
				myfn: fn [a b --verbose] [
					when verbose [
						print ["Adding" (a) "and" (b)]
					]
					a + b
				]
				myfn --verbose 10 20
			`,
			expected: "30",
		},
		{
			name: "refinement between arguments",
			script: `
				myfn: fn [a b c --debug] [
					when debug [print ["Debug:" (a) (b) (c)]]
					a + b + c
				]
				myfn 5 --debug 10 15
			`,
			expected: "30",
		},
		{
			name: "refinements interspersed with arguments",
			script: `
				myfn: fn [a b c --x --y []] [
					result: a + b + c
					when x [result: result * 2]
					when y [result: result + y]
					result
				]
				myfn 1 --x 2 --y 100 3
			`,
			expected: "112", // ((1 + 2 + 3) * 2) + 100 = 112
		},
		{
			name: "refinement after all arguments",
			script: `
				myfn: fn [a b --verbose] [
					result: a * b
					when verbose [print ["Result:" (result)]]
					result
				]
				myfn 7 8 --verbose
			`,
			expected: "56",
		},
		{
			name: "multiple refinements after arguments",
			script: `
				myfn: fn [a --double --triple] [
					result: a
					when double [result: result * 2]
					when triple [result: result * 3]
					result
				]
				myfn 5 --double --triple
			`,
			expected: "30", // 5 * 2 * 3 = 30
		},
		{
			name: "refinement with value between arguments",
			script: `
				myfn: fn [a b --scale []] [
					result: a + b
					when scale [result: result * scale]
					result
				]
				myfn 3 --scale 10 7
			`,
			expected: "100", // (3 + 7) * 10 = 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.script)
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			got := result.String()
			if got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestRefinementPlacementErrors verifies error handling for refinement placement
func TestRefinementPlacementErrors(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		expectError string
	}{
		{
			name: "unknown refinement",
			script: `
				myfn: fn [a b] [a + b]
				myfn 1 --unknown 2
			`,
			expectError: "Unknown refinement",
		},
		{
			name: "duplicate refinement",
			script: `
				myfn: fn [a b --verbose] [a + b]
				myfn --verbose 1 --verbose 2
			`,
			expectError: "Duplicate refinement",
		},
		{
			name: "refinement without required value",
			script: `
				myfn: fn [a --scale []] [a * scale]
				myfn 5 --scale
			`,
			expectError: "requires a value",
		},
		{
			name: "missing required argument due to refinement position",
			script: `
				myfn: fn [a b c] [a + b + c]
				myfn 1 --unknown 2
			`,
			expectError: "Unknown refinement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.script)
			if err == nil {
				t.Fatalf("Expected error containing '%s', but got no error", tt.expectError)
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.expectError) {
				t.Errorf("Expected error containing '%s', got: %s", tt.expectError, errMsg)
			}
		})
	}
}

// TestInfixWithRefinements verifies that infix operators work correctly with refinements
func TestInfixWithRefinements(t *testing.T) {
	t.Skip("Infix operators with refinements are not yet implemented")

	// This test documents expected behavior for future implementation
	script := `
		customAdd: fn [a b --verbose] [
			result: a + b
			if verbose [print ["Adding:" a "+" b "=" result]]
			result
		]
		5 customAdd --verbose 10
	`

	result, err := Evaluate(script)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}

	expected := "15"
	if result.String() != expected {
		t.Errorf("Expected %s, got %s", expected, result.String())
	}
}
