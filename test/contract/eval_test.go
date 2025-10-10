package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestLiteralEvaluation tests that literal values evaluate to themselves.
// Contract: Literals (integers, strings, logic, none) return self without modification.
//
// TDD: This test is written FIRST and will FAIL until evaluator is implemented.
func TestLiteralEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    value.Value
		expected value.Value
	}{
		{
			name:     "integer literal",
			input:    value.IntVal(42),
			expected: value.IntVal(42),
		},
		{
			name:     "negative integer",
			input:    value.IntVal(-100),
			expected: value.IntVal(-100),
		},
		{
			name:     "zero",
			input:    value.IntVal(0),
			expected: value.IntVal(0),
		},
		{
			name:     "string literal",
			input:    value.StrVal("hello"),
			expected: value.StrVal("hello"),
		},
		{
			name:     "empty string",
			input:    value.StrVal(""),
			expected: value.StrVal(""),
		},
		{
			name:     "logic true",
			input:    value.LogicVal(true),
			expected: value.LogicVal(true),
		},
		{
			name:     "logic false",
			input:    value.LogicVal(false),
			expected: value.LogicVal(false),
		},
		{
			name:     "none",
			input:    value.NoneVal(),
			expected: value.NoneVal(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()

			result, err := e.Do_Next(tt.input)

			if err != nil {
				t.Errorf("Do_Next(%v) unexpected error: %v", tt.input, err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Do_Next(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestBlockEvaluation tests that blocks evaluate to themselves (deferred evaluation).
// Contract: Block values return self without evaluating contents.
func TestBlockEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    value.Value
		expected value.Value
	}{
		{
			name:     "empty block",
			input:    value.BlockVal([]value.Value{}),
			expected: value.BlockVal([]value.Value{}),
		},
		{
			name: "block with integers",
			input: value.BlockVal([]value.Value{
				value.IntVal(1),
				value.IntVal(2),
				value.IntVal(3),
			}),
			expected: value.BlockVal([]value.Value{
				value.IntVal(1),
				value.IntVal(2),
				value.IntVal(3),
			}),
		},
		{
			name: "block with unevaluated expression",
			input: value.BlockVal([]value.Value{
				value.IntVal(1),
				value.WordVal("+"),
				value.IntVal(2),
			}),
			// Block returns self - does NOT evaluate to 3
			expected: value.BlockVal([]value.Value{
				value.IntVal(1),
				value.WordVal("+"),
				value.IntVal(2),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()

			result, err := e.Do_Next(tt.input)

			if err != nil {
				t.Errorf("Do_Next(%v) unexpected error: %v", tt.input, err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Do_Next(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParenEvaluation tests that parens evaluate their contents immediately.
// Contract: Paren values evaluate contents and return the result.
func TestParenEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    value.Value
		expected value.Value
	}{
		{
			name:     "empty paren",
			input:    value.ParenVal([]value.Value{}),
			expected: value.NoneVal(), // Empty block returns none
		},
		{
			name: "paren with single value",
			input: value.ParenVal([]value.Value{
				value.IntVal(42),
			}),
			expected: value.IntVal(42),
		},
		{
			name: "paren evaluates like block with extra tokens",
			input: value.ParenVal([]value.Value{
				value.WordVal("+"),
				value.IntVal(1),
				value.IntVal(2),
				value.IntVal(4),
			}),
			expected: value.IntVal(4),
		},
		// More complex tests require arithmetic implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()

			result, err := e.Do_Next(tt.input)

			if err != nil {
				t.Errorf("Do_Next(%v) unexpected error: %v", tt.input, err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Do_Next(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestWordEvaluation tests that words resolve to bound values.
// Contract: Words look up values in current frame, error if unbound.
func TestWordEvaluation(t *testing.T) {
	tests := []struct {
		name      string
		setupWord string      // Word to bind before test
		setupVal  value.Value // Value to bind to word
		input     value.Value // Word to evaluate
		expected  value.Value // Expected result
		wantErr   bool
	}{
		{
			name:      "bound word",
			setupWord: "x",
			setupVal:  value.IntVal(10),
			input:     value.WordVal("x"),
			expected:  value.IntVal(10),
			wantErr:   false,
		},
		{
			name:     "unbound word error",
			input:    value.WordVal("undefined"),
			expected: value.NoneVal(),
			wantErr:  true, // Should error: no value for word
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()

			// Set up frame if needed
			if tt.setupWord != "" {
				f := frame.NewFrame(frame.FrameFunctionArgs, -1)
				f.Bind(tt.setupWord, tt.setupVal)
				e.Frames = append(e.Frames, f)
			}

			result, err := e.Do_Next(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Do_Next(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Do_Next(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSetWordEvaluation tests that set-words bind values.
// Contract: Set-word evaluates next expression and binds result to word.
func TestSetWordEvaluation(t *testing.T) {
	tests := []struct {
		name      string
		sequence  []value.Value // Sequence containing set-word and value
		checkWord string        // Word to check after evaluation
		expected  value.Value   // Expected value bound to word
	}{
		{
			name: "set integer",
			sequence: []value.Value{
				value.SetWordVal("x"),
				value.IntVal(42),
			},
			checkWord: "x",
			expected:  value.IntVal(42),
		},
		{
			name: "set string",
			sequence: []value.Value{
				value.SetWordVal("name"),
				value.StrVal("Alice"),
			},
			checkWord: "name",
			expected:  value.StrVal("Alice"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()

			// Evaluate sequence (set-word will bind the next value)
			result, err := e.Do_Blk(tt.sequence)

			if err != nil {
				t.Errorf("Do_Blk(%v) unexpected error: %v", tt.sequence, err)
				return
			}

			// Result should be the bound value
			if !result.Equals(tt.expected) {
				t.Errorf("Do_Blk(%v) = %v, want %v", tt.sequence, result, tt.expected)
			}

			// Verify word was bound in frame
			if len(e.Frames) == 0 {
				t.Errorf("No frame created after set-word evaluation")
				return
			}

			boundValue, ok := e.Frames[0].Get(tt.checkWord)
			if !ok {
				t.Errorf("Word %s not bound after set-word evaluation", tt.checkWord)
				return
			}

			if !boundValue.Equals(tt.expected) {
				t.Errorf("Bound value for %s = %v, want %v", tt.checkWord, boundValue, tt.expected)
			}
		})
	}
}
