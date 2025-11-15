package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// Test suite for Feature 038: Block put support
// Contract tests validate block association list mutation via put native
// These tests follow TDD: they MUST FAIL initially before implementation

func TestBlockPut(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "put block update existing pair at head",
			input: "blk: [a 1 b 2]\nput blk 'a 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block update existing pair at later position",
			input: "blk: [a 1 b 2 c 3]\nput blk 'b 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block respect index cursor",
			input: "blk: next [a 1 a 2]\nput blk 'a 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block append when key missing",
			input: "blk: [a 1 b 2]\nput blk 'c 3",
			want:  value.NewIntVal(3),
		},
		{
			name:  "put block remove with none when key exists",
			input: "blk: [a 1 b 2]\nput blk 'a none",
			want:  value.NewNoneVal(),
		},
		{
			name:  "put block remove with none when key missing",
			input: "blk: [a 1 b 2]\nput blk 'c none",
			want:  value.NewNoneVal(),
		},
		{
			name:  "put block odd-length update dangling key",
			input: "blk: [a 1 b]\nput blk 'b 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block odd-length remove dangling key",
			input: "blk: [a 1 b]\nput blk 'b none",
			want:  value.NewNoneVal(),
		},
		{
			name:  "put block first-occurrence with duplicates",
			input: "blk: [a 1 a 2 a 3]\nput blk 'a 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block mixed key types word vs string",
			input: "blk: [a 1 \"b\" 2]\nput blk \"b\" 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block mixed key types integer",
			input: "blk: [a 1 42 2]\nput blk 42 99",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block mixed value types object",
			input: "blk: [a 1 b 2]\nput blk 'b (object [x: 10])",
			want: func() core.Value {
				objResult, _ := Evaluate("object [x: 10]")
				return objResult
			}(),
		},
		{
			name:  "put block respect index with removal",
			input: "blk: next [a 1 a 2]\nput blk 'a none",
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
			// Special case for object test - just check it's an object
			if tt.name == "put block mixed value types object" {
				if result.GetType() != value.TypeObject {
					t.Fatalf("expected object type, got %v", result.GetType())
				}
			} else if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestBlockPutMutation(t *testing.T) {
	mutationTests := []struct {
		name  string
		input string
		want  core.Value
	}{
		{
			name:  "put block mutation verified by select",
			input: "blk: [a 1 b 2]\nput blk 'a 99\nselect blk 'a",
			want:  value.NewIntVal(99),
		},
		{
			name:  "put block append verified by select",
			input: "blk: [a 1 b 2]\nput blk 'c 3\nselect blk 'c",
			want:  value.NewIntVal(3),
		},
		{
			name:  "put block remove verified by select",
			input: "blk: [a 1 b 2]\nput blk 'a none\nselect blk 'a",
			want:  value.NewNoneVal(),
		},
	}

	for _, tt := range mutationTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}
