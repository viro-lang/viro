package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// TestActionFirst tests the 'first' action on blocks and strings.
// Contract: series-actions.md - first
func TestActionFirst(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errID   string
	}{
		// Block tests
		{name: "block with integers", input: "first [1 2 3]", want: "1"},
		{name: "block with strings", input: `first ["a" "b"]`, want: `"a"`},
		{name: "nested blocks", input: "first [[1 2] [3 4]]", want: "[1 2]"},
		{name: "empty block", input: "first []", wantErr: true, errID: "out-of-bounds"},

		// String tests
		{name: "string", input: `first "hello"`, want: `"h"`},
		{name: "single char string", input: `first "a"`, want: `"a"`},
		{name: "empty string", input: `first ""`, wantErr: true, errID: "out-of-bounds"},

		// Error cases
		{name: "unsupported type", input: "first 42", wantErr: true, errID: "action-no-impl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			result, err := e.Do_Blk(tokens)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error with ID %s, got nil", tt.errID)
					return
				}
				evalErr := err.(*verror.Error)
				if evalErr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, evalErr.ID)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestActionLast tests the 'last' action on blocks and strings.
// Contract: series-actions.md - last
func TestActionLast(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errID   string
	}{
		// Block tests
		{name: "block with integers", input: "last [1 2 3]", want: "3"},
		{name: "block with strings", input: `last ["a" "b"]`, want: `"b"`},
		{name: "nested blocks", input: "last [[1 2] [3 4]]", want: "[3 4]"},
		{name: "empty block", input: "last []", wantErr: true, errID: "out-of-bounds"},

		// String tests
		{name: "string", input: `last "hello"`, want: `"o"`},
		{name: "single char string", input: `last "a"`, want: `"a"`},
		{name: "empty string", input: `last ""`, wantErr: true, errID: "out-of-bounds"},

		// Error cases
		{name: "unsupported type", input: "last 42", wantErr: true, errID: "action-no-impl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			result, err := e.Do_Blk(tokens)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error with ID %s, got nil", tt.errID)
					return
				}
				evalErr := err.(*verror.Error)
				if evalErr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, evalErr.ID)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestActionAppend tests the 'append' action on blocks and strings.
// Contract: series-actions.md - append
func TestActionAppend(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errID   string
	}{
		// Block tests
		{name: "append int to block", input: "b: [1 2]\nappend b 3\nb", want: "[1 2 3]"},
		{name: "append to empty block", input: "b: []\nappend b 'a\nb", want: `[a]`},
		{name: "append block to block", input: "b: [1]\nappend b [2 3]\nb", want: "[1 [2 3]]"},

		// String tests
		{name: "append string to string", input: "s: \"hel\"\nappend s \"lo\"\ns", want: `"hello"`},
		{name: "append to empty string", input: "s: \"\"\nappend s \"a\"\ns", want: `"a"`},
		{name: "string type mismatch", input: `append "test" 42`, wantErr: true, errID: "type-mismatch"},

		// Error cases
		{name: "unsupported type", input: "append 42 3", wantErr: true, errID: "action-no-impl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			result, err := e.Do_Blk(tokens)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error with ID %s, got nil", tt.errID)
					return
				}
				evalErr := err.(*verror.Error)
				if evalErr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, evalErr.ID)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestActionInsert tests the 'insert' action on blocks and strings.
// Contract: series-actions.md - insert
func TestActionInsert(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errID   string
	}{
		// Block tests
		{name: "insert int at beginning", input: "b: [2 3]\ninsert b 1\nb", want: "[1 2 3]"},
		{name: "insert into empty block", input: "b: []\ninsert b 'a\nb", want: `[a]`},
		{name: "insert block at beginning", input: "b: [3]\ninsert b [1 2]\nb", want: "[[1 2] 3]"},

		// String tests
		{name: "insert string at beginning", input: "s: \"orld\"\ninsert s \"W\"\ns", want: `"World"`},
		{name: "insert into empty string", input: "s: \"\"\ninsert s \"a\"\ns", want: `"a"`},
		{name: "string type mismatch", input: `insert "test" 42`, wantErr: true, errID: "type-mismatch"},

		// Error cases
		{name: "unsupported type", input: "insert 42 3", wantErr: true, errID: "action-no-impl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			result, err := e.Do_Blk(tokens)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error with ID %s, got nil", tt.errID)
					return
				}
				evalErr := err.(*verror.Error)
				if evalErr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, evalErr.ID)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestActionLength tests the 'length?' action on blocks and strings.
// Contract: series-actions.md - length?
func TestActionLength(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errID   string
	}{
		// Block tests
		{name: "block with 3 elements", input: "length? [1 2 3]", want: "3"},
		{name: "empty block", input: "length? []", want: "0"},
		{name: "nested blocks", input: "length? [[1 2] [3 4]]", want: "2"},

		// String tests
		{name: "string with 5 chars", input: `length? "hello"`, want: "5"},
		{name: "empty string", input: `length? ""`, want: "0"},
		{name: "single char string", input: `length? "a"`, want: "1"},

		// Error cases
		{name: "unsupported type", input: "length? 42", wantErr: true, errID: "action-no-impl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			result, err := e.Do_Blk(tokens)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error with ID %s, got nil", tt.errID)
					return
				}
				evalErr := err.(*verror.Error)
				if evalErr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, evalErr.ID)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}
