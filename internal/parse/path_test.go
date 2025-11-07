package parse

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestPathTokenization validates T090: path segment tokenizer
func TestPathTokenization(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  core.ValueType
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
		{
			name:     "path with eval segment",
			input:    "foo.(field).bar",
			wantType: value.TypePath,
			wantSegs: 3,
			checkPath: func(t *testing.T, path *value.PathExpression) {
				if len(path.Segments) != 3 {
					t.Fatalf("Expected 3 segments, got %d", len(path.Segments))
				}
				if path.Segments[0].Type != value.PathSegmentWord || path.Segments[0].Value != "foo" {
					t.Errorf("Expected first segment to be word 'foo'")
				}
				seg := path.Segments[1]
				if seg.Type != value.PathSegmentEval {
					t.Fatalf("Expected second segment to be eval, got %v", seg.Type)
				}
				block, ok := seg.Value.(*value.BlockValue)
				if !ok {
					t.Fatalf("Expected eval segment to store block")
				}
				if len(block.Elements) != 1 {
					t.Fatalf("Expected eval block to have 1 element, got %d", len(block.Elements))
				}
				if block.Elements[0].GetType() != value.TypeWord {
					t.Errorf("Expected eval element to be word, got %s", value.TypeToString(block.Elements[0].GetType()))
				}
				if path.Segments[2].Type != value.PathSegmentWord || path.Segments[2].Value != "bar" {
					t.Errorf("Expected third segment to be word 'bar'")
				}
			},
		},
		{
			name:     "path with nested eval segment",
			input:    "foo.(bar.(baz)).qux",
			wantType: value.TypePath,
			wantSegs: 3,
			checkPath: func(t *testing.T, path *value.PathExpression) {
				if len(path.Segments) != 3 {
					t.Fatalf("Expected 3 segments, got %d", len(path.Segments))
				}
				if path.Segments[0].Type != value.PathSegmentWord || path.Segments[0].Value != "foo" {
					t.Errorf("Expected first segment to be word 'foo'")
				}
				seg := path.Segments[1]
				if seg.Type != value.PathSegmentEval {
					t.Fatalf("Expected second segment to be eval, got %v", seg.Type)
				}
				block, ok := seg.Value.(*value.BlockValue)
				if !ok {
					t.Fatalf("Expected eval segment to store block")
				}
				if len(block.Elements) != 1 {
					t.Fatalf("Expected eval block to have 1 element, got %d", len(block.Elements))
				}
				nested := block.Elements[0]
				if nested.GetType() != value.TypePath {
					t.Fatalf("Expected nested value to be path, got %s", value.TypeToString(nested.GetType()))
				}
				nestedPath, ok := value.AsPath(nested)
				if !ok {
					t.Fatalf("Failed to extract nested path")
				}
				if len(nestedPath.Segments) != 2 {
					t.Fatalf("Expected nested path to have 2 segments, got %d", len(nestedPath.Segments))
				}
				if nestedPath.Segments[0].Type != value.PathSegmentWord || nestedPath.Segments[0].Value != "bar" {
					t.Errorf("Expected nested first segment to be 'bar'")
				}
				if nestedPath.Segments[1].Type != value.PathSegmentEval {
					t.Fatalf("Expected nested second segment to be eval, got %v", nestedPath.Segments[1].Type)
				}
				if path.Segments[2].Type != value.PathSegmentWord || path.Segments[2].Value != "qux" {
					t.Errorf("Expected third segment to be word 'qux'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, _, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if len(vals) != 1 {
				t.Fatalf("Expected 1 value, got %d", len(vals))
			}

			val := vals[0]
			if val.GetType() != tt.wantType {
				t.Fatalf("Expected type %s, got %s", value.TypeToString(tt.wantType), value.TypeToString(val.GetType()))
			}

			path, ok := value.AsPath(val)
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
		check func(*testing.T, []core.Value)
	}{
		{
			name:  "set-word followed by word",
			input: "name: \"Alice\"",
			check: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Fatalf("Expected 2 values, got %d", len(vals))
				}
				if vals[0].GetType() != value.TypeSetWord {
					t.Errorf("Expected TypeSetWord, got %s", value.TypeToString(vals[0].GetType()))
				}
				if vals[1].GetType() != value.TypeString {
					t.Errorf("Expected TypeString, got %s", value.TypeToString(vals[1].GetType()))
				}
			},
		},
		{
			name:  "path followed by set-word",
			input: "obj.name user: \"Bob\"",
			check: func(t *testing.T, vals []core.Value) {
				if len(vals) != 3 {
					t.Fatalf("Expected 3 values, got %d", len(vals))
				}
				if vals[0].GetType() != value.TypePath {
					t.Errorf("Expected TypePath, got %s", value.TypeToString(vals[0].GetType()))
				}
				if vals[1].GetType() != value.TypeSetWord {
					t.Errorf("Expected TypeSetWord, got %s", value.TypeToString(vals[1].GetType()))
				}
				if vals[2].GetType() != value.TypeString {
					t.Errorf("Expected TypeString, got %s", value.TypeToString(vals[2].GetType()))
				}
			},
		},
		{
			name:  "set-path (path used for assignment)",
			input: "obj.name: \"Alice\"",
			check: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Fatalf("Expected 2 values, got %d: %v", len(vals), vals)
				}
				if vals[0].GetType() != value.TypeSetPath {
					t.Errorf("Expected TypeSetPath for set-path, got %s", value.TypeToString(vals[0].GetType()))
				}
				// The set-path should be a proper path expression
				setPath, ok := value.AsSetPath(vals[0])
				if !ok {
					t.Errorf("Expected set-path value")
				} else if len(setPath.Segments) != 2 {
					t.Errorf("Expected 2 segments, got %d", len(setPath.Segments))
				}
				if vals[1].GetType() != value.TypeString {
					t.Errorf("Expected TypeString, got %s", value.TypeToString(vals[1].GetType()))
				}
			},
		},
		{
			name:  "nested set-path",
			input: "user.address.city: \"Portland\"",
			check: func(t *testing.T, vals []core.Value) {
				if len(vals) != 2 {
					t.Fatalf("Expected 2 values, got %d", len(vals))
				}
				if vals[0].GetType() != value.TypeSetPath {
					t.Errorf("Expected TypeSetPath, got %s", value.TypeToString(vals[0].GetType()))
				}
				setPath, ok := value.AsSetPath(vals[0])
				if !ok {
					t.Errorf("Expected set-path value")
				} else if len(setPath.Segments) != 3 {
					t.Errorf("Expected 3 segments, got %d", len(setPath.Segments))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, _, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if tt.check != nil {
				tt.check(t, vals)
			}
		})
	}
}
