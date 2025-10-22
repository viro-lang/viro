package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// TestActionNoImpl tests the action-no-impl error when an action is called
// on a type that doesn't support it.
// Contract: User Story 3 - T050
func TestActionNoImpl(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
		wantMsg   string // substring to check in error message
	}{
		{
			name:      "first on integer",
			input:     "first 42",
			wantErrID: "action-no-impl",
			wantMsg:   "first",
		},
		{
			name:      "last on integer",
			input:     "last 42",
			wantErrID: "action-no-impl",
			wantMsg:   "last",
		},
		{
			name:      "append on integer",
			input:     "append 42 1",
			wantErrID: "action-no-impl",
			wantMsg:   "append",
		},
		{
			name:      "insert on logic",
			input:     "insert true false",
			wantErrID: "action-no-impl",
			wantMsg:   "insert",
		},
		{
			name:      "length? on function",
			input:     "length? fn [x] [x]",
			wantErrID: "action-no-impl",
			wantMsg:   "length?",
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

// TestActionWrongArity tests wrong-arity errors for actions.
// Contract: User Story 3 - T051
func TestActionWrongArity(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
	}{
		{
			name:      "first with no args",
			input:     "first",
			wantErrID: "arg-count",
		},
		{
			name:      "last with no args",
			input:     "last",
			wantErrID: "arg-count",
		},
		{
			name:      "append with one arg",
			input:     "append [1 2]",
			wantErrID: "arg-count",
		},
		{
			name:      "insert with one arg",
			input:     "insert [1 2]",
			wantErrID: "arg-count",
		},
		{
			name:      "length? with no args",
			input:     "length?",
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

// TestActionEmptySeries tests out-of-bounds errors when accessing empty series.
// Contract: User Story 3 - T052
func TestActionEmptySeries(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
	}{
		{
			name:      "first on empty block",
			input:     "first []",
			wantErrID: "out-of-bounds",
		},
		{
			name:      "first on empty string",
			input:     `first ""`,
			wantErrID: "out-of-bounds",
		},
		{
			name:      "last on empty block",
			input:     "last []",
			wantErrID: "out-of-bounds",
		},
		{
			name:      "last on empty string",
			input:     `last ""`,
			wantErrID: "out-of-bounds",
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

// TestActionTypeMismatch tests type-mismatch errors for string operations.
// Contract: User Story 3 - T053
func TestActionTypeMismatch(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
	}{
		{
			name:      "append integer to string",
			input:     `append "hello" 42`,
			wantErrID: "type-mismatch",
		},
		{
			name:      "append block to string",
			input:     `append "hello" [1 2]`,
			wantErrID: "type-mismatch",
		},
		{
			name:      "insert integer to string",
			input:     `insert "hello" 42`,
			wantErrID: "type-mismatch",
		},
		{
			name:      "insert block to string",
			input:     `insert "hello" [1 2]`,
			wantErrID: "type-mismatch",
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
