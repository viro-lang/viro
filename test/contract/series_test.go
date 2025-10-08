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

func evaluate(src string) (value.Value, *verror.Error) {
	vals, err := parse.Parse(src)
	if err != nil {
		return value.NoneVal(), err
	}

	e := eval.NewEvaluator()
	return e.Do_Blk(vals)
}
