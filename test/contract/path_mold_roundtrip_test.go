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
			expVal, expOk := expected[i].AsWord()
			actVal, actOk := actual[i].AsWord()
			if !expOk || !actOk {
				t.Errorf("segment[%d] word value extraction failed: expOk=%v, actOk=%v", i, expOk, actOk)
				continue
			}
			if expVal != actVal {
				t.Errorf("segment[%d] word value mismatch: expected %q, got %q", i, expVal, actVal)
			}

		case value.PathSegmentIndex:
			expVal, expOk := expected[i].AsIndex()
			actVal, actOk := actual[i].AsIndex()
			if !expOk || !actOk {
				t.Errorf("segment[%d] index value extraction failed: expOk=%v, actOk=%v", i, expOk, actOk)
				continue
			}
			if expVal != actVal {
				t.Errorf("segment[%d] index value mismatch: expected %d, got %d", i, expVal, actVal)
			}

		case value.PathSegmentEval:
			expBlock, expOk := expected[i].AsEvalBlock()
			actBlock, actOk := actual[i].AsEvalBlock()
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

		default:
			t.Fatalf("unknown segment type: %s", expected[i].Type.String())
		}
	}
}

type pathKind struct {
	name         string
	constructor  func([]value.PathSegment, core.Value) core.Value
	expectedType core.ValueType
	extractor    func(core.Value) (core.Value, bool)
	segments     func(core.Value) []value.PathSegment
}

func TestPathMoldRoundtrip(t *testing.T) {
	pathKinds := []pathKind{
		{
			name:         "path",
			constructor:  func(segments []value.PathSegment, base core.Value) core.Value { return value.NewPath(segments, base) },
			expectedType: value.TypePath,
			extractor: func(v core.Value) (core.Value, bool) {
				p, ok := value.AsPath(v)
				return p, ok
			},
			segments: func(v core.Value) []value.PathSegment {
				p := v.(*value.PathExpression)
				return p.Segments
			},
		},
		{
			name: "get-path",
			constructor: func(segments []value.PathSegment, base core.Value) core.Value {
				return value.NewGetPath(segments, base)
			},
			expectedType: value.TypeGetPath,
			extractor: func(v core.Value) (core.Value, bool) {
				p, ok := value.AsGetPath(v)
				return p, ok
			},
			segments: func(v core.Value) []value.PathSegment {
				p := v.(*value.GetPathExpression)
				return p.Segments
			},
		},
		{
			name: "set-path",
			constructor: func(segments []value.PathSegment, base core.Value) core.Value {
				return value.NewSetPath(segments, base)
			},
			expectedType: value.TypeSetPath,
			extractor: func(v core.Value) (core.Value, bool) {
				p, ok := value.AsSetPath(v)
				return p, ok
			},
			segments: func(v core.Value) []value.PathSegment {
				p := v.(*value.SetPathExpression)
				return p.Segments
			},
		},
	}

	tests := []struct {
		name     string
		segments []value.PathSegment
	}{
		{
			name: "two-word path",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewWordSegment("field"),
			},
		},
		{
			name: "three-word path",
			segments: []value.PathSegment{
				value.NewWordSegment("obj"),
				value.NewWordSegment("field"),
				value.NewWordSegment("name"),
			},
		},
		{
			name: "path with index",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewIndexSegment(1),
			},
		},
		{
			name: "path with multiple indices",
			segments: []value.PathSegment{
				value.NewWordSegment("matrix"),
				value.NewIndexSegment(2),
				value.NewIndexSegment(3),
			},
		},
		{
			name: "path with eval segment word",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})),
			},
		},
		{
			name: "path with eval segment string",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewStrVal("idx"),
				})),
			},
		},
		{
			name: "nested eval segments",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})),
			},
		},
		{
			name: "mixed segments",
			segments: []value.PathSegment{
				value.NewWordSegment("obj"),
				value.NewIndexSegment(2),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("key"),
				})),
				value.NewWordSegment("name"),
			},
		},
		{
			name: "eval then index",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("idx"),
				})),
				value.NewIndexSegment(3),
			},
		},
		{
			name: "word eval index eval",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("field"),
				})),
				value.NewIndexSegment(1),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("key"),
				})),
			},
		},
		{
			name: "multiple consecutive evals",
			segments: []value.PathSegment{
				value.NewWordSegment("data"),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("a"),
				})),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("b"),
				})),
				value.NewEvalSegment(value.NewBlockValue([]core.Value{
					value.NewWordVal("c"),
				})),
			},
		},
	}

	for _, kind := range pathKinds {
		for _, tt := range tests {
			t.Run(kind.name+"/"+tt.name, func(t *testing.T) {
				original := kind.constructor(tt.segments, nil)

				molded := original.Mold()

				vals, _, err := parse.Parse(molded)
				if err != nil {
					t.Fatalf("parse error: %v", err)
				}

				if len(vals) != 1 {
					t.Fatalf("expected 1 value, got %d", len(vals))
				}

				if vals[0].GetType() != kind.expectedType {
					t.Fatalf("expected %s type, got %s", value.TypeToString(kind.expectedType), value.TypeToString(vals[0].GetType()))
				}

				parsed, ok := kind.extractor(vals[0])
				if !ok {
					t.Fatalf("failed to extract %s", kind.name)
				}

				if parsed.Mold() != molded {
					t.Errorf("mold mismatch: got %q, want %q", parsed.Mold(), molded)
				}

				segments := kind.segments(parsed)

				comparePathSegments(t, tt.segments, segments)
			})
		}
	}
}

func TestPathMoldRoundtripErrors(t *testing.T) {
	invalidMolds := []string{
		"data.(",
		":data.",
		"data..field",
		"data.(field",
		"data.field)",
		"data.(field).",
		"data.()",
		"data.(field).()",
	}

	for _, mold := range invalidMolds {
		t.Run("invalid_"+mold, func(t *testing.T) {
			_, _, err := parse.Parse(mold)
			if err == nil {
				t.Errorf("expected parse error for invalid mold %q, but parsing succeeded", mold)
			}
		})
	}
}
