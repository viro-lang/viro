package value

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
)

func TestPathExpressionMold(t *testing.T) {
	tests := []struct {
		name     string
		segments []PathSegment
		want     string
	}{
		{
			name: "simple word path",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentWord, Value: "field"},
			},
			want: "data.field",
		},
		{
			name: "path with index",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentIndex, Value: int64(1)},
			},
			want: "data.1",
		},
		{
			name: "path with eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewStrVal("idx")})},
			},
			want: "data.(\"idx\")",
		},
		{
			name: "path with nested eval segments",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("field")})},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
			},
			want: "data.(field).(idx)",
		},
		{
			name: "path with eval index ordering",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
				{Type: PathSegmentIndex, Value: int64(3)},
			},
			want: "data.(idx).3",
		},
		{
			name: "mixed path segments",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "obj"},
				{Type: PathSegmentIndex, Value: int64(2)},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("key")})},
				{Type: PathSegmentWord, Value: "name"},
			},
			want: "obj.2.(key).name",
		},
		{
			name: "path with non-block eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: "not-a-block"},
			},
			want: "data.(eval)",
		},
		{
			name:     "empty path",
			segments: []PathSegment{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewPath(tt.segments, nil)
			got := path.Mold()
			if got != tt.want {
				t.Errorf("PathExpression.Mold() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPathExpressionMold(t *testing.T) {
	tests := []struct {
		name     string
		segments []PathSegment
		want     string
	}{
		{
			name: "simple get-path",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentWord, Value: "field"},
			},
			want: ":data.field",
		},
		{
			name: "get-path with eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
			},
			want: ":data.(idx)",
		},
		{
			name: "get-path with nested eval segments",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("field")})},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
			},
			want: ":data.(field).(idx)",
		},
		{
			name: "get-path with eval index ordering",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
				{Type: PathSegmentIndex, Value: int64(3)},
			},
			want: ":data.(idx).3",
		},
		{
			name: "get-path with non-block eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: "not-a-block"},
			},
			want: ":data.(eval)",
		},
		{
			name:     "empty get-path",
			segments: []PathSegment{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewGetPath(tt.segments, nil)
			got := path.Mold()
			if got != tt.want {
				t.Errorf("GetPathExpression.Mold() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetPathExpressionMold(t *testing.T) {
	tests := []struct {
		name     string
		segments []PathSegment
		want     string
	}{
		{
			name: "simple set-path",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentWord, Value: "field"},
			},
			want: "data.field:",
		},
		{
			name: "set-path with index",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentIndex, Value: int64(3)},
			},
			want: "data.3:",
		},
		{
			name: "set-path with eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
			},
			want: "data.(idx):",
		},
		{
			name: "set-path with nested eval segments",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("field")})},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
			},
			want: "data.(field).(idx):",
		},
		{
			name: "set-path with eval index ordering",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
				{Type: PathSegmentIndex, Value: int64(3)},
			},
			want: "data.(idx).3:",
		},
		{
			name: "set-path with non-block eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: "not-a-block"},
			},
			want: "data.(eval):",
		},
		{
			name:     "empty set-path",
			segments: []PathSegment{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewSetPath(tt.segments, nil)
			got := path.Mold()
			if got != tt.want {
				t.Errorf("SetPathExpression.Mold() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPathExpressionForm(t *testing.T) {
	segments := []PathSegment{
		{Type: PathSegmentWord, Value: "data"},
		{Type: PathSegmentWord, Value: "field"},
	}
	path := NewPath(segments, nil)

	if path.Form() != path.Mold() {
		t.Error("PathExpression.Form() should equal Mold()")
	}
}

func TestPathExpressionString(t *testing.T) {
	tests := []struct {
		name     string
		segments []PathSegment
		want     string
	}{
		{
			name:     "nil path",
			segments: nil,
			want:     "path[]",
		},
		{
			name:     "empty path",
			segments: []PathSegment{},
			want:     "path[]",
		},
		{
			name: "simple word path",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentWord, Value: "field"},
			},
			want: "path[data.field]",
		},
		{
			name: "path with eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewWordVal("idx")})},
			},
			want: "path[data.(idx)]",
		},
		{
			name: "path with eval block containing string literal",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewStrVal("idx")})},
			},
			want: "path[data.(\"idx\")]",
		},
		{
			name: "path with non-block eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: "not-a-block"},
			},
			want: "path[data.(eval)]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewPath(tt.segments, nil)
			got := path.String()
			if got != tt.want {
				t.Errorf("PathExpression.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPathExpressionString(t *testing.T) {
	tests := []struct {
		name     string
		segments []PathSegment
		want     string
	}{
		{
			name:     "nil get-path",
			segments: nil,
			want:     "get-path[]",
		},
		{
			name:     "empty get-path",
			segments: []PathSegment{},
			want:     "get-path[]",
		},
		{
			name: "simple get-path",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentWord, Value: "field"},
			},
			want: "get-path[data.field]",
		},
		{
			name: "get-path with eval block containing string literal",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewStrVal("idx")})},
			},
			want: "get-path[data.(\"idx\")]",
		},
		{
			name: "get-path with non-block eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: "not-a-block"},
			},
			want: "get-path[data.(eval)]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewGetPath(tt.segments, nil)
			got := path.String()
			if got != tt.want {
				t.Errorf("GetPathExpression.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetPathExpressionString(t *testing.T) {
	tests := []struct {
		name     string
		segments []PathSegment
		want     string
	}{
		{
			name:     "nil set-path",
			segments: nil,
			want:     "set-path[]",
		},
		{
			name:     "empty set-path",
			segments: []PathSegment{},
			want:     "set-path[]",
		},
		{
			name: "simple set-path",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentWord, Value: "field"},
			},
			want: "set-path[data.field]",
		},
		{
			name: "set-path with eval block containing string literal",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: NewBlockValue([]core.Value{NewStrVal("idx")})},
			},
			want: "set-path[data.(\"idx\")]",
		},
		{
			name: "set-path with non-block eval segment",
			segments: []PathSegment{
				{Type: PathSegmentWord, Value: "data"},
				{Type: PathSegmentEval, Value: "not-a-block"},
			},
			want: "set-path[data.(eval)]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewSetPath(tt.segments, nil)
			got := path.String()
			if got != tt.want {
				t.Errorf("SetPathExpression.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPathSegmentTypeString(t *testing.T) {
	tests := []struct {
		segType PathSegmentType
		want    string
	}{
		{PathSegmentWord, "word"},
		{PathSegmentIndex, "index"},
		{PathSegmentEval, "eval"},
		{PathSegmentType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.segType.String()
			if got != tt.want {
				t.Errorf("PathSegmentType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEvalSegmentWithoutBlock(t *testing.T) {
	segments := []PathSegment{
		{Type: PathSegmentWord, Value: "data"},
		{Type: PathSegmentEval, Value: "not-a-block"},
	}

	path := NewPath(segments, nil)
	got := path.Mold()
	want := "data.(eval)"

	if got != want {
		t.Errorf("Eval segment without BlockValue Mold() = %q, want %q", got, want)
	}
}
