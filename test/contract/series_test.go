package contract

import (
	"errors"
	"strings"
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
			name:    "empty block error",
			input:   "first []",
			wantErr: true,
			errID:   verror.ErrIDEmptySeries,
		},
		{
			name:    "empty string error",
			input:   "first \"\"",
			wantErr: true,
			errID:   verror.ErrIDEmptySeries,
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
			name:    "empty block error",
			input:   "last []",
			wantErr: true,
			errID:   verror.ErrIDEmptySeries,
		},
		{
			name:    "empty string error",
			input:   "last \"\"",
			wantErr: true,
			errID:   verror.ErrIDEmptySeries,
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
		input := "copy --part 5 [1 2]"
		evalResult, err := Evaluate(input)
		if err == nil {
			t.Fatalf("expected error but got result %v", evalResult)
		}
		var scriptErr *verror.Error
		if !errors.As(err, &scriptErr) {
			t.Fatalf("expected script error, got %v", err)
		}
	})

	t.Run("copy --part negative count", func(t *testing.T) {
		input := "copy --part -1 [1 2]"
		evalResult, err := Evaluate(input)
		if err == nil {
			t.Fatalf("expected error but got result %v", evalResult)
		}
		var scriptErr *verror.Error
		if errors.As(err, &scriptErr) {
			if scriptErr.ID != verror.ErrIDOutOfBounds {
				t.Fatalf("expected error ID %v, got %v", verror.ErrIDOutOfBounds, scriptErr.ID)
			}
			if len(scriptErr.Args) < 3 || scriptErr.Args[0] != "-1" || scriptErr.Args[1] != "2" || scriptErr.Args[2] != "0" {
				t.Fatalf("expected error args ['-1', '2', '0'], got %v", scriptErr.Args)
			}
		} else {
			t.Fatalf("expected ScriptError, got %T", err)
		}
	})

	t.Run("copy --part zero count", func(t *testing.T) {
		input := "copy --part 0 [1 2 3]"
		want := value.NewBlockVal([]core.Value{})
		evalResult, err := Evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("copy --part string out of range", func(t *testing.T) {
		input := "copy --part 10 \"hello\""
		evalResult, err := Evaluate(input)
		if err == nil {
			t.Fatalf("expected error but got result %v", evalResult)
		}
		var scriptErr *verror.Error
		if !errors.As(err, &scriptErr) {
			t.Fatalf("expected script error, got %v", err)
		}
	})

	t.Run("copy --part from advanced index", func(t *testing.T) {
		input := `
			b: [1 2 3 4 5]
			b: next next b
			copy --part 5 b
		`
		want := value.NewBlockVal([]core.Value{
			value.NewIntVal(3), value.NewIntVal(4), value.NewIntVal(5),
		})
		evalResult, err := Evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("copy --part string from advanced index", func(t *testing.T) {
		input := `
			s: "hello"
			s: next next s
			copy --part 5 s
		`
		want := value.NewStrVal("llo")
		evalResult, err := Evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("copy part from tail position yields empty", func(t *testing.T) {
		input := `
			b: [1 2 3]
			b: skip b 3
			c: copy --part 2 b
			length? c
		`
		want := value.NewIntVal(0)
		evalResult, err := Evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})

	t.Run("string copy part from tail position", func(t *testing.T) {
		input := `
			s: "abc"
			s: skip s 3
			c: copy --part 2 s
			length? c
		`
		want := value.NewIntVal(0)
		evalResult, err := Evaluate(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !evalResult.Equals(want) {
			t.Fatalf("expected %v, got %v", want, evalResult)
		}
	})
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
			if strings.Contains(tt.input, "#{") || strings.Contains(tt.input, "append #{}") {
				t.Skip("Binary literals not implemented in parser yet - cannot construct binary series for testing")
				return
			}

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
			name: "tail block",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "tail string",
			input: `str: "hello"
tailStr: tail str
first tailStr`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name: "first at tail position error args",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
			errArgs: []string{"3", "3", "3"},
		},
		{
			name: "first at tail position error args",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
			errArgs: []string{"3", "3", "3"},
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
			name: "tail on already at tail",
			input: `data: [1 2 3]
tailData: tail data
first tailData`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
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
			name:    "block index out of bounds negative",
			input:   "at [1 2 3] 0",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "block index out of bounds too large",
			input:   "at [1 2 3] 4",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "string index out of bounds",
			input:   `at "hi" 3`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "empty block error",
			input:   "at [] 1",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "empty string error",
			input:   `at "" 1`,
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
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
