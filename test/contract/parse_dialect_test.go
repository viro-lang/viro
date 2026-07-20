package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestParseDialect_StringLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "simple match",
			input:    `parse "hello" ["hello"]`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "no match",
			input:    `parse "hello" ["world"]`,
			expected: value.NewLogicVal(false),
			wantErr:  false,
		},
		{
			name:     "sequence match",
			input:    `parse "hello world" ["hello" " " "world"]`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "case insensitive by default",
			input:    `parse "Hello" ["hello"]`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
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
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}

func TestParseDialect_Block(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "datatype match",
			input:    `parse [1 2 3] [integer! integer! integer!]`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "mixed types",
			input:    `parse [1 "hello" 3] [integer! string! integer!]`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "type mismatch",
			input:    `parse [1 2 3] [string! string! string!]`,
			expected: value.NewLogicVal(false),
			wantErr:  false,
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
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}

func TestParseDialect_Charset(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "charset match",
			input: `
				cs: charset "abc"
				parse "abc" [cs cs cs]
			`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name: "charset no match",
			input: `
				cs: charset "abc"
				parse "abcd" [cs cs cs]
			`,
			expected: value.NewLogicVal(false),
			wantErr:  false,
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
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}

func TestParseDialect_Keywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "skip keyword",
			input:    `parse "abc" [skip "b" "c"]`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "end keyword success",
			input:    `parse "hello" ["hello" end]`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "end keyword failure",
			input:    `parse "hello world" ["hello" end]`,
			expected: value.NewLogicVal(false),
			wantErr:  false,
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
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}

func TestParseDialect_Alternation(t *testing.T) {
	input := `parse "hi" [["hello" | "hi"]]`
	result, err := Evaluate(input)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := value.NewLogicVal(true)
	if !result.Equals(expected) {
		t.Errorf("Expected %v, got %v", expected.Mold(), result.Mold())
	}
}
