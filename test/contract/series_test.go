package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestSeries_First(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    value.Value
		wantErr bool
	}{
		{
			name:  "block first element",
			input: "first [1 2 3]",
			want:  value.IntVal(1),
		},
		{
			name:  "single element block",
			input: "first [42]",
			want:  value.IntVal(42),
		},
		{
			name:  "string first character",
			input: "first \"hello\"",
			want:  value.StrVal("h"),
		},
		{
			name:    "empty block error",
			input:   "first []",
			wantErr: true,
		},
		{
			name:    "empty string error",
			input:   "first \"\"",
			wantErr: true,
		},
		{
			name:    "non series error",
			input:   "first 42",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !evalResult.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, evalResult)
			}
		})
	}
}

func TestSeries_Last(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    value.Value
		wantErr bool
	}{
		{
			name:  "block last element",
			input: "last [1 2 3]",
			want:  value.IntVal(3),
		},
		{
			name:  "string last character",
			input: "last \"hello\"",
			want:  value.StrVal("o"),
		},
		{
			name:    "empty block error",
			input:   "last []",
			wantErr: true,
		},
		{
			name:    "empty string error",
			input:   "last \"\"",
			wantErr: true,
		},
		{
			name:    "non series error",
			input:   "last true",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !evalResult.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, evalResult)
			}
		})
	}
}

func TestSeries_Append(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    value.Value
		wantErr bool
	}{
		{
			name: "append to block",
			input: `data: [1 2 3]
append data 4
data`,
			want: value.BlockVal([]value.Value{
				value.IntVal(1),
				value.IntVal(2),
				value.IntVal(3),
				value.IntVal(4),
			}),
		},
		{
			name: "append mixed type block",
			input: `data: [1]
append data "x"
data`,
			want: value.BlockVal([]value.Value{
				value.IntVal(1),
				value.StrVal("x"),
			}),
		},
		{
			name: "append to string",
			input: `str: "hi"
append str " there"
str`,
			want: value.StrVal("hi there"),
		},
		{
			name:    "non series error",
			input:   "append 42 1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !evalResult.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, evalResult)
			}
		})
	}
}

func TestSeries_Insert(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    value.Value
		wantErr bool
	}{
		{
			name: "insert into block",
			input: `data: [1 2 3]
insert data 0
data`,
			want: value.BlockVal([]value.Value{
				value.IntVal(0),
				value.IntVal(1),
				value.IntVal(2),
				value.IntVal(3),
			}),
		},
		{
			name: "insert into string",
			input: `str: "world"
insert str "hello "
str`,
			want: value.StrVal("hello world"),
		},
		{
			name:    "non series error",
			input:   "insert true 1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !evalResult.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, evalResult)
			}
		})
	}
}

func TestSeries_LengthQ(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    value.Value
		wantErr bool
	}{
		{
			name:  "block length",
			input: "length? [1 2 3]",
			want:  value.IntVal(3),
		},
		{
			name:  "empty block length",
			input: "length? []",
			want:  value.IntVal(0),
		},
		{
			name:  "string length",
			input: "length? \"hello\"",
			want:  value.IntVal(5),
		},
		{
			name:    "non series error",
			input:   "length? 42",
			wantErr: true,
		},
		{
			name: "length after append",
			input: `data: [1]
append data 2
length? data`,
			want: value.IntVal(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !evalResult.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, evalResult)
			}
		})
	}
}

// T100: copy, copy --part for blocks and strings
func TestSeries_Copy(t *testing.T) {
	t.Run("copy block", func(t *testing.T) {
		input := "copy [1 2 3]"
		want := value.BlockVal([]value.Value{
			value.IntVal(1), value.IntVal(2), value.IntVal(3),
		})
		evalResult, err := evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("copy string", func(t *testing.T) {
		input := "copy \"hello\""
		want := value.StrVal("hello")
		evalResult, err := evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("copy --part block", func(t *testing.T) {
		input := "copy --part 2 [1 2 3 4]"
		want := value.BlockVal([]value.Value{
			value.IntVal(1), value.IntVal(2),
		})
		evalResult, err := evaluate(input)
		if err == nil {
			if !evalResult.Equals(want) {
				t.Fatalf("expected %v, got %v", want, evalResult)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("copy --part string", func(t *testing.T) {
		input := "copy --part 3 \"abcdef\""
		want := value.StrVal("abc")
		evalResult, err := evaluate(input)
		if err == nil {
			if !evalResult.Equals(want) {
				t.Fatalf("expected %v, got %v", want, evalResult)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("copy non-series error", func(t *testing.T) {
		input := "copy 42"
		evalResult, err := evaluate(input)
		if err == nil {
			t.Fatalf("expected error but got result %v", evalResult)
		}
	})

	t.Run("copy --part out of range", func(t *testing.T) {
		input := "copy --part [1 2] 5"
		evalResult, err := evaluate(input)
		if err == nil {
			t.Fatalf("expected error but got result %v", evalResult)
		}
	})
}

// T101: find, find --last for blocks and strings
func TestSeries_Find(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    value.Value
		wantErr bool
	}{
		{
			name:  "find in block",
			input: `find [1 2 3 2 1] 2`,
			want:  value.IntVal(2),
		},
		{
			name:  "find in string",
			input: `find "hello world" "o"`,
			want:  value.IntVal(5),
		},
		{
			name:  "find --last in block",
			input: `find --last [1 2 3 2 1] 2`,
			want:  value.IntVal(4),
		},
		{
			name:  "find --last in string",
			input: `find --last "hello world" "o"`,
			want:  value.IntVal(8),
		},
		{
			name:  "find not found in block",
			input: `find [1 2 3] 4`,
			want:  value.NoneVal(),
		},
		{
			name:  "find not found in string",
			input: `find "hello" "z"`,
			want:  value.NoneVal(),
		},
		{
			name:    "find non-series error",
			input:   "find 42 1",
			wantErr: true,
		},
		{
			name:    "find string with non-string error",
			input:   `find "hello" 1`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !evalResult.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, evalResult)
			}
		})
	}
}

func evaluate(src string) (value.Value, *verror.Error) {
	vals, err := parse.Parse(src)
	if err != nil {
		return value.NoneVal(), err
	}

	e := eval.NewEvaluator()
	return e.Do_Blk(vals)
}

// T102: remove, remove --part for blocks and strings
func TestSeries_Remove(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    value.Value
		wantErr bool
	}{
		{
			name: "remove from block",
			input: `data: [1 2 3 4 5]
remove data
data`,
			want: value.BlockVal([]value.Value{
				value.IntVal(2),
				value.IntVal(3),
				value.IntVal(4),
				value.IntVal(5),
			}),
		},
		{
			name: "remove from string",
			input: `str: "hello"
remove str
str`,
			want: value.StrVal("ello"),
		},
		{
			name: "remove --part from block",
			input: `data: [1 2 3 4 5]
remove data --part 3
data`,
			want: value.BlockVal([]value.Value{
				value.IntVal(4),
				value.IntVal(5),
			}),
		},
		{
			name: "remove --part from string",
			input: `str: "hello"
remove str --part 2
str`,
			want: value.StrVal("llo"),
		},
		{
			name:    "remove from non-series error",
			input:   "remove 42",
			wantErr: true,
		},
		{
			name:    "remove --part with non-integer error",
			input:   `remove [1 2] --part "a"`,
			wantErr: true,
		},
		{
			name:    "remove --part out of range error",
			input:   `remove [1 2] --part 3`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !evalResult.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, evalResult)
			}
		})
	}
}
