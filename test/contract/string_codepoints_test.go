package contract

import (
	"errors"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestString_CodepointsOf(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "ASCII string",
			input: "codepoints-of \"ABC\"",
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(65), value.NewIntVal(66), value.NewIntVal(67)}),
		},
		{
			name:  "empty string",
			input: "codepoints-of \"\"",
			want:  value.NewBlockVal([]core.Value{}),
		},
		{
			name:  "emoji",
			input: "codepoints-of \"🚀\"",
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(128640)}),
		},
		{
			name:  "mixed ASCII and emoji",
			input: "codepoints-of \"A🚀B\"",
			want:  value.NewBlockVal([]core.Value{value.NewIntVal(65), value.NewIntVal(128640), value.NewIntVal(66)}),
		},
		{
			name:    "non-string argument",
			input:   "codepoints-of 42",
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

func TestString_CodepointAt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "ASCII character at valid index",
			input: "codepoint-at \"ABC\" 1",
			want:  value.NewIntVal(65),
		},
		{
			name:  "emoji at valid index",
			input: "codepoint-at \"🚀\" 1",
			want:  value.NewIntVal(128640),
		},
		{
			name:  "out of bounds index returns none",
			input: "codepoint-at \"ABC\" 10",
			want:  value.NewNoneVal(),
		},
		{
			name:  "negative index returns none",
			input: "codepoint-at \"ABC\" -1",
			want:  value.NewNoneVal(),
		},
		{
			name:  "default refinement with out of bounds",
			input: "codepoint-at \"ABC\" 10 --default 0",
			want:  value.NewIntVal(0),
		},
		{
			name:  "default refinement with negative index",
			input: "codepoint-at \"ABC\" -1 --default 0",
			want:  value.NewIntVal(0),
		},
		{
			name:    "non-string first argument",
			input:   "codepoint-at 42 1",
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "non-integer second argument",
			input:   "codepoint-at \"ABC\" \"1\"",
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

func TestString_StringFromCodepoints(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
		errID   string
	}{
		{
			name:  "ASCII code points",
			input: "string-from-codepoints [65 66 67]",
			want:  value.NewStrVal("ABC"),
		},
		{
			name:  "empty block",
			input: "string-from-codepoints []",
			want:  value.NewStrVal(""),
		},
		{
			name:  "emoji code point",
			input: "string-from-codepoints [128640]",
			want:  value.NewStrVal("🚀"),
		},
		{
			name:  "mixed ASCII and emoji",
			input: "string-from-codepoints [65 128640 66]",
			want:  value.NewStrVal("A🚀B"),
		},
		{
			name:    "negative code point",
			input:   "string-from-codepoints [-1]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "code point too large",
			input:   "string-from-codepoints [1114112]", // 0x110000
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "surrogate code point",
			input:   "string-from-codepoints [55296]", // 0xD800
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "non-integer in block",
			input:   "string-from-codepoints [\"65\"]",
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "non-block argument",
			input:   "string-from-codepoints 42",
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

func TestString_CodepointsRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "ASCII string",
			input: "\"ABC\"",
		},
		{
			name:  "empty string",
			input: "\"\"",
		},
		{
			name:  "emoji",
			input: "\"🚀\"",
		},
		{
			name:  "mixed content",
			input: "\"Hello 🌍!\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test round-trip: string -> codepoints -> string
			roundTripInput := "string-from-codepoints codepoints-of " + tt.input
			evalResult, err := Evaluate(roundTripInput)
			if err != nil {
				t.Fatalf("unexpected error in round-trip: %v", err)
			}

			originalResult, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error evaluating original: %v", err)
			}

			if !evalResult.Equals(originalResult) {
				t.Fatalf("round-trip failed: expected %v, got %v", originalResult, evalResult)
			}
		})
	}
}
