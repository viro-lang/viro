package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestRejoin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "empty block returns empty string",
			input:    `rejoin []`,
			expected: value.NewStrVal(""),
			wantErr:  false,
		},
		{
			name:     "single string",
			input:    `rejoin ["Hello"]`,
			expected: value.NewStrVal("Hello"),
			wantErr:  false,
		},
		{
			name:     "multiple strings with space",
			input:    `rejoin ["Hello" " " "World"]`,
			expected: value.NewStrVal("Hello World"),
			wantErr:  false,
		},
		{
			name:     "strings without explicit space",
			input:    `rejoin ["Hello" "World"]`,
			expected: value.NewStrVal("HelloWorld"),
			wantErr:  false,
		},
		{
			name:     "string with number",
			input:    `rejoin ["Number: " 42]`,
			expected: value.NewStrVal("Number: 42"),
			wantErr:  false,
		},
		{
			name:     "string with expression",
			input:    `rejoin ["Result: " (+ 10 5)]`,
			expected: value.NewStrVal("Result: 15"),
			wantErr:  false,
		},
		{
			name:     "mixed types",
			input:    `rejoin ["Value: " 3.14 " " true]`,
			expected: value.NewStrVal("Value: 3.14 true"),
			wantErr:  false,
		},
		{
			name:     "with infix expression",
			input:    `rejoin ["Sum: " 10 + 5]`,
			expected: value.NewStrVal("Sum: 15"),
			wantErr:  false,
		},
		{
			name:     "numbers only",
			input:    `rejoin [1 2 3]`,
			expected: value.NewStrVal("123"),
			wantErr:  false,
		},
		{
			name:     "with none value",
			input:    `rejoin ["Value: " none]`,
			expected: value.NewStrVal("Value: none"),
			wantErr:  false,
		},
		{
			name:     "non-block argument returns error",
			input:    `rejoin "not a block"`,
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name:     "wrong arity no args returns error",
			input:    `rejoin`,
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v (%s), got %v (%s)",
					tt.expected, tt.expected.Form(),
					result, result.Form())
			}
		})
	}
}
