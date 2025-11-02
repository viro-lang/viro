package value

import "github.com/marcin-radoszewski/viro/internal/core"

// Series represents a sequence of values that supports series operations.
// This interface enables polymorphic series navigation without type switches.
type Series interface {
	GetIndex() int
	SetIndex(int)
	Length() int
	Clone() Series
	ElementAt(int) core.Value
}
