package contract

import (
	"errors"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestSeries_Second(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block second element",
			input: "second [1 2 3]",
			want:  value.NewIntVal(2),
		},
		{
			name:  "string second character",
			input: "second \"hello\"",
			want:  value.NewStrVal("e"),
		},
		{
			name:  "binary second element",
			input: "second #{DEADBEEF}",
			want:  value.NewIntVal(173),
		},
		{
			name:  "one element block returns none",
			input: "second [42]",
			want:  value.NewNoneVal(),
		},
		{
			name:  "empty block returns none",
			input: "second []",
			want:  value.NewNoneVal(),
		},
		{
			name:  "one character string returns none",
			input: "second \"x\"",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "second 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		// Test skip functionality
		{
			name:  "second skip one position",
			input: "second skip [1 2 3] 1",
			want:  value.NewIntVal(3),
		},
		{
			name:  "second skip at start",
			input: "second skip [1 2 3] 0",
			want:  value.NewIntVal(2),
		},
		{
			name:    "second skip out of bounds",
			input:   "second skip [1 2 3] 2",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "second skip with longer series",
			input: "second skip [1 2 3 4 5] 2",
			want:  value.NewIntVal(4),
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

func TestSeries_Third(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block third element",
			input: "third [1 2 3]",
			want:  value.NewIntVal(3),
		},
		{
			name:  "string third character",
			input: "third \"hello\"",
			want:  value.NewStrVal("l"),
		},
		{
			name:  "binary third element",
			input: "third #{DEADBEEF}",
			want:  value.NewIntVal(190),
		},
		{
			name:  "two element block returns none",
			input: "third [1 2]",
			want:  value.NewNoneVal(),
		},
		{
			name:  "empty block returns none",
			input: "third []",
			want:  value.NewNoneVal(),
		},
		{
			name:  "two character string returns none",
			input: "third \"ab\"",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "third 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		// Test skip functionality
		{
			name:    "third skip one position",
			input:   "third skip [1 2 3] 1",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "third skip at start",
			input: "third skip [1 2 3] 0",
			want:  value.NewIntVal(3),
		},
		{
			name:  "third skip with longer series",
			input: "third skip [1 2 3 4 5] 1",
			want:  value.NewIntVal(4),
		},
		{
			name:    "third skip out of bounds",
			input:   "third skip [1 2 3] 2",
			wantErr: false,
			want:    value.NewNoneVal(),
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

func TestSeries_Fourth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block fourth element",
			input: "fourth [1 2 3 4]",
			want:  value.NewIntVal(4),
		},
		{
			name:  "string fourth character",
			input: "fourth \"hello\"",
			want:  value.NewStrVal("l"),
		},
		{
			name:  "binary fourth element",
			input: "fourth #{DEADBEEF}",
			want:  value.NewIntVal(239),
		},
		{
			name:  "three element block returns none",
			input: "fourth [1 2 3]",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "fourth 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "fourth skip one position",
			input: "fourth skip [1 2 3 4 5 6] 1",
			want:  value.NewIntVal(5),
		},
		{
			name:  "fourth skip at start",
			input: "fourth skip [1 2 3 4 5] 0",
			want:  value.NewIntVal(4),
		},
		{
			name:    "fourth skip out of bounds returns none",
			input:   "fourth skip [1 2 3] 1",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "fourth skip with longer series",
			input: "fourth skip [1 2 3 4 5 6 7 8] 2",
			want:  value.NewIntVal(6),
		},
		{
			name:    "fourth skip near tail boundary",
			input:   "fourth skip [1 2 3 4 5 6] 3",
			wantErr: false,
			want:    value.NewNoneVal(),
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

func TestSeries_Sixth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block sixth element",
			input: "sixth [1 2 3 4 5 6]",
			want:  value.NewIntVal(6),
		},
		{
			name:  "string sixth character",
			input: "sixth \"hello world\"",
			want:  value.NewStrVal(" "),
		},
		{
			name:  "binary sixth element",
			input: "sixth #{DEADBEEF0102}",
			want:  value.NewIntVal(2),
		},
		{
			name:  "five element block returns none",
			input: "sixth [1 2 3 4 5]",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "sixth 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "sixth skip one position",
			input: "sixth skip [1 2 3 4 5 6 7 8] 1",
			want:  value.NewIntVal(7),
		},
		{
			name:  "sixth skip at start",
			input: "sixth skip [1 2 3 4 5 6 7] 0",
			want:  value.NewIntVal(6),
		},
		{
			name:    "sixth skip out of bounds returns none",
			input:   "sixth skip [1 2 3 4 5] 1",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "sixth skip with longer series",
			input: "sixth skip [1 2 3 4 5 6 7 8 9 10] 2",
			want:  value.NewIntVal(8),
		},
		{
			name:    "sixth skip near tail boundary",
			input:   "sixth skip [1 2 3 4 5 6 7] 2",
			wantErr: false,
			want:    value.NewNoneVal(),
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

func TestSeries_Seventh(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block seventh element",
			input: "seventh [1 2 3 4 5 6 7]",
			want:  value.NewIntVal(7),
		},
		{
			name:  "string seventh character",
			input: "seventh \"hello world\"",
			want:  value.NewStrVal("w"),
		},
		{
			name:  "binary seventh element",
			input: "seventh #{DEADBEEF010203}",
			want:  value.NewIntVal(3),
		},
		{
			name:  "six element block returns none",
			input: "seventh [1 2 3 4 5 6]",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "seventh 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "seventh skip one position",
			input: "seventh skip [1 2 3 4 5 6 7 8 9] 1",
			want:  value.NewIntVal(8),
		},
		{
			name:  "seventh skip at start",
			input: "seventh skip [1 2 3 4 5 6 7 8] 0",
			want:  value.NewIntVal(7),
		},
		{
			name:    "seventh skip out of bounds returns none",
			input:   "seventh skip [1 2 3 4 5 6] 1",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "seventh skip with longer series",
			input: "seventh skip [1 2 3 4 5 6 7 8 9 10 11] 2",
			want:  value.NewIntVal(9),
		},
		{
			name:    "seventh skip near tail boundary",
			input:   "seventh skip [1 2 3 4 5 6 7 8] 2",
			wantErr: false,
			want:    value.NewNoneVal(),
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

func TestSeries_Eighth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block eighth element",
			input: "eighth [1 2 3 4 5 6 7 8]",
			want:  value.NewIntVal(8),
		},
		{
			name:  "string eighth character",
			input: "eighth \"hello world\"",
			want:  value.NewStrVal("o"),
		},
		{
			name:  "binary eighth element",
			input: "eighth #{DEADBEEF0102030405}",
			want:  value.NewIntVal(4),
		},
		{
			name:  "seven element block returns none",
			input: "eighth [1 2 3 4 5 6 7]",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "eighth 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "eighth skip one position",
			input: "eighth skip [1 2 3 4 5 6 7 8 9 10] 1",
			want:  value.NewIntVal(9),
		},
		{
			name:  "eighth skip at start",
			input: "eighth skip [1 2 3 4 5 6 7 8 9] 0",
			want:  value.NewIntVal(8),
		},
		{
			name:    "eighth skip out of bounds returns none",
			input:   "eighth skip [1 2 3 4 5 6 7] 1",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "eighth skip with longer series",
			input: "eighth skip [1 2 3 4 5 6 7 8 9 10 11 12] 2",
			want:  value.NewIntVal(10),
		},
		{
			name:    "eighth skip near tail boundary",
			input:   "eighth skip [1 2 3 4 5 6 7 8 9] 2",
			wantErr: false,
			want:    value.NewNoneVal(),
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

func TestSeries_Ninth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block ninth element",
			input: "ninth [1 2 3 4 5 6 7 8 9]",
			want:  value.NewIntVal(9),
		},
		{
			name:  "string ninth character",
			input: "ninth \"hello world\"",
			want:  value.NewStrVal("r"),
		},
		{
			name:  "binary ninth element",
			input: "ninth #{DEADBEEF010203040506}",
			want:  value.NewIntVal(5),
		},
		{
			name:  "eight element block returns none",
			input: "ninth [1 2 3 4 5 6 7 8]",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "ninth 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "ninth skip one position",
			input: "ninth skip [1 2 3 4 5 6 7 8 9 10 11] 1",
			want:  value.NewIntVal(10),
		},
		{
			name:  "ninth skip at start",
			input: "ninth skip [1 2 3 4 5 6 7 8 9 10] 0",
			want:  value.NewIntVal(9),
		},
		{
			name:    "ninth skip out of bounds returns none",
			input:   "ninth skip [1 2 3 4 5 6 7 8] 1",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "ninth skip with longer series",
			input: "ninth skip [1 2 3 4 5 6 7 8 9 10 11 12 13] 2",
			want:  value.NewIntVal(11),
		},
		{
			name:    "ninth skip near tail boundary",
			input:   "ninth skip [1 2 3 4 5 6 7 8 9 10] 2",
			wantErr: false,
			want:    value.NewNoneVal(),
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

func TestSeries_Tenth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "block tenth element",
			input: "tenth [1 2 3 4 5 6 7 8 9 10]",
			want:  value.NewIntVal(10),
		},
		{
			name:  "string tenth character",
			input: "tenth \"hello world\"",
			want:  value.NewStrVal("l"),
		},
		{
			name:  "binary tenth element",
			input: "tenth #{DEADBEEF01020304050607}",
			want:  value.NewIntVal(6),
		},
		{
			name:  "nine element block returns none",
			input: "tenth [1 2 3 4 5 6 7 8 9]",
			want:  value.NewNoneVal(),
		},
		{
			name:    "non series error",
			input:   "tenth 42",
			wantErr: true,
			errID:   verror.ErrIDActionNoImpl,
		},
		{
			name:  "tenth skip one position",
			input: "tenth skip [1 2 3 4 5 6 7 8 9 10 11 12] 1",
			want:  value.NewIntVal(11),
		},
		{
			name:  "tenth skip at start",
			input: "tenth skip [1 2 3 4 5 6 7 8 9 10 11] 0",
			want:  value.NewIntVal(10),
		},
		{
			name:    "tenth skip out of bounds returns none",
			input:   "tenth skip [1 2 3 4 5 6 7 8 9] 1",
			wantErr: false,
			want:    value.NewNoneVal(),
		},
		{
			name:  "tenth skip with longer series",
			input: "tenth skip [1 2 3 4 5 6 7 8 9 10 11 12 13 14] 2",
			want:  value.NewIntVal(12),
		},
		{
			name:    "tenth skip near tail boundary",
			input:   "tenth skip [1 2 3 4 5 6 7 8 9 10 11] 2",
			wantErr: false,
			want:    value.NewNoneVal(),
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
