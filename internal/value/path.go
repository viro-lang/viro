package value

import "fmt"

// PathExpression represents a path during evaluation (Feature 002).
//
// Design per data-model.md:
// - Segments: sequence of steps (word, index, refinement) for traversal
// - Base: starting value for path evaluation
//
// Per FR-010: evaluates across objects, blocks, and future maps using dot notation
// Note: TypePath is transient and should not persist outside evaluation context
type PathExpression struct {
	Segments []PathSegment // Path components (e.g., "user", "address", "city")
	Base     Value         // Starting value for traversal
}

// PathSegment represents a single step in a path traversal.
type PathSegment struct {
	Type  PathSegmentType // word, index, or refinement
	Value interface{}     // string (word/refinement) or int64 (index)
}

// PathSegmentType identifies the kind of path segment.
type PathSegmentType int

const (
	PathSegmentWord       PathSegmentType = iota // Field access (object.field)
	PathSegmentIndex                             // Series indexing (block.3)
	PathSegmentRefinement                        // Function refinement (func/ref)
)

func (t PathSegmentType) String() string {
	switch t {
	case PathSegmentWord:
		return "word"
	case PathSegmentIndex:
		return "index"
	case PathSegmentRefinement:
		return "refinement"
	default:
		return "unknown"
	}
}

// NewPath creates a PathExpression with the given segments and base value.
func NewPath(segments []PathSegment, base Value) *PathExpression {
	return &PathExpression{
		Segments: segments,
		Base:     base,
	}
}

// String returns a path-like representation for debugging.
func (p *PathExpression) String() string {
	if p == nil || len(p.Segments) == 0 {
		return "path[]"
	}
	result := "path["
	for i, seg := range p.Segments {
		if i > 0 {
			result += "."
		}
		switch seg.Type {
		case PathSegmentWord:
			result += seg.Value.(string)
		case PathSegmentIndex:
			result += fmt.Sprintf("%d", seg.Value.(int64))
		case PathSegmentRefinement:
			result += "/" + seg.Value.(string)
		}
	}
	result += "]"
	return result
}

// PathVal creates a Value wrapping a PathExpression.
func PathVal(path *PathExpression) Value {
	return Value{
		Type:    TypePath,
		Payload: path,
	}
}

// AsPath extracts the PathExpression from a Value, or returns nil if wrong type.
func (v Value) AsPath() (*PathExpression, bool) {
	if v.Type != TypePath {
		return nil, false
	}
	path, ok := v.Payload.(*PathExpression)
	return path, ok
}
