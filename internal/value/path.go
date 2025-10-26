package value

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
)

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
	Base     core.Value    // Starting value for traversal
}

// PathSegment represents a single step in a path traversal.
type PathSegment struct {
	Type  PathSegmentType // word, index, or refinement
	Value any             // string (word/refinement) or int64 (index)
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
func NewPath(segments []PathSegment, base core.Value) *PathExpression {
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

// Mold returns the mold-formatted path representation.
func (p *PathExpression) Mold() string {
	if p == nil || len(p.Segments) == 0 {
		return ""
	}
	result := ""
	for i, seg := range p.Segments {
		if i > 0 {
			// Use appropriate separator based on segment type
			switch seg.Type {
			case PathSegmentRefinement:
				result += "/"
			default:
				result += "."
			}
		}
		switch seg.Type {
		case PathSegmentWord:
			result += seg.Value.(string)
		case PathSegmentIndex:
			result += fmt.Sprintf("%d", seg.Value.(int64))
		case PathSegmentRefinement:
			result += seg.Value.(string)
		}
	}
	return result
}

// Form returns the form-formatted path representation (same as mold for paths).
func (p *PathExpression) Form() string {
	return p.Mold()
}

// PathVal creates a Value wrapping a PathExpression.
func PathVal(path *PathExpression) Value {
	return Value{
		Type:    TypePath,
		Payload: path,
	}
}

// AsPath extracts the PathExpression from a Value, or returns nil if wrong type.
func AsPath(v core.Value) (*PathExpression, bool) {
	if v.GetType() != TypePath {
		return nil, false
	}
	path, ok := v.GetPayload().(*PathExpression)
	return path, ok
}

func (p *PathExpression) GetType() core.ValueType {
	return TypePath
}

func (p *PathExpression) GetPayload() any {
	return p
}

func (p *PathExpression) Equals(other core.Value) bool {
	if other.GetType() != TypePath {
		return false
	}
	return other.GetPayload() == p
}
