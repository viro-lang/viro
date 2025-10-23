package value

import (
	"fmt"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// BlockValue represents an ordered sequence of values.
// Used for both blocks [...] (deferred evaluation) and parens (...)  (immediate evaluation).
// The distinction is made by the Value wrapper's Type field, not the BlockValue itself.
//
// Design per data-model.md:
// - Blocks evaluate to themselves (return self without evaluating contents)
// - Parens evaluate their contents immediately
// - Both share the same underlying structure
type BlockValue struct {
	Elements []core.Value // ordered value sequence
	Index    int          // current series position (0-based)
}

// NewBlockValue creates a BlockValue with given elements.
func NewBlockValue(elements []core.Value) *BlockValue {
	if elements == nil {
		elements = []core.Value{}
	}
	return &BlockValue{
		Elements: elements,
		Index:    0,
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
		parts[i] = elem.String()
	}
	return strings.Join(parts, " ")
}

// Equals performs deep equality comparison.
func (b *BlockValue) Equals(other *BlockValue) bool {
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

// Series operations (contracts/series.md)

// First returns the first element (error if empty handled by caller).
func (b *BlockValue) First() core.Value {
	return b.Elements[0]
}

// Last returns the last element (error if empty handled by caller).
func (b *BlockValue) Last() core.Value {
	return b.Elements[len(b.Elements)-1]
}

// At returns element at index (bounds checking by caller).
func (b *BlockValue) At(index int) core.Value {
	return b.Elements[index]
}

// Length returns element count.
func (b *BlockValue) Length() int {
	return len(b.Elements)
}

// Append adds a value to the end (in-place mutation).
func (b *BlockValue) Append(val core.Value) {
	b.Elements = append(b.Elements, val)
}

// Insert adds a value at current position (in-place mutation).
func (b *BlockValue) Insert(val core.Value) {
	// Insert at current index, shifting remaining elements right
	b.Elements = append(b.Elements[:b.Index], append([]core.Value{val}, b.Elements[b.Index:]...)...)
}

// Remove removes a specified number of elements from the current position (in-place mutation).
func (b *BlockValue) Remove(count int) {
	if b.Index+count <= len(b.Elements) {
		b.Elements = append(b.Elements[:b.Index], b.Elements[b.Index+count:]...)
	}
}

// Index returns current series position.
func (b *BlockValue) GetIndex() int {
	return b.Index
}

// SetIndex updates current series position (bounds checking by caller).
func (b *BlockValue) SetIndex(idx int) {
	b.Index = idx
}

// Clone creates a shallow copy of the block.
// Elements are shared (not deep cloned).
func (b *BlockValue) Clone() *BlockValue {
	elemsCopy := make([]core.Value, len(b.Elements))
	copy(elemsCopy, b.Elements)
	return &BlockValue{
		Elements: elemsCopy,
		Index:    b.Index,
	}
}

// Empty returns true if block has no elements.
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
			iVal, _ := AsInteger(elemI)
			jVal, _ := AsInteger(elemJ)
			return iVal < jVal
		case TypeString:
			iVal, _ := AsString(elemI)
			jVal, _ := AsString(elemJ)
			return iVal.String() < jVal.String()
		default:
			return false
		}
	})
}
