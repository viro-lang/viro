package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestSeries_First(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "block first element",
			input: "first [1 2 3]",
			want:  value.NewIntVal(1),
		},
		{
			name:  "single element block",
			input: "first [42]",
			want:  value.NewIntVal(42),
		},
		{
			name:  "string first character",
			input: "first \"hello\"",
			want:  value.NewStrVal("h"),
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
			evalResult, err := Evaluate(tt.input)
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
		want    core.Value
		wantErr bool
	}{
		{
			name:  "block last element",
			input: "last [1 2 3]",
			want:  value.NewIntVal(3),
		},
		{
			name:  "string last character",
			input: "last \"hello\"",
			want:  value.NewStrVal("o"),
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
			evalResult, err := Evaluate(tt.input)
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
		want    core.Value
		wantErr bool
	}{
		{
			name: "append to block",
			input: `data: [1 2 3]
append data 4
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1),
				value.NewIntVal(2),
				value.NewIntVal(3),
				value.NewIntVal(4),
			}),
		},
		{
			name: "append mixed type block",
			input: `data: [1]
append data "x"
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1),
				value.NewStrVal("x"),
			}),
		},
		{
			name: "append to string",
			input: `str: "hi"
append str " there"
str`,
			want: value.NewStrVal("hi there"),
		},
		{
			name:    "non series error",
			input:   "append 42 1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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
		want    core.Value
		wantErr bool
	}{
		{
			name: "insert into block",
			input: `data: [1 2 3]
insert data 0
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(0),
				value.NewIntVal(1),
				value.NewIntVal(2),
				value.NewIntVal(3),
			}),
		},
		{
			name: "insert into string",
			input: `str: "world"
insert str "hello "
str`,
			want: value.NewStrVal("hello world"),
		},
		{
			name:    "non series error",
			input:   "insert true 1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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
		want    core.Value
		wantErr bool
	}{
		{
			name:  "block length",
			input: "length? [1 2 3]",
			want:  value.NewIntVal(3),
		},
		{
			name:  "empty block length",
			input: "length? []",
			want:  value.NewIntVal(0),
		},
		{
			name:  "string length",
			input: "length? \"hello\"",
			want:  value.NewIntVal(5),
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
			want: value.NewIntVal(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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
		want := value.NewBlockVal([]core.Value{
			value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3),
		})
		evalResult, err := Evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("copy string", func(t *testing.T) {
		input := "copy \"hello\""
		want := value.NewStrVal("hello")
		evalResult, err := Evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("copy --part block", func(t *testing.T) {
		input := "copy --part 2 [1 2 3 4]"
		want := value.NewBlockVal([]core.Value{
			value.NewIntVal(1), value.NewIntVal(2),
		})
		evalResult, err := Evaluate(input)
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
		want := value.NewStrVal("abc")
		evalResult, err := Evaluate(input)
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
		evalResult, err := Evaluate(input)
		if err == nil {
			t.Fatalf("expected error but got result %v", evalResult)
		}
	})

	t.Run("copy --part out of range", func(t *testing.T) {
		input := "copy --part [1 2] 5"
		evalResult, err := Evaluate(input)
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
		want    core.Value
		wantErr bool
	}{
		{
			name:  "find in block",
			input: `find [1 2 3 2 1] 2`,
			want:  value.NewIntVal(2),
		},
		{
			name:  "find in string",
			input: `find "hello world" "o"`,
			want:  value.NewIntVal(5),
		},
		{
			name:  "find --last in block",
			input: `find --last [1 2 3 2 1] 2`,
			want:  value.NewIntVal(4),
		},
		{
			name:  "find --last in string",
			input: `find --last "hello world" "o"`,
			want:  value.NewIntVal(8),
		},
		{
			name:  "find not found in block",
			input: `find [1 2 3] 4`,
			want:  value.NewNoneVal(),
		},
		{
			name:  "find not found in string",
			input: `find "hello" "z"`,
			want:  value.NewNoneVal(),
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
			evalResult, err := Evaluate(tt.input)
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

// T102: remove, remove --part for blocks and strings
func TestSeries_Remove(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name: "remove from block",
			input: `data: [1 2 3 4 5]
remove data
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(2),
				value.NewIntVal(3),
				value.NewIntVal(4),
				value.NewIntVal(5),
			}),
		},
		{
			name: "remove from string",
			input: `str: "hello"
remove str
str`,
			want: value.NewStrVal("ello"),
		},
		{
			name: "remove --part from block",
			input: `data: [1 2 3 4 5]
remove data --part 3
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(4),
				value.NewIntVal(5),
			}),
		},
		{
			name: "remove --part from string",
			input: `str: "hello"
remove str --part 2
str`,
			want: value.NewStrVal("llo"),
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
			evalResult, err := Evaluate(tt.input)
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

// T103: skip, take operations
func TestSeries_SkipTake(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name: "skip and take block",
			input: `data: [1 2 3 4 5]
skip data 2
take data 2`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(3),
				value.NewIntVal(4),
			}),
		},
		{
			name: "skip and take string",
			input: `str: "hello"
skip str 1
take str 3`,
			want: value.NewStrVal("ell"),
		},
		{
			name: "take returns a new series",
			input: `data: [1 2 3]
part: take data 2
part`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1),
				value.NewIntVal(2),
			}),
		},
		{
			name:    "skip non-series error",
			input:   "skip 42 1",
			wantErr: true,
		},
		{
			name:    "take non-series error",
			input:   "take 42 1",
			wantErr: true,
		},
		{
			name:    "skip with non-integer error",
			input:   `skip [1 2] "a"`,
			wantErr: true,
		},
		{
			name:    "take with non-integer error",
			input:   `take [1 2] "a"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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

// T105: next operations
func TestSeries_Next(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name: "next block",
			input: `data: [1 2 3]
nextData: next data
first nextData`,
			want: value.NewIntVal(2),
		},
		{
			name: "next string",
			input: `str: "hello"
nextStr: next str
first nextStr`,
			want: value.NewStrVal("e"),
		},
		{
			name: "next preserves original position",
			input: `data: [1 2 3]
nextData: next data
first data`,
			want: value.NewIntVal(1),
		},
		{
			name:    "next non-series error",
			input:   "next 42",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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

// T106: head operations
func TestSeries_Head(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name: "head block",
			input: `data: [1 2 3]
headData: head data
first headData`,
			want: value.NewIntVal(1),
		},
		{
			name: "head string",
			input: `str: "hello"
headStr: head str
first headStr`,
			want: value.NewStrVal("h"),
		},
		{
			name: "head preserves original position",
			input: `data: [1 2 3]
movedData: next next data
headData: head movedData
first headData`,
			want: value.NewIntVal(1),
		},
		{
			name: "head on already at head",
			input: `data: [1 2 3]
headData: head data
first headData`,
			want: value.NewIntVal(1),
		},
		{
			name:    "head non-series error",
			input:   "head 42",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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

// T103: index? on series
func TestSeries_Index(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name: "index? block at head",
			input: `data: [1 2 3]
index? data`,
			want: value.NewIntVal(1),
		},
		{
			name: "index? string at head",
			input: `str: "hello"
index? str`,
			want: value.NewIntVal(1),
		},
		{
			name: "index? block after next",
			input: `data: [1 2 3]
moved: next data
index? moved`,
			want: value.NewIntVal(2),
		},
		{
			name: "index? string after skip",
			input: `str: "hello"
moved: skip str 2
index? moved`,
			want: value.NewIntVal(3),
		},
		{
			name: "index? block at tail",
			input: `data: [1 2 3]
moved: skip data 3
index? moved`,
			want: value.NewIntVal(4),
		},
		{
			name:    "index? non-series error",
			input:   "index? 42",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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

// T104: sort, reverse on series
func TestSeries_SortReverse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name: "sort block of integers",
			input: `data: [3 1 4 1 5 9 2 6]
sort data
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1), value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3),
				value.NewIntVal(4), value.NewIntVal(5), value.NewIntVal(6), value.NewIntVal(9),
			}),
		},
		{
			name: "sort block of strings",
			input: `data: ["c" "a" "b"]
sort data
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewStrVal("a"), value.NewStrVal("b"), value.NewStrVal("c"),
			}),
		},
		{
			name: "sort string",
			input: `str: "cba"
sort str
str`,
			want: value.NewStrVal("abc"),
		},
		{
			name: "reverse block",
			input: `data: [1 2 3]
reverse data
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(3), value.NewIntVal(2), value.NewIntVal(1),
			}),
		},
		{
			name: "reverse string",
			input: `str: "hello"
reverse str
str`,
			want: value.NewStrVal("olleh"),
		},
		{
			name:    "sort non-series error",
			input:   "sort 42",
			wantErr: true,
		},
		{
			name:    "reverse non-series error",
			input:   "reverse 42",
			wantErr: true,
		},
		{
			name:    "sort mixed types error",
			input:   "sort [1 \"a\"]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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

func TestSeries_At(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "block at valid index",
			input: "at [1 2 3 4 5] 3",
			want:  value.NewIntVal(3),
		},
		{
			name:  "block at first index",
			input: "at [1 2 3] 1",
			want:  value.NewIntVal(1),
		},
		{
			name:  "block at last index",
			input: "at [1 2 3] 3",
			want:  value.NewIntVal(3),
		},
		{
			name:  "string at valid index",
			input: `at "hello" 2`,
			want:  value.NewStrVal("e"),
		},
		{
			name:  "string at first index",
			input: `at "world" 1`,
			want:  value.NewStrVal("w"),
		},
		{
			name:    "block index out of bounds negative",
			input:   "at [1 2 3] 0",
			wantErr: true,
		},
		{
			name:    "block index out of bounds too large",
			input:   "at [1 2 3] 4",
			wantErr: true,
		},
		{
			name:    "string index out of bounds",
			input:   `at "hi" 3`,
			wantErr: true,
		},
		{
			name:    "empty block error",
			input:   "at [] 1",
			wantErr: true,
		},
		{
			name:    "empty string error",
			input:   `at "" 1`,
			wantErr: true,
		},
		{
			name:    "wrong series type error",
			input:   "at 42 1",
			wantErr: true,
		},
		{
			name:    "wrong index type error",
			input:   `at [1 2 3] "a"`,
			wantErr: true,
		},
		{
			name:    "too few arguments error",
			input:   "at [1 2 3]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
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
