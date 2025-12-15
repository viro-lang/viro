package contract

import (
	"errors"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestSeries_Replace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		// String replacement - first occurrence only
		{
			name:  "string replace basic",
			input: `replace "hello world" "world" "Viro"`,
			want:  value.NewStrVal("hello Viro"),
		},
		{
			name:  "string replace pattern not found",
			input: `replace "hello" "xyz" "abc"`,
			want:  value.NewStrVal("hello"),
		},
		{
			name:  "string replace multiple occurrences first only",
			input: `replace "abc abc abc" "abc" "x"`,
			want:  value.NewStrVal("x abc abc"),
		},
		{
			name:  "string replace pattern at start",
			input: `replace "hello world" "hello" "hi"`,
			want:  value.NewStrVal("hi world"),
		},
		{
			name:  "string replace pattern at end",
			input: `replace "hello world" "world" "Viro"`,
			want:  value.NewStrVal("hello Viro"),
		},
		{
			name:  "string replace overlapping patterns",
			input: `replace "aaa" "aa" "b"`,
			want:  value.NewStrVal("ba"),
		},
		{
			name:  "string replace empty series",
			input: `replace "" "x" "y"`,
			want:  value.NewStrVal(""),
		},
		{
			name:  "string replace empty replacement",
			input: `replace "hello" "l" ""`,
			want:  value.NewStrVal("helo"),
		},
		{
			name:  "string replace single char pattern",
			input: `replace "hello" "e" "a"`,
			want:  value.NewStrVal("hallo"),
		},
		{
			name:  "string replace single char to multiple",
			input: `replace "hello" "e" "xyz"`,
			want:  value.NewStrVal("hxylzlo"),
		},
		{
			name:  "string replace multiple char pattern",
			input: `replace "hello world" "lo" "ya"`,
			want:  value.NewStrVal("heya world"),
		},
		{
			name:  "string replace pattern longer than series",
			input: `replace "hi" "hello" "x"`,
			want:  value.NewStrVal("hi"),
		},

		// String replacement - all occurrences
		{
			name:  "string replace all basic",
			input: `replace/all "abc abc abc" "abc" "x"`,
			want:  value.NewStrVal("x x x"),
		},
		{
			name:  "string replace all overlapping",
			input: `replace/all "aaa" "aa" "b"`,
			want:  value.NewStrVal("ba"),
		},
		{
			name:  "string replace all not found",
			input: `replace/all "hello" "xyz" "abc"`,
			want:  value.NewStrVal("hello"),
		},
		{
			name:  "string replace all at edges",
			input: `replace/all "aaa aaa aaa" "aaa" "b"`,
			want:  value.NewStrVal("b b b"),
		},
		{
			name:  "string replace all single char",
			input: `replace/all "hello" "l" "x"`,
			want:  value.NewStrVal("hexxo"),
		},

		// Block replacement - first occurrence only
		{
			name:  "block replace basic",
			input: `replace [1 2 3 2 1] 2 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(99), value.NewIntVal(3), value.NewIntVal(2), value.NewIntVal(1)}),
		},
		{
			name:  "block replace pattern not found",
			input: `replace [1 2 3] 4 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "block replace multiple occurrences first only",
			input: `replace [a a a] 'a 'b`,
			want:  value.NewBlockVal([]core.Value{value.NewLitWordVal("b"), value.NewLitWordVal("a"), value.NewLitWordVal("a")}),
		},
		{
			name:  "block replace at start",
			input: `replace [1 2 3] 1 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(99), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "block replace at end",
			input: `replace [1 2 3] 3 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(99)}),
		},
		{
			name:  "block replace empty block",
			input: `replace [] 1 2`,
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "block replace single element match",
			input: `replace [1] 1 2`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(2)}),
		},
		{
			name:  "block replace single element no match",
			input: `replace [1] 2 3`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1)}),
		},

		// Block replacement - all occurrences
		{
			name:  "block replace all basic",
			input: `replace/all [1 2 3 2 1] 2 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(99), value.NewIntVal(3), value.NewIntVal(99), value.NewIntVal(1)}),
		},
		{
			name:  "block replace all not found",
			input: `replace/all [1 2 3] 4 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "block replace all multiple",
			input: `replace/all [1 1 1] 1 2`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(2), value.NewIntVal(2), value.NewIntVal(2)}),
		},
		{
			name:  "block replace all at edges",
			input: `replace/all [2 1 2 1 2] 2 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(99), value.NewIntVal(1), value.NewIntVal(99), value.NewIntVal(1), value.NewIntVal(99)}),
		},

		// Copy mode - strings
		{
			name:  "string replace copy mode",
			input: `s: "test" replace/copy s "t" "x"`,
			want:  value.NewStrVal("xest"),
		},
		{
			name:  "string replace copy mode original unchanged",
			input: `s: "test" replace/copy s "t" "x" s`,
			want:  value.NewStrVal("test"),
		},
		{
			name:  "string replace copy all mode",
			input: `s: "test" replace/copy/all s "t" "x"`,
			want:  value.NewStrVal("xexx"),
		},

		// Copy mode - blocks
		{
			name:  "block replace copy mode",
			input: `b: [1 2 3] replace/copy b 2 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(99), value.NewIntVal(3)}),
		},
		{
			name:  "block replace copy mode original unchanged",
			input: `b: [1 2 3] replace/copy b 2 99 b`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "block replace copy all mode",
			input: `b: [1 2 3 2 1] replace/copy/all b 2 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(99), value.NewIntVal(3), value.NewIntVal(99), value.NewIntVal(1)}),
		},

		// Type mixing in blocks
		{
			name:  "block replace int to string",
			input: `replace [1 2 3] 2 "two"`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewStrVal("two"), value.NewIntVal(3)}),
		},
		{
			name:  "block replace string to int",
			input: `replace ["a" "b"] "a" 99`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(99), value.NewStrVal("b")}),
		},
		{
			name:  "block replace word to word",
			input: `replace [a b c] 'a 'x`,
			want:  value.NewBlockVal([]core.Value{value.NewLitWordVal("x"), value.NewLitWordVal("b"), value.NewLitWordVal("c")}),
		},

		// Edge cases
		{
			name:  "string replace empty pattern error",
			input: `replace "hello" "" "x"`,
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:  "block replace empty block unchanged",
			input: `replace [] 1 2`,
			want:  value.NewBlockVal([]core.Value{}),
		},

		// Error cases
		{
			name:    "non-series input error",
			input:   `replace 42 "x" "y"`,
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "string with non-string pattern error",
			input:   `replace "hello" 1 "x"`,
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "empty pattern error",
			input:   `replace "hello" "" "x"`,
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				if tt.errID != "" {
					var scriptErr *verror.Error
					if errors.As(err, &scriptErr) {
						if scriptErr.ID != tt.errID {
							t.Fatalf("expected error ID %v, got %v", tt.errID, scriptErr.ID)
						}
					} else {
						t.Fatalf("expected ScriptError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !evalResult.Equals(tt.want) {
					t.Fatalf("expected %v, got %v", tt.want, evalResult)
				}
			}
		})
	}
}