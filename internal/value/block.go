package value

import (
	"fmt"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// BlockValue represents an ordered sequence of values.
// Used for both blocks [...] (deferred evaluation) and parens (...)  (immediate evaluation).
// The distinction is made by the typ field.
//
// Design per data-model.md:
// - Blocks evaluate to themselves (return self without evaluating contents)
// - Parens evaluate their contents immediately
// - Both share the same underlying structure
type BlockValue struct {
	Elements []core.Value   // ordered value sequence
	Index    int            // current series position (0-based)
	typ      core.ValueType // TypeBlock or TypeParen
}

// NewBlockValue creates a BlockValue with given elements.
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

// String returns a string representation with block syntax.
func (b *BlockValue) String() string {
	return "[" + b.StringElements() + "]"
}

// Mold returns the mold-formatted block representation (with brackets).
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

// Form returns the form-formatted block representation (without brackets).
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

// MoldElements returns space-separated molded element representations.
// Used by Paren mold formatting.
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

// StringElements returns space-separated element representations.
// Used by both Block and Paren string formatting.
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

// EqualsBlock performs deep equality comparison with another BlockValue.
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

// Series operations (contracts/series.md)

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
	// Insert at current index, shifting remaining elements right
	b.Elements = append(b.Elements[:b.Index], append([]core.Value{val}, b.Elements[b.Index:]...)...)
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

func (b *BlockValue) TailValue() core.Value {
	if len(b.Elements) == 0 {
		return NewBlockVal([]core.Value{})
	}
	return NewBlockVal(append([]core.Value{}, b.Elements[1:]...))
}

func (b *BlockValue) Empty() bool {
	return len(b.Elements) == 0
}

// For debugging/testing
func (b *BlockValue) GoString() string {
	return fmt.Sprintf("Block{Elements: %d, Index: %d}", len(b.Elements), b.Index)
}

// SortBlock sorts a block in-place.
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
