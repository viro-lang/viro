package contract

import (
	"testing"
)

// TestSelectWithDefaultRefinement tests the --default refinement for select native.
// This ensures that refinements are properly passed from evaluator to native functions
// after the unification of native and user function invocation.
func TestSelectWithDefaultRefinement(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
		isError  bool
	}{
		{
			name:     "select existing field without default",
			script:   "obj: object [x: 10 y: 20]\nselect obj 'x",
			expected: "10",
			isError:  false,
		},
		{
			name:     "select missing field without default returns none",
			script:   "obj: object [x: 10 y: 20]\nselect obj 'z",
			expected: "none",
			isError:  false,
		},
		{
			name:     "select missing field with default returns default value",
			script:   "obj: object [x: 10 y: 20]\nselect obj 'z --default 99",
			expected: "99",
			isError:  false,
		},
		{
			name:     "select existing field with default returns field value (not default)",
			script:   "obj: object [x: 10 y: 20]\nselect obj 'x --default 99",
			expected: "10",
			isError:  false,
		},
		{
			name:     "select with default using expression",
			script:   "obj: object [x: 10 y: 20]\nselect obj 'z --default (5 + 5)",
			expected: "10",
			isError:  false,
		},
		{
			name:     "select from block with default",
			script:   "data: ['name \"Alice\" 'age 30]\nselect data 'name",
			expected: `"Alice"`, // String values include quotes in String() output
			isError:  false,
		},
		{
			name:     "select missing key from block with default",
			script:   "data: ['name \"Alice\" 'age 30]\nselect data 'city --default \"Unknown\"",
			expected: `"Unknown"`, // String values include quotes in String() output
			isError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.script)

			if tt.isError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Mold() != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, result.Mold())
			}
		})
	}
}
