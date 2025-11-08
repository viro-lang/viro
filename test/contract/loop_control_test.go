package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestLoopControl_BreakBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "break in loop returns none",
			input:    "loop 3 [break]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "break exits early - counter check",
			input:    "x: 0\nloop 10 [x: x + 1\nwhen (= x 3) [break]]\nx",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "break on first iteration",
			input:    "x: 0\nloop 5 [when (= x 0) [break]\nx: x + 1]\nx",
			expected: value.NewIntVal(0),
			wantErr:  false,
		},
		{
			name:     "break in while returns none",
			input:    "x: 0\nwhile [x < 10] [x: x + 1\nwhen (= x 3) [break]]\nx",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "break in foreach returns none",
			input:    "x: 0\nforeach [1 2 3 4 5] 'val [x: x + val\nwhen (= x 6) [break]]\nx",
			expected: value.NewIntVal(6),
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
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoopControl_ContinueBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "continue skips iteration in loop",
			input:    "x: 0\nloop 3 [x: x + 1\ncontinue\nx: x + 100]\nx",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "continue in while re-evaluates condition",
			input:    "x: 0\nwhile [x < 3] [x: x + 1\ncontinue\nx: x + 100]\nx",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "continue in foreach processes next element",
			input:    "x: 0\nforeach [1 2 3] 'val [x: x + val\ncontinue\nx: x + 100]\nx",
			expected: value.NewIntVal(6),
			wantErr:  false,
		},
		{
			name:     "continue on selective iterations",
			input:    "x: 0\nloop 5 --with-index 'i [when (= (mod i 2) 0) [continue]\nx: x + 1]\nx",
			expected: value.NewIntVal(2),
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
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoopControl_Nested(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "break in inner loop only",
			input: `
				outer: 0
				inner: 0
				loop 3 [
					outer: outer + 1
					loop 3 [
						inner: inner + 1
						break
					]
				]
				outer
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "inner loop completes fully, outer breaks",
			input: `
				x: 0
				loop 3 [
					loop 2 [
						x: x + 1
					]
					when (= x 4) [break]
				]
				x
			`,
			expected: value.NewIntVal(4),
			wantErr:  false,
		},
		{
			name: "continue in inner loop only",
			input: `
				x: 0
				loop 2 [
					loop 3 [
						x: x + 1
						continue
						x: x + 100
					]
				]
				x
			`,
			expected: value.NewIntVal(6),
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
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoopControl_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errID   string
	}{
		{
			name:    "break outside loop",
			input:   "break",
			wantErr: true,
			errID:   verror.ErrIDBreakOutsideLoop,
		},
		{
			name:    "continue outside loop",
			input:   "continue",
			wantErr: true,
			errID:   verror.ErrIDContinueOutsideLoop,
		},
		{
			name:    "break in function called from loop - boundary blocks it",
			input:   "f: fn [] [break]\nloop 3 [f]",
			wantErr: true,
			errID:   verror.ErrIDBreakOutsideLoop,
		},
		{
			name:    "continue in function called from loop - boundary blocks it",
			input:   "f: fn [] [continue]\nloop 3 [f]",
			wantErr: true,
			errID:   verror.ErrIDContinueOutsideLoop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error but got none, result: %v", result)
				return
			}

			if verr, ok := err.(*verror.Error); ok {
				if verr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, verr.ID)
				}
			} else {
				t.Errorf("Expected verror.Error, got %T", err)
			}
		})
	}
}

func TestLoopControl_TransparentBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "break in do block - works (no boundary)",
			input:    "x: 0\nloop 10 [x: x + 1\ndo [when (= x 3) [break]]]\nx",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "continue in do block - works (no boundary)",
			input:    "x: 0\nloop 3 [x: x + 1\ndo [continue]\nx: x + 100]\nx",
			expected: value.NewIntVal(3),
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
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoopControl_WithIndex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "break with --with-index refinement",
			input:    "sum: 0\nloop 10 --with-index 'i [sum: sum + i\nwhen (= i 3) [break]]\nsum",
			expected: value.NewIntVal(6),
			wantErr:  false,
		},
		{
			name:     "continue with --with-index refinement",
			input:    "sum: 0\nloop 5 --with-index 'i [when (= (mod i 2) 0) [continue]\nsum: sum + i]\nsum",
			expected: value.NewIntVal(4),
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
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoopControl_MultiLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "break --levels 2 in nested loop",
			input: `
				x: 0
				loop 3 [
					loop 3 [
						x: x + 1
						when (= x 2) [break --levels 2]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(2),
			wantErr:  false,
		},
		{
			name: "break --levels 3 in triple-nested loop",
			input: `
				x: 0
				loop 3 [
					loop 3 [
						loop 3 [
							x: x + 1
							when (= x 3) [break --levels 3]
						]
						x: x + 100
					]
					x: x + 1000
				]
				x
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "continue --levels 2 in nested loop",
			input: `
				x: 0
				loop 3 --with-index 'i [
					loop 3 --with-index 'j [
						x: x + 1
						when (and (= i 0) (= j 2)) [continue --levels 2]
						x: x + 10
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(289),
			wantErr:  false,
		},
		{
			name: "break --levels 1 is same as break",
			input: `
				x: 0
				loop 3 [
					loop 3 [
						x: x + 1
						when (= x 2) [break --levels 1]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(308),
			wantErr:  false,
		},
		{
			name: "continue --levels 1 is same as continue",
			input: `
				x: 0
				loop 3 [
					x: x + 1
					continue --levels 1
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "break --levels in while loops",
			input: `
				x: 0
				while [x < 10] [
					while [x < 10] [
						x: x + 1
						when (= x 3) [break --levels 2]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "break --levels in foreach",
			input: `
				x: 0
				foreach [1 2 3] 'a [
					foreach [10 20 30] 'b [
						x: x + a + b
						when (= x 32) [break --levels 2]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(32),
			wantErr:  false,
		},
		{
			name: "break --levels with transparent blocks",
			input: `
				x: 0
				loop 3 [
					loop 3 [
						x: x + 1
						do [when (= x 2) [break --levels 2]]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(2),
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
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoopControl_MultiLevelErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errID   string
	}{
		{
			name:    "break --levels 0 is invalid",
			input:   "loop 3 [break --levels 0]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "break --levels -1 is invalid",
			input:   "loop 3 [break --levels -1]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "continue --levels 0 is invalid",
			input:   "loop 3 [continue --levels 0]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "break --levels requires integer",
			input:   "loop 3 [break --levels \"two\"]",
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "continue --levels requires integer",
			input:   "loop 3 [continue --levels \"two\"]",
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "break --levels 2 in function crosses boundary",
			input:   "loop 3 [loop 3 [f: fn [] [break --levels 2]\nf]]",
			wantErr: true,
			errID:   verror.ErrIDBreakOutsideLoop,
		},
		{
			name:    "continue --levels 2 in function crosses boundary",
			input:   "loop 3 [loop 3 [f: fn [] [continue --levels 2]\nf]]",
			wantErr: true,
			errID:   verror.ErrIDContinueOutsideLoop,
		},
		{
			name:    "break --levels 3 overshoots loop depth",
			input:   "loop 2 [loop 2 [break --levels 3]]",
			wantErr: true,
			errID:   verror.ErrIDBreakOutsideLoop,
		},
		{
			name:    "continue --levels 3 overshoots loop depth",
			input:   "loop 2 [loop 2 [continue --levels 3]]",
			wantErr: true,
			errID:   verror.ErrIDContinueOutsideLoop,
		},
		{
			name:    "multi-level continue in while/while loops",
			input:   "x: 0\nwhile [x < 10] [while [x < 10] [x: x + 1\nwhen (= x 3) [continue --levels 2]\nx: x + 10]]\nx",
			wantErr: false,
			errID:   "",
		},
		{
			name:    "multi-level continue in foreach/foreach loops",
			input:   "x: 0\nforeach [1 2 3] 'a [foreach [10 20 30] 'b [x: x + 1\nwhen (= x 3) [continue --levels 2]\nx: x + 10]]\nx",
			wantErr: false,
			errID:   "",
		},
		{
			name:    "mixed loop types - break --levels 2",
			input:   "x: 0\nloop 3 [foreach [1 2 3] 'a [x: x + 1\nwhen (= x 2) [break --levels 2]]]\nx",
			wantErr: false,
			errID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error but got none, result: %v", result)
				return
			}

			if verr, ok := err.(*verror.Error); ok {
				if verr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, verr.ID)
				}
			} else {
				t.Errorf("Expected verror.Error, got %T", err)
			}
		})
	}
}
