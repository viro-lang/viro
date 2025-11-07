package value

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type PathExpression struct {
	Segments []PathSegment
	Base     core.Value
}

// PathSegment represents a single step in a path traversal.
type PathSegment struct {
	Type  PathSegmentType
	Value any
}

// PathSegmentType identifies the kind of path segment.
type PathSegmentType int

const (
	PathSegmentWord       PathSegmentType = iota
	PathSegmentIndex
	PathSegmentRefinement
	PathSegmentEval
)

func (t PathSegmentType) String() string {
	switch t {
	case PathSegmentWord:
		return "word"
	case PathSegmentIndex:
		return "index"
	case PathSegmentRefinement:
		return "refinement"
	case PathSegmentEval:
		return "eval"
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

func renderPathSegments(segments []PathSegment, prefix, suffix string) string {
	result := prefix
	for i, seg := range segments {
		if i > 0 {
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
		case PathSegmentEval:
			if block, ok := seg.Value.(*BlockValue); ok {
				result += "(" + block.MoldElements() + ")"
			} else {
				result += "(eval)"
			}
		}
	}
	result += suffix
	return result
}

func (p *PathExpression) String() string {
	if p == nil || len(p.Segments) == 0 {
		return "path[]"
	}
	result := "path["
	for i, seg := range p.Segments {
		if i > 0 && seg.Type != PathSegmentRefinement && seg.Type != PathSegmentEval {
			result += "."
		}
		switch seg.Type {
		case PathSegmentWord:
			result += seg.Value.(string)
		case PathSegmentIndex:
			result += fmt.Sprintf("%d", seg.Value.(int64))
		case PathSegmentRefinement:
			result += "/" + seg.Value.(string)
		case PathSegmentEval:
			evalPrefix := ""
			if i > 0 {
				evalPrefix = "."
			}
			if block, ok := seg.Value.(*BlockValue); ok {
				result += evalPrefix + "(" + block.Form() + ")"
			} else {
				result += evalPrefix + "(eval)"
			}
		}
	}
	result += "]"
	return result
}

func (p *PathExpression) Mold() string {
	if p == nil || len(p.Segments) == 0 {
		return ""
	}
	return renderPathSegments(p.Segments, "", "")
}

func (p *PathExpression) Form() string {
	return p.Mold()
}

func PathVal(path *PathExpression) core.Value {
	return path
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

// GetPathExpression marks a path as non-invoking (like get-words)
type GetPathExpression struct {
	*PathExpression
}

// NewGetPath creates a GetPathExpression with the given segments and base value.
func NewGetPath(segments []PathSegment, base core.Value) *GetPathExpression {
	return &GetPathExpression{
		PathExpression: NewPath(segments, base),
	}
}

// String returns a get-path-like representation for debugging.
func (g *GetPathExpression) String() string {
	if g == nil || len(g.Segments) == 0 {
		return "get-path[]"
	}
	result := "get-path["
	for i, seg := range g.Segments {
		if i > 0 && seg.Type != PathSegmentRefinement && seg.Type != PathSegmentEval {
			result += "."
		}
		switch seg.Type {
		case PathSegmentWord:
			result += seg.Value.(string)
		case PathSegmentIndex:
			result += fmt.Sprintf("%d", seg.Value.(int64))
		case PathSegmentRefinement:
			result += "/" + seg.Value.(string)
		case PathSegmentEval:
			evalPrefix := ""
			if i > 0 {
				evalPrefix = "."
			}
			if block, ok := seg.Value.(*BlockValue); ok {
				result += evalPrefix + "(" + block.Form() + ")"
			} else {
				result += evalPrefix + "(eval)"
			}
		}
	}
	result += "]"
	return result
}

func (g *GetPathExpression) Mold() string {
	if g == nil || len(g.Segments) == 0 {
		return ""
	}
	return renderPathSegments(g.Segments, ":", "")
}

func (g *GetPathExpression) Form() string {
	return g.Mold()
}

func GetPathVal(path *GetPathExpression) core.Value {
	return path
}

// AsGetPath extracts the GetPathExpression from a Value, or returns nil if wrong type.
func AsGetPath(v core.Value) (*GetPathExpression, bool) {
	if v.GetType() != TypeGetPath {
		return nil, false
	}
	path, ok := v.GetPayload().(*GetPathExpression)
	return path, ok
}

func (g *GetPathExpression) GetType() core.ValueType {
	return TypeGetPath
}

func (g *GetPathExpression) GetPayload() any {
	return g
}

func (g *GetPathExpression) Equals(other core.Value) bool {
	if other.GetType() != TypeGetPath {
		return false
	}
	return other.GetPayload() == g
}

// SetPathExpression marks a path as assignment target (like set-words)
type SetPathExpression struct {
	*PathExpression
}

// NewSetPath creates a SetPathExpression with the given segments and base value.
func NewSetPath(segments []PathSegment, base core.Value) *SetPathExpression {
	return &SetPathExpression{
		PathExpression: NewPath(segments, base),
	}
}

// String returns a set-path-like representation for debugging.
func (s *SetPathExpression) String() string {
	if s == nil || len(s.Segments) == 0 {
		return "set-path[]"
	}
	result := "set-path["
	for i, seg := range s.Segments {
		if i > 0 && seg.Type != PathSegmentRefinement && seg.Type != PathSegmentEval {
			result += "."
		}
		switch seg.Type {
		case PathSegmentWord:
			result += seg.Value.(string)
		case PathSegmentIndex:
			result += fmt.Sprintf("%d", seg.Value.(int64))
		case PathSegmentRefinement:
			result += "/" + seg.Value.(string)
		case PathSegmentEval:
			evalPrefix := ""
			if i > 0 {
				evalPrefix = "."
			}
			if block, ok := seg.Value.(*BlockValue); ok {
				result += evalPrefix + "(" + block.Form() + ")"
			} else {
				result += evalPrefix + "(eval)"
			}
		}
	}
	result += "]"
	return result
}

func (s *SetPathExpression) Mold() string {
	if s == nil || len(s.Segments) == 0 {
		return ""
	}
	return renderPathSegments(s.Segments, "", ":")
}

func (s *SetPathExpression) Form() string {
	return s.Mold()
}

func SetPathVal(path *SetPathExpression) core.Value {
	return path
}

// AsSetPath extracts the SetPathExpression from a Value, or returns nil if wrong type.
func AsSetPath(v core.Value) (*SetPathExpression, bool) {
	if v.GetType() != TypeSetPath {
		return nil, false
	}
	path, ok := v.GetPayload().(*SetPathExpression)
	return path, ok
}

func (s *SetPathExpression) GetType() core.ValueType {
	return TypeSetPath
}

func (s *SetPathExpression) GetPayload() any {
	return s
}

func (s *SetPathExpression) Equals(other core.Value) bool {
	if other.GetType() != TypeSetPath {
		return false
	}
	return other.GetPayload() == s
}
