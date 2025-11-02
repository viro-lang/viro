package value

import "github.com/marcin-radoszewski/viro/internal/core"

// Series represents a sequence of values that supports series operations.
// This interface enables polymorphic series navigation without type switches.
type Series interface {
	core.Value
	GetIndex() int
	SetIndex(int)
	Length() int
	Clone() Series
	ElementAt(int) core.Value
	FirstValue() (core.Value, error)
	LastValue() (core.Value, error)
	AppendValue(core.Value) error
	InsertValue(core.Value) error
	CopyPart(count int) (Series, error)
	RemoveCount(count int) error
	SkipBy(count int)
	TakeCount(count int) Series
	ChangeValue(core.Value) error
	ClearSeries()
}
