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
			name:    "one element block error",
			input:   "second [42]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "empty block error",
			input:   "second []",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "one character string error",
			input:   "second \"x\"",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "second 42",
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
			name:    "two element block error",
			input:   "third [1 2]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "empty block error",
			input:   "third []",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "two character string error",
			input:   "third \"ab\"",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "third 42",
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
			name:    "three element block error",
			input:   "fourth [1 2 3]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "fourth 42",
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
			name:    "five element block error",
			input:   "sixth [1 2 3 4 5]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "sixth 42",
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
			name:    "six element block error",
			input:   "seventh [1 2 3 4 5 6]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "seventh 42",
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
			name:    "seven element block error",
			input:   "eighth [1 2 3 4 5 6 7]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "eighth 42",
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
			name:    "eight element block error",
			input:   "ninth [1 2 3 4 5 6 7 8]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "ninth 42",
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
			name:    "nine element block error",
			input:   "tenth [1 2 3 4 5 6 7 8 9]",
			wantErr: true,
			errID:   verror.ErrIDOutOfBounds,
		},
		{
			name:    "non series error",
			input:   "tenth 42",
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
