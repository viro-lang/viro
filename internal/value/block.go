package value

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type BlockValue struct {
	Elements []core.Value
	Index    int
	typ      core.ValueType
}

func NewBlockValue(elements []core.Value) *BlockValue {
	if elements == nil {
		elements = []core.Value{}
	}
	return &BlockValue{
		Elements: elements,
		Index:    0,
		typ:      TypeBlock,
	}
}

func NewBlockValueWithType(elements []core.Value, typ core.ValueType) *BlockValue {
	if elements == nil {
		elements = []core.Value{}
	}
	return &BlockValue{
		Elements: elements,
		Index:    0,
		typ:      typ,
	}
}

func (b *BlockValue) String() string {
	return "[" + b.StringElements() + "]"
}

func (b *BlockValue) Mold() string {
	if len(b.Elements) == 0 {
		return "[]"
	}
	parts := make([]string, len(b.Elements))
	for i, elem := range b.Elements {
		parts[i] = elem.Mold()
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func (b *BlockValue) Form() string {
	if len(b.Elements) == 0 {
		return ""
	}
	parts := make([]string, len(b.Elements))
	for i, elem := range b.Elements {
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
	b.Elements = append(b.Elements, val)
}

func (b *BlockValue) Insert(val core.Value) {
	b.Elements = append([]core.Value{val}, b.Elements...)
}

func (b *BlockValue) Remove(count int) {
	if b.Index+count <= len(b.Elements) {
		b.Elements = append(b.Elements[:b.Index], b.Elements[b.Index+count:]...)
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
	return &BlockValue{
		Elements: elemsCopy,
		Index:    b.Index,
		typ:      b.typ,
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
	if count < 0 {
		return nil, fmt.Errorf("out of bounds: count %d < 0", count)
	}
	remaining := len(b.Elements) - b.Index
	if count > remaining {
		count = remaining
	}
	elemsCopy := make([]core.Value, count)
	copy(elemsCopy, b.Elements[b.Index:b.Index+count])
	return NewBlockValue(elemsCopy), nil
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
	if count > b.Length()-b.Index {
		count = b.Length() - b.Index
	}
	end := b.Index + count
	if end > len(b.Elements) {
		end = len(b.Elements)
	}
	elemsCopy := make([]core.Value, count)
	copy(elemsCopy, b.Elements[b.Index:end])
	return NewBlockValue(elemsCopy)
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
