package value

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type PathExpression struct {
	Segments []PathSegment
	Base     core.Value
}

type PathSegment struct {
	Type  PathSegmentType
	Value any
}

type PathSegmentType int

const (
	PathSegmentWord PathSegmentType = iota
	PathSegmentIndex
	PathSegmentEval
)

func (t PathSegmentType) String() string {
	switch t {
	case PathSegmentWord:
		return "word"
	case PathSegmentIndex:
		return "index"
	case PathSegmentEval:
		return "eval"
	default:
		return "unknown"
	}
}

func (seg PathSegment) IsWord() bool {
	return seg.Type == PathSegmentWord
}

func (seg PathSegment) AsWord() (string, bool) {
	if !seg.IsWord() {
		return "", false
	}
	str, ok := seg.Value.(string)
	return str, ok
}

func (seg PathSegment) IsIndex() bool {
	return seg.Type == PathSegmentIndex
}

func (seg PathSegment) AsIndex() (int64, bool) {
	if !seg.IsIndex() {
		return 0, false
	}
	num, ok := seg.Value.(int64)
	return num, ok
}

func (seg PathSegment) IsEval() bool {
	return seg.Type == PathSegmentEval
}

func (seg PathSegment) AsEvalBlock() (*BlockValue, bool) {
	if !seg.IsEval() {
		return nil, false
	}
	block, ok := seg.Value.(*BlockValue)
	if !ok || block == nil {
		return nil, false
	}
	return block, true
}

func NewWordSegment(word string) PathSegment {
	return PathSegment{Type: PathSegmentWord, Value: word}
}

func NewIndexSegment(index int64) PathSegment {
	return PathSegment{Type: PathSegmentIndex, Value: index}
}

func NewEvalSegment(block *BlockValue) PathSegment {
	return PathSegment{Type: PathSegmentEval, Value: block}
}

func NewPath(segments []PathSegment, base core.Value) *PathExpression {
	return &PathExpression{
		Segments: segments,
		Base:     base,
	}
}

func renderPathSegments(segments []PathSegment, prefix, suffix string) string {
	result := prefix
	for i, seg := range segments {
		if i > 0 || (i == 0 && seg.Type == PathSegmentEval) {
			result += "."
		}
		switch seg.Type {
		case PathSegmentWord:
			if word, ok := seg.AsWord(); ok {
				result += word
			} else {
				result += ""
			}
		case PathSegmentIndex:
			if index, ok := seg.AsIndex(); ok {
				result += fmt.Sprintf("%d", index)
			} else {
				result += "<invalid-index>"
			}
		case PathSegmentEval:
			if block, ok := seg.AsEvalBlock(); ok {
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
	return renderPathSegments(p.Segments, "path[", "]")
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

type GetPathExpression struct {
	*PathExpression
}

func NewGetPath(segments []PathSegment, base core.Value) *GetPathExpression {
	return &GetPathExpression{
		PathExpression: NewPath(segments, base),
	}
}

func (g *GetPathExpression) String() string {
	if g == nil || len(g.Segments) == 0 {
		return "get-path[]"
	}
	return renderPathSegments(g.Segments, "get-path[", "]")
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

type SetPathExpression struct {
	*PathExpression
}

func NewSetPath(segments []PathSegment, base core.Value) *SetPathExpression {
	return &SetPathExpression{
		PathExpression: NewPath(segments, base),
	}
}

func (s *SetPathExpression) String() string {
	if s == nil || len(s.Segments) == 0 {
		return "set-path[]"
	}
	return renderPathSegments(s.Segments, "set-path[", "]")
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
