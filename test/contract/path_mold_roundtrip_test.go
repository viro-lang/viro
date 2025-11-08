package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func comparePathSegments(t *testing.T, expected, actual []value.PathSegment) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Fatalf("segment count mismatch: expected %d, got %d", len(expected), len(actual))
	}

	for i := range expected {
		if expected[i].Type != actual[i].Type {
			t.Errorf("segment[%d] type mismatch: expected %s, got %s",
				i, expected[i].Type.String(), actual[i].Type.String())
			continue
		}

		switch expected[i].Type {
		case value.PathSegmentWord:
			expVal, _ := expected[i].Value.(string)
			actVal, _ := actual[i].Value.(string)
			if expVal != actVal {
				t.Errorf("segment[%d] word value mismatch: expected %q, got %q", i, expVal, actVal)
			}

		case value.PathSegmentIndex:
			expVal, _ := expected[i].Value.(int64)
			actVal, _ := actual[i].Value.(int64)
			if expVal != actVal {
				t.Errorf("segment[%d] index value mismatch: expected %d, got %d", i, expVal, actVal)
			}

		case value.PathSegmentEval:
			expBlock, expOk := expected[i].Value.(*value.BlockValue)
			actBlock, actOk := actual[i].Value.(*value.BlockValue)
			if !expOk || !actOk {
				t.Errorf("segment[%d] eval value is not BlockValue", i)
				continue
			}
			if len(expBlock.Elements) != len(actBlock.Elements) {
				t.Errorf("segment[%d] eval block element count mismatch: expected %d, got %d",
					i, len(expBlock.Elements), len(actBlock.Elements))
				continue
			}
			for j := range expBlock.Elements {
				expMold := expBlock.Elements[j].Mold()
				actMold := actBlock.Elements[j].Mold()
				if expMold != actMold {
					t.Errorf("segment[%d] eval block element[%d] mismatch: expected %q, got %q",
						i, j, expMold, actMold)
				}
			}
		}
	}
}

func TestPathMoldRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		segments []value.PathSegment
	}{
		{
			name: "two-word path",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentWord, Value: "field"},
			},
		},
		{
			name: "three-word path",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "obj"},
				{Type: value.PathSegmentWord, Value: "field"},
				{Type: value.PathSegmentWord, Value: "name"},
			},
		},
		{
			name: "path with index",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentIndex, Value: int64(1)},
			},
		},
		{
			name: "path with multiple indices",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "matrix"},
				{Type: value.PathSegmentIndex, Value: int64(2)},
				{Type: value.PathSegmentIndex, Value: int64(3)},
			},
		},
		{
			name: "path with eval segment word",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})},
			},
		},
		{
			name: "path with eval segment string",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewStrVal("idx"),
				})},
			},
		},
		{
			name: "nested eval segments",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})},
			},
		},
		{
			name: "mixed segments",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "obj"},
				{Type: value.PathSegmentIndex, Value: int64(2)},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("key"),
				})},
				{Type: value.PathSegmentWord, Value: "name"},
			},
		},
		{
			name: "eval then index",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})},
				{Type: value.PathSegmentIndex, Value: int64(3)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := value.NewPath(tt.segments, nil)

			molded := original.Mold()

			vals, _, err := parse.Parse(molded)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if len(vals) != 1 {
				t.Fatalf("expected 1 value, got %d", len(vals))
			}

			if vals[0].GetType() != value.TypePath {
				t.Fatalf("expected path type, got %s", value.TypeToString(vals[0].GetType()))
			}

			parsed, ok := value.AsPath(vals[0])
			if !ok {
				t.Fatalf("failed to extract path")
			}

			if parsed.Mold() != molded {
				t.Errorf("mold mismatch: got %q, want %q", parsed.Mold(), molded)
			}

			comparePathSegments(t, tt.segments, parsed.Segments)
		})
	}
}

func TestGetPathMoldRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		segments []value.PathSegment
	}{
		{
			name: "two-word get-path",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentWord, Value: "field"},
			},
		},
		{
			name: "three-word get-path",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "obj"},
				{Type: value.PathSegmentWord, Value: "field"},
				{Type: value.PathSegmentWord, Value: "name"},
			},
		},
		{
			name: "get-path with index",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentIndex, Value: int64(1)},
			},
		},
		{
			name: "get-path with multiple indices",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "matrix"},
				{Type: value.PathSegmentIndex, Value: int64(2)},
				{Type: value.PathSegmentIndex, Value: int64(3)},
			},
		},
		{
			name: "get-path with eval segment word",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})},
			},
		},
		{
			name: "get-path with eval segment string",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewStrVal("idx"),
				})},
			},
		},
		{
			name: "nested eval segments",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})},
			},
		},
		{
			name: "mixed segments",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "obj"},
				{Type: value.PathSegmentIndex, Value: int64(2)},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("key"),
				})},
				{Type: value.PathSegmentWord, Value: "name"},
			},
		},
		{
			name: "eval then index",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})},
				{Type: value.PathSegmentIndex, Value: int64(3)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := value.NewGetPath(tt.segments, nil)

			molded := original.Mold()

			vals, _, err := parse.Parse(molded)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if len(vals) != 1 {
				t.Fatalf("expected 1 value, got %d", len(vals))
			}

			if vals[0].GetType() != value.TypeGetPath {
				t.Fatalf("expected get-path type, got %s", value.TypeToString(vals[0].GetType()))
			}

			parsed, ok := value.AsGetPath(vals[0])
			if !ok {
				t.Fatalf("failed to extract get-path")
			}

			if parsed.Mold() != molded {
				t.Errorf("mold mismatch: got %q, want %q", parsed.Mold(), molded)
			}

			comparePathSegments(t, tt.segments, parsed.Segments)
		})
	}
}

func TestSetPathMoldRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		segments []value.PathSegment
	}{
		{
			name: "two-word set-path",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentWord, Value: "field"},
			},
		},
		{
			name: "three-word set-path",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "obj"},
				{Type: value.PathSegmentWord, Value: "field"},
				{Type: value.PathSegmentWord, Value: "name"},
			},
		},
		{
			name: "set-path with index",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentIndex, Value: int64(1)},
			},
		},
		{
			name: "set-path with multiple indices",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "matrix"},
				{Type: value.PathSegmentIndex, Value: int64(2)},
				{Type: value.PathSegmentIndex, Value: int64(3)},
			},
		},
		{
			name: "set-path with eval segment word",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})},
			},
		},
		{
			name: "set-path with eval segment string",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewStrVal("idx"),
				})},
			},
		},
		{
			name: "nested eval segments",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})},
			},
		},
		{
			name: "mixed segments",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "obj"},
				{Type: value.PathSegmentIndex, Value: int64(2)},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("key"),
				})},
				{Type: value.PathSegmentWord, Value: "name"},
			},
		},
		{
			name: "eval then index",
			segments: []value.PathSegment{
				{Type: value.PathSegmentWord, Value: "data"},
				{Type: value.PathSegmentEval, Value: value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})},
				{Type: value.PathSegmentIndex, Value: int64(3)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := value.NewSetPath(tt.segments, nil)

			molded := original.Mold()

			vals, _, err := parse.Parse(molded)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if len(vals) != 1 {
				t.Fatalf("expected 1 value, got %d", len(vals))
			}

			if vals[0].GetType() != value.TypeSetPath {
				t.Fatalf("expected set-path type, got %s", value.TypeToString(vals[0].GetType()))
			}

			parsed, ok := value.AsSetPath(vals[0])
			if !ok {
				t.Fatalf("failed to extract set-path")
			}

			if parsed.Mold() != molded {
				t.Errorf("mold mismatch: got %q, want %q", parsed.Mold(), molded)
			}

			comparePathSegments(t, tt.segments, parsed.Segments)
		})
	}
}
