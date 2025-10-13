package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
)

// TestActionDispatchBasics tests fundamental action dispatch behavior.
// Contract: action-dispatch.md
func TestActionDispatchBasics(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errID   string
	}{
		{
			name:  "dispatch to block first",
			input: "first [1 2 3]",
			want:  "1",
		},
		{
			name:  "dispatch to string first",
			input: `first "hello"`,
			want:  `"h"`,
		},
		{
			name:    "unsupported type error",
			input:   "first 42",
			wantErr: true,
			errID:   "action-no-impl",
		},
		{
			name:    "arity error - no arguments",
			input:   "first",
			wantErr: true,
			errID:   "arg-count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			result, evalErr := e.Do_Blk(tokens)

			if tt.wantErr {
				if evalErr == nil {
					t.Errorf("Expected error with ID %s, got nil", tt.errID)
					return
				}
				if evalErr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, evalErr.ID)
				}
				return
			}

			if evalErr != nil {
				t.Errorf("Unexpected error: %v", evalErr)
				return
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestActionShadowing tests that local bindings can shadow actions.
// Contract: action-dispatch.md Test 8
func TestActionShadowing(t *testing.T) {
	e := eval.NewEvaluator()

	// Shadow the action with a local function
	input := `
		first: fn [x] [x * 2]
		first 5
	`

	tokens, parseErr := parse.Parse(input)
	if parseErr != nil {
		t.Fatalf("Parse error: %v", parseErr)
	}

	result, evalErr := e.Do_Blk(tokens)
	if evalErr != nil {
		t.Fatalf("Unexpected error: %v", evalErr)
	}

	// Should use the local function, not the action
	if val, ok := result.AsInteger(); !ok || val != 10 {
		t.Errorf("Expected 10, got %s", result.String())
	}
}

// TestActionMultipleArguments tests actions with multiple parameters.
// Contract: action-dispatch.md Test 3
func TestActionMultipleArguments(t *testing.T) {
	e := eval.NewEvaluator()

	input := `
		b: [1 2]
		append b 3
		b
	`

	tokens, parseErr := parse.Parse(input)
	if parseErr != nil {
		t.Fatalf("Parse error: %v", parseErr)
	}

	result, evalErr := e.Do_Blk(tokens)
	if evalErr != nil {
		t.Fatalf("Unexpected error: %v", evalErr)
	}

	// Block should be modified in-place
	blk, ok := result.AsBlock()
	if !ok {
		t.Fatalf("Expected block, got %s", result.Type.String())
	}

	if len(blk.Elements) != 3 {
		t.Errorf("Expected block length 3, got %d", len(blk.Elements))
	}

	// Check last element is 3
	lastVal, ok := blk.Elements[2].AsInteger()
	if !ok || lastVal != 3 {
		t.Errorf("Expected last element to be 3, got %s", blk.Elements[2].String())
	}
}
