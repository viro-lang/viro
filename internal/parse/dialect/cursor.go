package dialect

import (
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// SeriesCursor provides a unified interface for navigating string and block series.
type SeriesCursor interface {
	// Length returns the total length of the series
	Length() int

	// At returns the element at the given position
	At(pos int) (core.Value, bool)

	// Peek returns the element at the current position without advancing
	Peek(pos int) (core.Value, bool)

	// Slice returns a slice of the series from start to end
	Slice(start, end int) core.Value

	// String returns a string representation of the element at pos
	StringAt(pos int) string

	// IsString returns true if this is a string cursor
	IsString() bool
}

// StringCursor implements SeriesCursor for strings.
type StringCursor struct {
	runes []rune
}

// NewStringCursor creates a cursor for a string value.
func NewStringCursor(s string) *StringCursor {
	return &StringCursor{
		runes: []rune(s),
	}
}

func (sc *StringCursor) Length() int {
	return len(sc.runes)
}

func (sc *StringCursor) At(pos int) (core.Value, bool) {
	if pos < 0 || pos >= len(sc.runes) {
		return value.NewNoneVal(), false
	}
	return value.NewStrVal(string(sc.runes[pos])), true
}

func (sc *StringCursor) Peek(pos int) (core.Value, bool) {
	return sc.At(pos)
}

func (sc *StringCursor) Slice(start, end int) core.Value {
	if start < 0 {
		start = 0
	}
	if end > len(sc.runes) {
		end = len(sc.runes)
	}
	if start >= end {
		return value.NewStrVal("")
	}
	return value.NewStrVal(string(sc.runes[start:end]))
}

func (sc *StringCursor) StringAt(pos int) string {
	if pos < 0 || pos >= len(sc.runes) {
		return ""
	}
	return string(sc.runes[pos])
}

func (sc *StringCursor) IsString() bool {
	return true
}

// RuneAt returns the rune at the given position.
func (sc *StringCursor) RuneAt(pos int) (rune, bool) {
	if pos < 0 || pos >= len(sc.runes) {
		return 0, false
	}
	return sc.runes[pos], true
}

// BlockCursor implements SeriesCursor for blocks.
type BlockCursor struct {
	elements []core.Value
}

// NewBlockCursor creates a cursor for a block value.
func NewBlockCursor(elements []core.Value) *BlockCursor {
	return &BlockCursor{
		elements: elements,
	}
}

func (bc *BlockCursor) Length() int {
	return len(bc.elements)
}

func (bc *BlockCursor) At(pos int) (core.Value, bool) {
	if pos < 0 || pos >= len(bc.elements) {
		return value.NewNoneVal(), false
	}
	return bc.elements[pos], true
}

func (bc *BlockCursor) Peek(pos int) (core.Value, bool) {
	return bc.At(pos)
}

func (bc *BlockCursor) Slice(start, end int) core.Value {
	if start < 0 {
		start = 0
	}
	if end > len(bc.elements) {
		end = len(bc.elements)
	}
	if start >= end {
		return value.NewBlockVal([]core.Value{})
	}
	return value.NewBlockVal(bc.elements[start:end])
}

func (bc *BlockCursor) StringAt(pos int) string {
	if pos < 0 || pos >= len(bc.elements) {
		return ""
	}
	return bc.elements[pos].Form()
}

func (bc *BlockCursor) IsString() bool {
	return false
}

// NewCursor creates an appropriate cursor for the given value.
func NewCursor(v core.Value) (SeriesCursor, bool) {
	switch v.GetType() {
	case value.TypeString:
		if strVal, ok := value.AsStringValue(v); ok {
			return NewStringCursor(strVal.String()), true
		}
	case value.TypeBlock, value.TypeParen:
		if blockVal, ok := value.AsBlockValue(v); ok {
			return NewBlockCursor(blockVal.Elements), true
		}
	}
	return nil, false
}

// MatchString performs case-sensitive or case-insensitive string matching.
func MatchString(s1, s2 string, caseSensitive bool) bool {
	if caseSensitive {
		return s1 == s2
	}
	return strings.EqualFold(s1, s2)
}
