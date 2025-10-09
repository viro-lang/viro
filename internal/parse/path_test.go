package parse

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestPathTokenization validates T090: path segment tokenizer
func TestPathTokenization(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  value.ValueType
		wantSegs  int
		checkPath func(*testing.T, *value.PathExpression)
	}{
		{
			name:     "simple path",
			input:    "obj.name",
			wantType: value.TypePath,
			wantSegs: 2,
			checkPath: func(t *testing.T, path *value.PathExpression) {
				if len(path.Segments) != 2 {
					t.Fatalf("Expected 2 segments, got %d", len(path.Segments))
				}
				if path.Segments[0].Type != value.PathSegmentWord || path.Segments[0].Value != "obj" {
					t.Errorf("Expected first segment to be word 'obj', got %v %v", path.Segments[0].Type, path.Segments[0].Value)
				}
				if path.Segments[1].Type != value.PathSegmentWord || path.Segments[1].Value != "name" {
					t.Errorf("Expected second segment to be word 'name', got %v %v", path.Segments[1].Type, path.Segments[1].Value)
				}
			},
		},
		{
			name:     "nested path",
			input:    "user.address.city",
			wantType: value.TypePath,
			wantSegs: 3,
			checkPath: func(t *testing.T, path *value.PathExpression) {
				if len(path.Segments) != 3 {
					t.Fatalf("Expected 3 segments, got %d", len(path.Segments))
				}
				expected := []string{"user", "address", "city"}
				for i, exp := range expected {
					if path.Segments[i].Type != value.PathSegmentWord || path.Segments[i].Value != exp {
						t.Errorf("Segment %d: expected word %q, got %v %v", i, exp, path.Segments[i].Type, path.Segments[i].Value)
					}
				}
			},
		},
		{
			name:     "path with index",
			input:    "data.2",
			wantType: value.TypePath,
			wantSegs: 2,
			checkPath: func(t *testing.T, path *value.PathExpression) {
				if len(path.Segments) != 2 {
					t.Fatalf("Expected 2 segments, got %d", len(path.Segments))
				}
				if path.Segments[0].Type != value.PathSegmentWord || path.Segments[0].Value != "data" {
					t.Errorf("Expected first segment to be word 'data', got %v %v", path.Segments[0].Type, path.Segments[0].Value)
				}
				if path.Segments[1].Type != value.PathSegmentIndex || path.Segments[1].Value != int64(2) {
					t.Errorf("Expected second segment to be index 2, got %v %v", path.Segments[1].Type, path.Segments[1].Value)
				}
			},
		},
		{
			name:     "path with multiple indices",
			input:    "matrix.2.3",
			wantType: value.TypePath,
			wantSegs: 3,
			checkPath: func(t *testing.T, path *value.PathExpression) {
				if len(path.Segments) != 3 {
					t.Fatalf("Expected 3 segments, got %d", len(path.Segments))
				}
				if path.Segments[0].Type != value.PathSegmentWord || path.Segments[0].Value != "matrix" {
					t.Errorf("Expected first segment to be word 'matrix'")
				}
				if path.Segments[1].Type != value.PathSegmentIndex || path.Segments[1].Value != int64(2) {
					t.Errorf("Expected second segment to be index 2")
				}
				if path.Segments[2].Type != value.PathSegmentIndex || path.Segments[2].Value != int64(3) {
					t.Errorf("Expected third segment to be index 3")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if len(vals) != 1 {
				t.Fatalf("Expected 1 value, got %d", len(vals))
			}

			val := vals[0]
			if val.Type != tt.wantType {
				t.Fatalf("Expected type %s, got %s", tt.wantType, val.Type)
			}

			path, ok := val.AsPath()
			if !ok {
				t.Fatalf("Failed to extract path from value")
			}

			if tt.checkPath != nil {
				tt.checkPath(t, path)
			}
		})
	}
}

// Test that paths don't break other parsing
func TestPathsWithOtherSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, []value.Value)
	}{
		{
			name:  "set-word followed by word",
			input: "name: \"Alice\"",
			check: func(t *testing.T, vals []value.Value) {
				if len(vals) != 2 {
					t.Fatalf("Expected 2 values, got %d", len(vals))
				}
				if vals[0].Type != value.TypeSetWord {
					t.Errorf("Expected TypeSetWord, got %s", vals[0].Type)
				}
				if vals[1].Type != value.TypeString {
					t.Errorf("Expected TypeString, got %s", vals[1].Type)
				}
			},
		},
		{
			name:  "path followed by set-word",
			input: "obj.name user: \"Bob\"",
			check: func(t *testing.T, vals []value.Value) {
				if len(vals) != 3 {
					t.Fatalf("Expected 3 values, got %d", len(vals))
				}
				if vals[0].Type != value.TypePath {
					t.Errorf("Expected TypePath, got %s", vals[0].Type)
				}
				if vals[1].Type != value.TypeSetWord {
					t.Errorf("Expected TypeSetWord, got %s", vals[1].Type)
				}
				if vals[2].Type != value.TypeString {
					t.Errorf("Expected TypeString, got %s", vals[2].Type)
				}
			},
		},
		{
			name:  "set-path (path used for assignment)",
			input: "obj.name: \"Alice\"",
			check: func(t *testing.T, vals []value.Value) {
				if len(vals) != 2 {
					t.Fatalf("Expected 2 values, got %d: %v", len(vals), vals)
				}
				if vals[0].Type != value.TypeSetWord {
					t.Errorf("Expected TypeSetWord for set-path, got %s", vals[0].Type)
				}
				// The set-word should contain the full path string
				word, ok := vals[0].AsWord()
				if !ok || word != "obj.name" {
					t.Errorf("Expected set-word to be 'obj.name', got %q", word)
				}
				if vals[1].Type != value.TypeString {
					t.Errorf("Expected TypeString, got %s", vals[1].Type)
				}
			},
		},
		{
			name:  "nested set-path",
			input: "user.address.city: \"Portland\"",
			check: func(t *testing.T, vals []value.Value) {
				if len(vals) != 2 {
					t.Fatalf("Expected 2 values, got %d", len(vals))
				}
				if vals[0].Type != value.TypeSetWord {
					t.Errorf("Expected TypeSetWord, got %s", vals[0].Type)
				}
				word, ok := vals[0].AsWord()
				if !ok || word != "user.address.city" {
					t.Errorf("Expected set-word to be 'user.address.city', got %q", word)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if tt.check != nil {
				tt.check(t, vals)
			}
		})
	}
}
