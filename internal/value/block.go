package value

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type BlockValue struct {
	Elements  []core.Value
	Index     int
	typ       core.ValueType
	locations []core.SourceLocation
}

func NewBlockValue(elements []core.Value) *BlockValue {
	if elements == nil {
		elements = []core.Value{}
	}
	return &BlockValue{
		Elements:  elements,
		Index:     0,
		typ:       TypeBlock,
		locations: make([]core.SourceLocation, len(elements)),
	}
}

func NewBlockValueWithType(elements []core.Value, typ core.ValueType) *BlockValue {
	if elements == nil {
		elements = []core.Value{}
	}
	return &BlockValue{
		Elements:  elements,
		Index:     0,
		typ:       typ,
		locations: make([]core.SourceLocation, len(elements)),
	}
}

func (b *BlockValue) ensureLocationCapacity() {
	if len(b.locations) != len(b.Elements) {
		newLocations := make([]core.SourceLocation, len(b.Elements))
		copy(newLocations, b.locations)
		b.locations = newLocations
	}
}

func (b *BlockValue) SetLocations(locations []core.SourceLocation) {
	b.locations = make([]core.SourceLocation, len(b.Elements))
	copy(b.locations, locations)
}

func (b *BlockValue) SetLocationAt(index int, location core.SourceLocation) {
	if index < 0 || index >= len(b.Elements) {
		return
	}
	b.ensureLocationCapacity()
	b.locations[index] = location
}

func (b *BlockValue) LocationAt(index int) core.SourceLocation {
	if index < 0 || index >= len(b.locations) {
		return core.SourceLocation{}
	}
	return b.locations[index]
}

func (b *BlockValue) Locations() []core.SourceLocation {
	if len(b.locations) == 0 {
		return nil
	}
	return b.locations
}

func (b *BlockValue) String() string {
	return "[" + b.StringElements() + "]"
}

func (b *BlockValue) Mold() string {
	if len(b.Elements) == 0 {
		return "[]"
	}
	if b.Index >= len(b.Elements) {
		return "[]"
	}
	visibleElements := b.Elements[b.Index:]
	parts := make([]string, len(visibleElements))
	for i, elem := range visibleElements {
		parts[i] = elem.Mold()
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func (b *BlockValue) Form() string {
	if len(b.Elements) == 0 {
		return ""
	}
	if b.Index >= len(b.Elements) {
		return ""
	}
	visibleElements := b.Elements[b.Index:]
	parts := make([]string, len(visibleElements))
	for i, elem := range visibleElements {
		parts[i] = elem.Form()
	}
	return strings.Join(parts, " ")
}

func (b *BlockValue) MoldElements() string {
	if len(b.Elements) == 0 {
		return ""
	}
	parts := make([]string, len(b.Elements))
	for i, elem := range b.Elements {
		parts[i] = elem.Mold()
	}
	return strings.Join(parts, " ")
}

func (b *BlockValue) StringElements() string {
	if len(b.Elements) == 0 {
		return ""
	}
	parts := make([]string, len(b.Elements))
	for i, elem := range b.Elements {
		parts[i] = elem.Form()
	}
	return strings.Join(parts, " ")
}

func (b *BlockValue) EqualsBlock(other *BlockValue) bool {
	if len(b.Elements) != len(other.Elements) {
		return false
	}
	for i := range b.Elements {
		if !b.Elements[i].Equals(other.Elements[i]) {
			return false
		}
	}
	return true
}

func (b *BlockValue) Equals(other core.Value) bool {
	if other.GetType() != TypeBlock && other.GetType() != TypeParen {
		return false
	}
	otherBlock, ok := other.(*BlockValue)
	if !ok {
		return false
	}
	return b.EqualsBlock(otherBlock)
}

func (b *BlockValue) GetType() core.ValueType {
	return b.typ
}

func (b *BlockValue) GetPayload() any {
	return b
}

func (b *BlockValue) First() core.Value {
	return b.Elements[0]
}

func (b *BlockValue) Last() core.Value {
	return b.Elements[len(b.Elements)-1]
}

func (b *BlockValue) At(index int) core.Value {
	return b.Elements[index]
}

func (b *BlockValue) ElementAt(index int) core.Value {
	return b.At(index)
}

func (b *BlockValue) Length() int {
	return len(b.Elements)
}

func (b *BlockValue) Append(val core.Value) {
	b.ensureLocationCapacity()
	b.Elements = append(b.Elements, val)
	b.locations = append(b.locations, core.SourceLocation{})
}

func (b *BlockValue) Insert(val core.Value) {
	b.ensureLocationCapacity()
	b.Elements = append([]core.Value{val}, b.Elements...)
	b.locations = append([]core.SourceLocation{{}}, b.locations...)
}

func (b *BlockValue) Remove(count int) {
	b.ensureLocationCapacity()
	b.Elements = append(b.Elements[:b.Index], b.Elements[b.Index+count:]...)
	if len(b.locations) >= b.Index+count {
		b.locations = append(b.locations[:b.Index], b.locations[b.Index+count:]...)
	} else if len(b.locations) > b.Index {
		b.locations = b.locations[:b.Index]
	}
}

func (b *BlockValue) GetIndex() int {
	return b.Index
}

func (b *BlockValue) SetIndex(idx int) {
	b.Index = idx
}

func (b *BlockValue) Clone() Series {
	elemsCopy := make([]core.Value, len(b.Elements))
	copy(elemsCopy, b.Elements)
	locCopy := make([]core.SourceLocation, len(b.Elements))
	copy(locCopy, b.locations)
	return &BlockValue{
		Elements:  elemsCopy,
		Index:     b.Index,
		typ:       b.typ,
		locations: locCopy,
	}
}

func (b *BlockValue) GoString() string {
	return fmt.Sprintf("Block{Elements: %d, Index: %d}", len(b.Elements), b.Index)
}

func (b *BlockValue) FirstValue() (core.Value, error) {
	if len(b.Elements) == 0 {
		return NewNoneVal(), errors.New("empty series: first element")
	}
	if b.Index >= len(b.Elements) {
		return NewNoneVal(), fmt.Errorf("out of bounds: %d >= %d", b.Index, len(b.Elements))
	}
	return b.Elements[b.Index], nil
}

func (b *BlockValue) LastValue() (core.Value, error) {
	if len(b.Elements) == 0 {
		return NewNoneVal(), errors.New("empty series: last element")
	}
	return b.Last(), nil
}

func (b *BlockValue) AppendValue(val core.Value) error {
	b.Append(val)
	return nil
}

func (b *BlockValue) InsertValue(val core.Value) error {
	b.SetIndex(0)
	b.Insert(val)
	return nil
}

func (b *BlockValue) CopyPart(count int) (Series, error) {
	b.ensureLocationCapacity()
	clampedCount := ClampToRemaining(b.Index, len(b.Elements), count)
	elemsCopy := make([]core.Value, clampedCount)
	copy(elemsCopy, b.Elements[b.Index:b.Index+clampedCount])
	copyBlock := NewBlockValue(elemsCopy)
	if clampedCount > 0 {
		locCopy := make([]core.SourceLocation, clampedCount)
		copy(locCopy, b.locations[b.Index:b.Index+clampedCount])
		copyBlock.SetLocations(locCopy)
	}
	return copyBlock, nil
}

func (b *BlockValue) RemoveCount(count int) error {
	if count < 0 {
		return fmt.Errorf("out of bounds: %d must be non-negative", count)
	}
	if b.Index+count > len(b.Elements) {
		return fmt.Errorf("out of bounds: index %d + count %d > length %d", b.Index, count, len(b.Elements))
	}
	b.Remove(count)
	return nil
}

func (b *BlockValue) SkipBy(count int) {
	newIndex := b.Index + count
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex > len(b.Elements) {
		newIndex = len(b.Elements)
	}
	b.SetIndex(newIndex)
}

func (b *BlockValue) TakeCount(count int) Series {
	b.ensureLocationCapacity()
	if count > b.Length()-b.Index {
		count = b.Length() - b.Index
	}
	end := b.Index + count
	if end > len(b.Elements) {
		end = len(b.Elements)
	}
	elemsCopy := make([]core.Value, count)
	copy(elemsCopy, b.Elements[b.Index:end])
	taken := NewBlockValue(elemsCopy)
	if count > 0 {
		locCopy := make([]core.SourceLocation, count)
		copy(locCopy, b.locations[b.Index:end])
		taken.SetLocations(locCopy)
	}
	return taken
}

func (b *BlockValue) ChangeValue(val core.Value) error {
	if b.Index >= len(b.Elements) {
		return fmt.Errorf("out of bounds: index %d >= length %d", b.Index, len(b.Elements))
	}
	b.Elements[b.Index] = val
	return nil
}

func (b *BlockValue) ClearSeries() {
	b.Elements = []core.Value{}
	b.Index = 0
	b.locations = []core.SourceLocation{}
}

func SortBlock(b *BlockValue) {
	sort.SliceStable(b.Elements, func(i, j int) bool {
		elemI := b.Elements[i]
		elemJ := b.Elements[j]
		switch elemI.GetType() {
		case TypeInteger:
			iVal, _ := AsIntValue(elemI)
			jVal, _ := AsIntValue(elemJ)
			return iVal < jVal
		case TypeString:
			iVal, _ := AsStringValue(elemI)
			jVal, _ := AsStringValue(elemJ)
			return iVal.Form() < jVal.Form()
		default:
			return false
		}
	})
}
