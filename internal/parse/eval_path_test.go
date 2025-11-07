package parse

import (
	"testing"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
)

func TestParsePathWithEvalSegment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		fail  bool
	}{
		{"simple eval", "data.(idx)", false},
		{"set-path eval", "data.(idx):", false},
		{"get-path eval", ":data.(idx)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := tokenize.NewTokenizer(tt.input)
			tokens, err := tokenizer.Tokenize()
			if err != nil {
				t.Fatalf("Tokenize error: %v", err)
			}

			t.Logf("Tokens: %+v", tokens)

			parser := NewParser(tokens, "test")
			values, _, err := parser.Parse()
			
			if tt.fail {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if len(values) == 0 {
				t.Fatalf("Expected values, got 0")
			}

			t.Logf("Parsed value: %+v (type=%d)", values[0], values[0].GetType())
		})
	}
}
