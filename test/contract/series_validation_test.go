package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// TestSeriesActionExcessiveArgs tests that series actions reject excessive arguments.
func TestSeriesActionExcessiveArgs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
	}{
		{
			name:      "first with too many args",
			input:     "first [1 2 3] 999",
			wantErrID: "arg-count",
		},
		{
			name:      "last with too many args",
			input:     "last [1 2 3] 999",
			wantErrID: "arg-count",
		},
		{
			name:      "append with too many args",
			input:     "append [1 2] 3 4",
			wantErrID: "arg-count",
		},
		{
			name:      "insert with too many args",
			input:     "insert [1 2] 3 4",
			wantErrID: "arg-count",
		},
		{
			name:      "length? with too many args",
			input:     "length? [1 2] 3",
			wantErrID: "arg-count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			_, evalErr := e.DoBlock(tokens)

			if evalErr == nil {
				t.Fatal("Expected error but got nil")
			}

			vErr, ok := evalErr.(*verror.Error)
			if !ok {
				t.Fatalf("Expected *verror.Error, got %T", evalErr)
			}

			if vErr.ID != tt.wantErrID {
				t.Errorf("Expected error ID %s, got %s", tt.wantErrID, vErr.ID)
			}

			t.Logf("Error message: %s", vErr.Error())
		})
	}
}
