package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestCharset_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(core.Value) bool
		wantErr  bool
	}{
		{
			name:  "charset from string",
			input: `charset "abc"`,
			expected: func(v core.Value) bool {
				bs, ok := value.AsBitsetValue(v)
				return ok && bs.Test('a') && bs.Test('b') && bs.Test('c') && !bs.Test('d')
			},
			wantErr: false,
		},
		{
			name:  "charset type",
			input: `type? charset "abc"`,
			expected: func(v core.Value) bool {
				word, ok := value.AsWordValue(v)
				return ok && word == "bitset!"
			},
			wantErr: false,
		},
		{
			name:  "charset empty",
			input: `charset ""`,
			expected: func(v core.Value) bool {
				bs, ok := value.AsBitsetValue(v)
				return ok && bs.IsEmpty()
			},
			wantErr: false,
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

			if !tt.expected(result) {
				t.Errorf("Result validation failed for input: %s, got: %v", tt.input, result.Mold())
			}
		})
	}
}

func TestCharset_Clone(t *testing.T) {
	input := `
		cs1: charset "abc"
		cs2: charset cs1
		type? cs2
	`
	result, err := Evaluate(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	word, ok := value.AsWordValue(result)
	if !ok || word != "bitset!" {
		t.Errorf("Expected bitset! type, got: %v", result.Mold())
	}
}
