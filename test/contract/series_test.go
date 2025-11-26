package contract

import (
	"errors"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestSeries_First(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			name:  "empty block returns none",
			input: "first []",
			want:  value.NewNoneVal(),
		},
		{
			name:  "empty string returns none",
			input: "first \"\"",
			want:  value.NewNoneVal(),
		},
		{
			name: "first at tail position returns none",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			want: value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "first 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_Last(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
		errArgs []string
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
			name:  "empty block returns none",
			input: "last []",
			want:  value.NewNoneVal(),
		},
		{
			name:  "empty string returns none",
			input: "last \"\"",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "last true",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name: "last at tail position",
			input: `data: [1 2 3]
tailData: tail data
last tailData`,
			want: value.NewIntVal(3),
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

func TestSeries_Append(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_Insert(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_LengthQ(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_Copy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "copy block at head",
			input: "copy [1 2 3]",
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3),
			}),
		},
		{
			name:  "copy string at head",
			input: `copy "hello"`,
			want:  value.NewStrVal("hello"),
		},
		{
			name:  "copy binary at head",
			input: "copy #{AABBCC}",
			want:  value.NewBinaryVal([]byte{0xAA, 0xBB, 0xCC}),
		},
		{
			name:    "copy non-series error",
			input:   "copy 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "copy --part block at head",
			input: "copy --part 2 [1 2 3 4]",
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1), value.NewIntVal(2),
			}),
		},
		{
			name:  "copy --part string at head",
			input: `copy --part 3 "abcdef"`,
			want:  value.NewStrVal("abc"),
		},
		{
			name:  "copy --part binary at head",
			input: "copy --part 2 #{AABBCCDD}",
			want:  value.NewBinaryVal([]byte{0xAA, 0xBB}),
		},
		{
			name:  "copy --part zero count",
			input: "copy --part 0 [1 2 3]",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "copy --part count equals remaining at head",
			input: "copy --part 3 [1 2 3]",
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3),
			}),
		},
		{
			name:    "copy --part count exceeds remaining at head",
			input:   "copy --part 5 [1 2]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "copy --part negative count",
			input:   "copy --part -1 [1 2]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "copy --part string out of range",
			input:   `copy --part 10 "hello"`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "copy block from advanced index",
			input: `
				a: next next [1 2 3 4]
				copy a
			`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(3), value.NewIntVal(4),
			}),
		},
		{
			name: "copy string from advanced index",
			input: `
				a: next next "hello"
				copy a
			`,
			want: value.NewStrVal("llo"),
		},
		{
			name: "copy binary from advanced index",
			input: `
				a: next next #{AABBCCDD}
				copy a
			`,
			want: value.NewBinaryVal([]byte{0xCC, 0xDD}),
		},
		{
			name: "copy --part from advanced index with count in range",
			input: `
				b: next next [1 2 3 4 5]
				copy --part 3 b
			`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(3), value.NewIntVal(4), value.NewIntVal(5),
			}),
		},
		{
			name: "copy --part string from advanced index",
			input: `
				s: next next "hello"
				copy --part 3 s
			`,
			want: value.NewStrVal("llo"),
		},
		{
			name: "copy --part count equals remaining from advanced index",
			input: `
				b: next [1 2 3]
				copy --part 2 b
			`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(2), value.NewIntVal(3),
			}),
		},
		{
			name: "copy --part count exceeds remaining from advanced index",
			input: `
				b: next next [1 2 3]
				copy --part 5 b
			`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "copy from tail returns empty block",
			input: `
				a: tail [1 2 3]
				copy a
			`,
			want: value.NewBlockVal([]core.Value{}),
		},
		{
			name: "copy from tail returns empty string",
			input: `
				a: tail "hello"
				copy a
			`,
			want: value.NewStrVal(""),
		},
		{
			name: "copy from tail returns empty binary",
			input: `
				a: tail #{AABBCC}
				copy a
			`,
			want: value.NewBinaryVal([]byte{}),
		},
		{
			name: "copy --part from tail yields empty",
			input: `
				b: tail [1 2 3]
				copy --part 2 b
			`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "copy --part zero from advanced index",
			input: `
				b: next [1 2 3]
				copy --part 0 b
			`,
			want: value.NewBlockVal([]core.Value{}),
		},
		{
			name: "copy --part zero from advanced index string",
			input: `
				s: next next "hello"
				copy --part 0 s
			`,
			want: value.NewStrVal(""),
		},
		{
			name: "copy --part zero from advanced index binary",
			input: `
				b: next #{AABBCCDD}
				copy --part 0 b
			`,
			want: value.NewBinaryVal([]byte{}),
		},
		{
			name: "copy --part zero from tail block",
			input: `
				b: tail [1 2 3]
				copy --part 0 b
			`,
			want: value.NewBlockVal([]core.Value{}),
		},
		{
			name: "copy --part zero from tail string",
			input: `
				s: tail "hello"
				copy --part 0 s
			`,
			want: value.NewStrVal(""),
		},
		{
			name: "copy --part zero from tail binary",
			input: `
				b: tail #{AABBCC}
				copy --part 0 b
			`,
			want: value.NewBinaryVal([]byte{}),
		},
		{
			name:    "copy --part negative count binary",
			input:   "copy --part -1 #{AABBCC}",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "copy --part binary from advanced index",
			input: `
				b: next #{AABBCCDDEE}
				copy --part 2 b
			`,
			want: value.NewBinaryVal([]byte{0xBB, 0xCC}),
		},
		{
			name: "copy --part count equals remaining from advanced index binary",
			input: `
				b: next next #{AABBCCDD}
				copy --part 2 b
			`,
			want: value.NewBinaryVal([]byte{0xCC, 0xDD}),
		},
		{
			name: "copy --part count exceeds remaining from advanced index binary",
			input: `
				b: next next #{AABBCC}
				copy --part 5 b
			`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "copy string with UTF-8 multibyte from advanced position",
			input: `
				s: next next "żółć"
				copy s
			`,
			want: value.NewStrVal("łć"),
		},
		{
			name: "copy result resets to head block",
			input: `
				a: next next [1 2 3 4]
				b: copy a
				head? b
			`,
			want: value.NewLogicVal(true),
		},
		{
			name: "copy result resets to head string",
			input: `
				a: next next "hello"
				b: copy a
				head? b
			`,
			want: value.NewLogicVal(true),
		},
		{
			name: "copy result resets to head binary",
			input: `
				a: next next #{AABBCCDD}
				b: copy a
				head? b
			`,
			want: value.NewLogicVal(true),
		},
		{
			name:  "copy empty block",
			input: "copy []",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "copy --part 0 from empty series",
			input: "copy --part 0 []",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name: "copy full series after head advancement (migration pattern)",
			input: `
				a: next next [1 2 3 4]
				copy head a
			`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3), value.NewIntVal(4),
			}),
		},
		{
			name: "copy does not modify source series index",
			input: `
				a: next next [1 2 3 4]
				b: copy a
				index? a
			`,
			want: value.NewIntVal(3),
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

func TestSeries_Pick(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "pick block valid index",
			input: "pick [1 2 3] 2",
			want:  value.NewIntVal(2),
		},
		{
			name:  "pick block first element",
			input: "pick [1 2 3] 1",
			want:  value.NewIntVal(1),
		},
		{
			name:  "pick block last element",
			input: "pick [1 2 3] 3",
			want:  value.NewIntVal(3),
		},
		{
			name:  "pick block out of bounds returns none",
			input: "pick [1 2 3] 10",
			want:  value.NewNoneVal(),
		},
		{
			name:  "pick block index zero returns none",
			input: "pick [1 2 3] 0",
			want:  value.NewNoneVal(),
		},
		{
			name:  "pick block negative index returns none",
			input: "pick [1 2 3] -1",
			want:  value.NewNoneVal(),
		},
		{
			name:  "pick string valid index",
			input: `pick "hello" 1`,
			want:  value.NewStrVal("h"),
		},
		{
			name:  "pick string last char",
			input: `pick "hello" 5`,
			want:  value.NewStrVal("o"),
		},
		{
			name:  "pick string out of bounds returns none",
			input: `pick "hello" 10`,
			want:  value.NewNoneVal(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestSeries_Poke(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "poke block valid index",
			input: "poke [1 2 3] 2 99",
			want:  value.NewIntVal(99),
		},
		{
			name:    "poke block out of bounds",
			input:   "poke [1 2 3] 10 99",
			wantErr: true,
		},
		{
			name:    "poke block index zero errors",
			input:   "poke [1 2 3] 0 99",
			wantErr: true,
		},
		{
			name:    "poke block negative index errors",
			input:   "poke [1 2 3] -1 99",
			wantErr: true,
		},
		{
			name:  "poke string valid single char",
			input: `poke "hello" 1 "H"`,
			want:  value.NewStrVal("H"),
		},
		{
			name:    "poke string with non-string",
			input:   `poke "hello" 1 123`,
			wantErr: true,
		},
		{
			name:    "poke string with multi-char",
			input:   `poke "hello" 1 "ab"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestSeries_Select(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "select block found",
			input: "select [name \"Alice\" age 30] 'age",
			want:  value.NewIntVal(30),
		},
		{
			name:  "select block not found returns none",
			input: "select [1 2 3 4] 5",
			want:  value.NewNoneVal(),
		},
		{
			name:  "select block with default when found returns value",
			input: "select [a 1 b 2] 'b --default 99",
			want:  value.NewIntVal(2),
		},
		{
			name:  "select block with default when not found returns default",
			input: "select [a 1 b 2] 'c --default 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "select block word-like match lit-word vs word",
			input: "select ['name \"Alice\" 'age 30] 'age",
			want:  value.NewIntVal(30),
		},
		{
			name:  "select string found",
			input: `select "hello world" " "`,
			want:  value.NewStrVal("world"),
		},
		{
			name:  "select string not found returns none",
			input: `select "hello" "z"`,
			want:  value.NewNoneVal(),
		},
		{
			name:  "select string with default when not found",
			input: `select "hello" "z" --default "fallback"`,
			want:  value.NewStrVal("fallback"),
		},
		{
			name:  "select string with default when found returns value",
			input: `select "hello world" "o" --default "fallback"`,
			want:  value.NewStrVal(" world"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestSeries_Clear(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "clear block",
			input: "clear [1 2 3]",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "clear string",
			input: `clear "hello"`,
			want:  value.NewStrVal(""),
		},

		{
			name:  "clear empty block",
			input: "clear []",
			want:  value.NewBlockVal([]core.Value{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestSeries_Change(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "change block",
			input: "change next [1 2 3] 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "change string",
			input: `change next "hello" "a"`,
			want:  value.NewStrVal("a"),
		},
		{
			name:    "change at tail errors",
			input:   `change tail "hello" "x"`,
			wantErr: true,
		},
		{
			name:    "change block at tail errors",
			input:   "change tail [1 2 3] 99",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestSeries_Trim(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:    "trim with no arguments",
			input:   "trim",
			wantErr: true,
			errID:   verror.ErrIDArgCount,
		},
		{
			name:  "trim string with whitespace",
			input: `trim "  hello  "`,
			want:  value.NewStrVal("hello"),
		},
		{
			name:  "trim empty string",
			input: `trim ""`,
			want:  value.NewStrVal(""),
		},
		{
			name:  "trim string no whitespace",
			input: `trim "hello"`,
			want:  value.NewStrVal("hello"),
		},
		{
			name:  "trim string with internal whitespace",
			input: `trim "  hello world  "`,
			want:  value.NewStrVal("hello world"),
		},
		{
			name:  "trim --head removes leading whitespace",
			input: `trim --head "  hello  "`,
			want:  value.NewStrVal("hello  "),
		},
		{
			name:  "trim --head with no leading whitespace",
			input: `trim --head "hello  "`,
			want:  value.NewStrVal("hello  "),
		},
		{
			name:  "trim --tail removes trailing whitespace",
			input: `trim --tail "  hello  "`,
			want:  value.NewStrVal("  hello"),
		},
		{
			name:  "trim --tail with no trailing whitespace",
			input: `trim --tail "  hello"`,
			want:  value.NewStrVal("  hello"),
		},
		{
			name: "trim --auto with indented text",
			input: `trim --auto "    line1
    line2
        line3"`,
			want: value.NewStrVal("line1\nline2\n    line3"),
		},
		{
			name:  "trim --auto with no common indentation",
			input: `trim --auto "  hello  "`,
			want:  value.NewStrVal("hello"),
		},
		{
			name: "trim --lines removes line breaks and extra spaces",
			input: `trim --lines "hello
world"`,
			want: value.NewStrVal("hello world"),
		},
		{
			name:  "trim --lines collapses multiple spaces",
			input: `trim --lines "hello   world"`,
			want:  value.NewStrVal("hello world"),
		},
		{
			name:  "trim --all removes all whitespace",
			input: `trim --all "  hello world  "`,
			want:  value.NewStrVal("helloworld"),
		},
		{
			name:  "trim --all with tabs and spaces",
			input: `trim --all "  hello	 world  "`,
			want:  value.NewStrVal("helloworld"),
		},
		{
			name:  "trim --with removes specified characters",
			input: `trim --with "-" "a-b-c"`,
			want:  value.NewStrVal("abc"),
		},
		{
			name:  "trim --with removes multiple characters",
			input: `trim --with "123" "abc123def"`,
			want:  value.NewStrVal("abcdef"),
		},
		{
			name:    "trim with mutually exclusive refinements",
			input:   `trim --head --tail "  hello  "`,
			wantErr: true,
		},
		{
			name:    "trim --with with non-string argument",
			input:   `trim --with 123 "hello"`,
			wantErr: true,
		},
		{
			name:    "trim with non-string input",
			input:   `trim 123`,
			wantErr: true,
		},
		{
			name:  "trim --with empty pattern does not change string",
			input: `trim --with "" "abc"`,
			want:  value.NewStrVal("abc"),
		},
		{
			name:    "trim --with and --all are mutually exclusive",
			input:   `trim --with "-" --all "a-b"`,
			wantErr: true,
		},
		{
			name:  "trim --lines with CRLF and multiple blank lines",
			input: "trim --lines \"a\r\n\n b\r\n c\"",
			want:  value.NewStrVal("a b c"),
		},
		{
			name:  "trim block default removes leading and trailing none",
			input: `trim [none none 1 2 3]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "trim block default removes trailing none",
			input: `trim [1 2 3 none none]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "trim block default preserves internal none",
			input: `trim [none 1 none 2 none]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewWordVal("none"), value.NewIntVal(2)}),
		},
		{
			name:  "trim block --head removes leading none",
			input: `trim --head [none none 1 2 3]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "trim block --tail removes trailing none",
			input: `trim --tail [1 2 3 none none]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:  "trim block --all removes all none",
			input: `trim --all [none 1 none 2 none]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}),
		},
		{
			name:  "trim block --with removes specific value",
			input: `trim --with 5 [5 1 5 2 5]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}),
		},
		{
			name:  "trim block --with removes specific string",
			input: `trim --with "x" ["x" "a" "x" "b" "x"]`,
			want:  value.NewBlockVal([]core.Value{value.NewStrVal("a"), value.NewStrVal("b")}),
		},
		{
			name:    "trim block with mutually exclusive refinements",
			input:   `trim --head --tail [none 1 none]`,
			wantErr: true,
		},
		{
			name:  "trim block --with with non-matching type",
			input: `trim --with "x" [1 2 3]`,
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
		},
		{
			name:    "trim --auto on block returns error",
			input:   "trim --auto [none 1 none]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "trim --lines on block returns error",
			input:   "trim --lines [1 none 2]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:  "trim all none values",
			input: "trim [none none none]",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "trim --all all none values",
			input: "trim --all [none none none]",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "trim --with none removes literal none",
			input: "trim --with none [none 1 none]",
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(1)}),
		},
		{
			name:  "trim --with 5 complete removal",
			input: "trim --with 5 [5]",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "trim mutates string in place",
			input: `s: "  hi  " trim s s`,
			want:  value.NewStrVal("hi"),
		},
		{
			name:  "trim --all mutates string in place",
			input: `s: "  h i  " trim --all s s`,
			want:  value.NewStrVal("hi"),
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

func TestSeries_Find(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "find string with non-string error",
			input:   `find "hello" 1`,
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
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

func TestSeries_Remove(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "remove --part with non-integer error",
			input:   `remove [1 2] --part "a"`,
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "remove --part out of range error",
			input:   `remove [1 2] --part 3`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "remove --part negative count error",
			input:   `remove [1 2] --part -1`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "remove --part zero count (no-op)",
			input: `data: [1 2 3]
remove data --part 0
data`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1),
				value.NewIntVal(2),
				value.NewIntVal(3),
			}),
		},
		{
			name: "remove --part from string negative count",
			input: `str: "hello"
remove str --part -1`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
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

func TestSeries_SkipTake(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
		errArgs []string
	}{
		{
			name: "skip and take block",
			input: `data: [1 2 3 4 5]
skipped: skip data 2
take skipped 2`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(3),
				value.NewIntVal(4),
			}),
		},
		{
			name: "skip and take string",
			input: `str: "hello"
skipped: skip str 1
take skipped 3`,
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
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "take non-series error",
			input:   "take 42 1",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "skip with non-integer error",
			input:   `skip [1 2] "a"`,
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "take with non-integer error",
			input:   `take [1 2] "a"`,
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "take negative count error block",
			input:   "take [1 2 3] -1",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
			errArgs: []string{"-1", "3", "0"},
		},
		{
			name:    "take negative count error string",
			input:   `take "hello" -1`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
			errArgs: []string{"-1", "5", "0"},
		},
		{
			name:  "take zero count block",
			input: "take [1 2 3] 0",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "take zero count string",
			input: `take "hello" 0`,
			want:  value.NewStrVal(""),
		},
		{
			name:  "take oversized count clamps block",
			input: "take [1 2 3] 10",
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(1),
				value.NewIntVal(2),
				value.NewIntVal(3),
			}),
		},
		{
			name:  "take oversized count clamps string",
			input: `take "hello" 10`,
			want:  value.NewStrVal("hello"),
		},
		{
			name: "take from advanced index clamps",
			input: `data: [1 2 3 4 5]
data: next next data
take data 10`,
			want: value.NewBlockVal([]core.Value{
				value.NewIntVal(3),
				value.NewIntVal(4),
				value.NewIntVal(5),
			}),
		},
		{
			name: "take from tail position returns empty",
			input: `data: [1 2 3]
data: skip data 3
take data 2`,
			want: value.NewBlockVal([]core.Value{}),
		},
		{
			name: "take string from tail position returns empty",
			input: `str: "abc"
str: skip str 3
take str 2`,
			want: value.NewStrVal(""),
		},
		{
			name: "skip does not mutate original series",
			input: `data: [1 2 3 4 5]
skipped: skip data 2
index? data`,
			want: value.NewIntVal(1),
		},
		{
			name: "skip returns new view at correct position",
			input: `data: [1 2 3 4 5]
skipped: skip data 2
index? skipped`,
			want: value.NewIntVal(3),
		},
		{
			name: "skip preserves original and allows independent navigation",
			input: `data: [1 2 3 4 5]
skipped: skip data 2
first data`,
			want: value.NewIntVal(1),
		},
		{
			name: "skip string does not mutate original",
			input: `str: "hello"
skipped: skip str 2
index? str`,
			want: value.NewIntVal(1),
		},
		{
			name: "skip string returns new view at correct position",
			input: `str: "hello"
skipped: skip str 2
first skipped`,
			want: value.NewStrVal("l"),
		},
		{
			name: "skip with negative count clamps to zero",
			input: `data: next next [1 2 3 4 5]
skipped: skip data -5
index? skipped`,
			want: value.NewIntVal(1),
		},
		{
			name: "skip with oversized count clamps to tail",
			input: `data: [1 2 3]
skipped: skip data 10
tail? skipped`,
			want: value.NewLogicVal(true),
		},
		{
			name: "skip from advanced position does not mutate original",
			input: `data: [1 2 3 4 5]
advanced: next next data
skipped: skip advanced 2
index? advanced`,
			want: value.NewIntVal(3),
		},
		{
			name: "skip binary does not mutate original",
			input: `bin: #{AABBCCDD}
skipped: skip bin 2
index? bin`,
			want: value.NewIntVal(1),
		},
		{
			name: "skip zero returns new cloned view",
			input: `data: [1 2 3]
skipped: skip data 0
index? skipped`,
			want: value.NewIntVal(1),
		},
		{
			name: "skip zero does not mutate original",
			input: `data: [1 2 3]
skipped: skip data 0
first data`,
			want: value.NewIntVal(1),
		},
		{
			name: "independent navigation after skip",
			input: `data: [1 2 3 4 5]
skipped: skip data 2
advanced: next skipped
index? data`,
			want: value.NewIntVal(1),
		},
		{
			name: "advancing returned skip view does not affect original",
			input: `data: [1 2 3 4 5]
skipped: skip data 2
advanced: next next skipped
first data`,
			want: value.NewIntVal(1),
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
						if tt.errArgs != nil {
							if len(scriptErr.Args) != len(tt.errArgs) {
								t.Fatalf("expected error args length %d, got %d", len(tt.errArgs), len(scriptErr.Args))
							}
							for i, expected := range tt.errArgs {
								if scriptErr.Args[i] != expected {
									t.Fatalf("expected error arg[%d] %v, got %v", i, expected, scriptErr.Args[i])
								}
							}
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

func TestSeries_Next(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_Back(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
		errArgs []string
	}{
		{
			name: "back block",
			input: `data: [1 2 3]
backData: back next data
first backData`,
			want: value.NewIntVal(1),
		},
		{
			name: "back string",
			input: `str: "hello"
backStr: back next str
first backStr`,
			want: value.NewStrVal("h"),
		},
		{
			name: "back preserves original position",
			input: `data: [1 2 3]
movedData: next data
backData: back movedData
first data`,
			want: value.NewIntVal(1),
		},
		{
			name:    "back on series at head error",
			input:   `back [1 2 3]`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
			errArgs: []string{"-1", "3", "0"},
		},
		{
			name:    "back on empty block at head error",
			input:   `back []`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "back on empty string at head error",
			input:   `back ""`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "back after multiple next operations",
			input: `data: [1 2 3 4]
moved: next next next data
backData: back moved
first backData`,
			want: value.NewIntVal(3),
		},
		{
			name:    "back non-series error",
			input:   "back 42",
			wantErr: true,
		},
		{
			name: "back at tail position",
			input: `data: [1 2 3]
tailData: skip data 3
backData: back tailData
first backData`,
			want: value.NewIntVal(3),
		},
		{
			name: "back after skip operations",
			input: `data: [1 2 3 4 5]
skipped: skip data 3
backData: back skipped
first backData`,
			want: value.NewIntVal(3),
		},
		{
			name: "back binary skip",
			input: `bin: append #{} 1
bin: append bin 2
bin: append bin 3
moved: next next bin
backData: back moved
first backData`,
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
				if tt.errID != "" {
					var scriptErr *verror.Error
					if errors.As(err, &scriptErr) {
						if scriptErr.ID != tt.errID {
							t.Fatalf("expected error ID %v, got %v", tt.errID, scriptErr.ID)
						}
						if tt.errArgs != nil {
							if len(scriptErr.Args) != len(tt.errArgs) {
								t.Fatalf("expected error args length %d, got %d", len(tt.errArgs), len(scriptErr.Args))
							}
							for i, expected := range tt.errArgs {
								if scriptErr.Args[i] != expected {
									t.Fatalf("expected error arg[%d] %v, got %v", i, expected, scriptErr.Args[i])
								}
							}
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

func TestSeries_Head(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_Tail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
		errArgs []string
	}{
		{
			name: "tail block then first returns none",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			want: value.NewNoneVal(),
		},
		{
			name: "tail string then first returns none",
			input: `str: "hello"
tailStr: tail str
first tailStr`,
			want: value.NewNoneVal(),
		},
		{
			name: "first at tail position returns none",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			want: value.NewNoneVal(),
		},
		{
			name: "tail preserves original position",
			input: `data: [1 2 3]
movedData: next next data
tailData: tail movedData
first data`,
			want: value.NewIntVal(1),
		},
		{
			name: "tail on already at tail then first returns none",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			want: value.NewNoneVal(),
		},
		{
			name:    "tail? non-series error",
			input:   "tail? 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_Index(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
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

func TestSeries_SortReverse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "reverse non-series error",
			input:   "reverse 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "sort mixed types error",
			input:   "sort [1 \"a\"]",
			wantErr: true,
			errID:   verror.ErrIDNotComparable,
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

func TestSeries_At(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
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
			name:  "block index zero returns none",
			input: "at [1 2 3] 0",
			want:  value.NewNoneVal(),
		},
		{
			name:  "block index out of bounds too large returns none",
			input: "at [1 2 3] 4",
			want:  value.NewNoneVal(),
		},
		{
			name:  "string index out of bounds returns none",
			input: `at "hi" 3`,
			want:  value.NewNoneVal(),
		},
		{
			name:  "empty block returns none",
			input: "at [] 1",
			want:  value.NewNoneVal(),
		},
		{
			name:  "empty string returns none",
			input: `at "" 1`,
			want:  value.NewNoneVal(),
		},
		{
			name:    "wrong series type error",
			input:   "at 42 1",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:    "wrong index type error",
			input:   `at [1 2 3] "a"`,
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name: "at next with index 1 returns current element",
			input: `data: [1 2]
			at next data 1`,
			want: value.NewIntVal(2),
		},
		{
			name: "at next with index 2 moves to next element",
			input: `data: [1 2 3]
			at next data 2`,
			want: value.NewIntVal(3),
		},
		{
			name: "at next with index 2 returns none when beyond bounds",
			input: `data: [1 2]
			at next data 2`,
			want: value.NewNoneVal(),
		},
		{
			name: "at skip from position 1 with index 2 returns element 3",
			input: `data: [1 2 3]
			at skip data 1 2`,
			want: value.NewIntVal(3),
		},
		{
			name: "at skip from position 1 with index 2 returns none when beyond position",
			input: `data: [1 2]
			at skip data 1 2`,
			want: value.NewNoneVal(),
		},
		{
			name: "at next next with index 1 returns element at position 2",
			input: `data: [1 2 3 4]
			at next next data 1`,
			want: value.NewIntVal(3),
		},
		{
			name: "at skip from position 1 with index 1 returns element 2",
			input: `data: [1 2 3]
			at skip data 1 1`,
			want: value.NewIntVal(2),
		},
		{
			name: "at skip from position 1 with index 2 returns element 3",
			input: `data: [1 2 3]
			at skip data 1 2`,
			want: value.NewIntVal(3),
		},
		{
			name: "at skip from position 1 with index 2 returns none when beyond position",
			input: `data: [1 2]
			at skip data 1 2`,
			want: value.NewNoneVal(),
		},
		{
			name:    "too few arguments error",
			input:   "at [1 2 3]",
			wantErr: true,
			errID:   verror.ErrIDArgCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalResult, err := Evaluate(tt.input)
			if tt.wantErr {
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
				if err == nil {
					t.Fatalf("expected error but got nil result %v", evalResult)
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

func TestSeries_QueryFunctions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "empty? empty block",
			input: "empty? []",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? non-empty block",
			input: "empty? [1 2 3]",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "empty? empty string",
			input: `empty? ""`,
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? non-empty string",
			input: `empty? "hello"`,
			want:  value.NewLogicVal(false),
		},
		{
			name:  "empty? single char string",
			input: `empty? "a"`,
			want:  value.NewLogicVal(false),
		},
		{
			name:    "empty? non-series error",
			input:   "empty? 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "empty? block at tail",
			input: "empty? tail [1 2 3]",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? block after skip to end",
			input: "empty? skip [1 2 3] 3",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? block after back from tail",
			input: "empty? back tail [1 2 3]",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "empty? string at tail",
			input: `empty? tail "hello"`,
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? string after next single char",
			input: `empty? next "a"`,
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? string after skip to end",
			input: `empty? skip "hello" 5`,
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? binary at head",
			input: "empty? #{DEADBEEF}",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "empty? binary at tail",
			input: "empty? tail #{DEADBEEF}",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? binary after next",
			input: "empty? next #{DE}",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? binary after back from tail",
			input: "empty? back tail #{DEADBEEF}",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "empty? binary after skip to end",
			input: "empty? skip #{DEADBEEF} 4",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "empty? block mid-series false",
			input: "empty? next [1 2 3]",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "empty? empty block at tail",
			input: "empty? tail []",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "head? block at head",
			input: "head? [1 2 3]",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "head? block not at head",
			input: "head? next [1 2 3]",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "head? string at head",
			input: `head? "hello"`,
			want:  value.NewLogicVal(true),
		},
		{
			name:  "head? string not at head",
			input: `head? next "hello"`,
			want:  value.NewLogicVal(false),
		},
		{
			name:  "head? block after skip",
			input: "head? skip [1 2 3] 2",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "head? block after back to head",
			input: "head? head next [1 2 3]",
			want:  value.NewLogicVal(true),
		},
		{
			name:    "head? non-series error",
			input:   "head? 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "tail? block not at tail",
			input: "tail? [1 2 3]",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "tail? block at tail",
			input: "tail? tail [1 2 3]",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "tail? string not at tail",
			input: `tail? "hello"`,
			want:  value.NewLogicVal(false),
		},
		{
			name:  "tail? string at tail",
			input: `tail? tail "hello"`,
			want:  value.NewLogicVal(true),
		},
		{
			name:  "tail? empty block",
			input: "tail? []",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "tail? empty string",
			input: `tail? ""`,
			want:  value.NewLogicVal(true),
		},
		{
			name:  "tail? block after skip to end",
			input: "tail? skip [1 2 3] 3",
			want:  value.NewLogicVal(true),
		},
		{
			name:    "tail? non-series error",
			input:   "tail? 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
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
