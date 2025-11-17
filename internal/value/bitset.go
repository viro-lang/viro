package value

import (
	"fmt"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// BitsetValue represents a character set (charset) for parse dialect.
// It uses a bitmap for efficient character membership testing.
// The bitmap covers the full Unicode range using a map for characters >= 256.
type BitsetValue struct {
	// Bitmap for ASCII/Latin-1 characters (0-255)
	lowBits [256 / 64]uint64
	// Map for Unicode characters >= 256
	highChars map[rune]bool
	// Cached string representation
	cachedMold string
	molded     bool
}

// NewBitsetValue creates an empty bitset.
func NewBitsetValue() *BitsetValue {
	return &BitsetValue{
		highChars: make(map[rune]bool),
	}
}

// NewBitsetFromString creates a bitset containing all characters from the string.
func NewBitsetFromString(s string) *BitsetValue {
	bs := NewBitsetValue()
	for _, r := range s {
		bs.Set(r)
	}
	return bs
}

// NewBitsetFromRange creates a bitset containing all characters in the range [start, end].
func NewBitsetFromRange(start, end rune) *BitsetValue {
	bs := NewBitsetValue()
	for r := start; r <= end; r++ {
		bs.Set(r)
	}
	return bs
}

// Set adds a character to the bitset.
func (b *BitsetValue) Set(r rune) {
	b.molded = false
	if r < 256 {
		idx := r / 64
		bit := r % 64
		b.lowBits[idx] |= (1 << bit)
	} else {
		b.highChars[r] = true
	}
}

// Clear removes a character from the bitset.
func (b *BitsetValue) Clear(r rune) {
	b.molded = false
	if r < 256 {
		idx := r / 64
		bit := r % 64
		b.lowBits[idx] &^= (1 << bit)
	} else {
		delete(b.highChars, r)
	}
}

// Test checks if a character is in the bitset.
func (b *BitsetValue) Test(r rune) bool {
	if r < 256 {
		idx := r / 64
		bit := r % 64
		return (b.lowBits[idx] & (1 << bit)) != 0
	}
	return b.highChars[r]
}

// Clone creates a copy of the bitset.
func (b *BitsetValue) Clone() *BitsetValue {
	newBS := NewBitsetValue()
	newBS.lowBits = b.lowBits
	for r := range b.highChars {
		newBS.highChars[r] = true
	}
	return newBS
}

// Union returns a new bitset containing characters from both bitsets.
func (b *BitsetValue) Union(other *BitsetValue) *BitsetValue {
	result := b.Clone()
	for i := range other.lowBits {
		result.lowBits[i] |= other.lowBits[i]
	}
	for r := range other.highChars {
		result.highChars[r] = true
	}
	return result
}

// Intersect returns a new bitset containing characters present in both bitsets.
func (b *BitsetValue) Intersect(other *BitsetValue) *BitsetValue {
	result := NewBitsetValue()
	for i := range b.lowBits {
		result.lowBits[i] = b.lowBits[i] & other.lowBits[i]
	}
	for r := range b.highChars {
		if other.highChars[r] {
			result.highChars[r] = true
		}
	}
	return result
}

// Complement returns a new bitset containing all ASCII characters NOT in this bitset.
func (b *BitsetValue) Complement() *BitsetValue {
	result := NewBitsetValue()
	for i := range b.lowBits {
		result.lowBits[i] = ^b.lowBits[i]
	}
	return result
}

// IsEmpty returns true if the bitset contains no characters.
func (b *BitsetValue) IsEmpty() bool {
	for _, bits := range b.lowBits {
		if bits != 0 {
			return false
		}
	}
	return len(b.highChars) == 0
}

// Count returns the number of characters in the bitset.
func (b *BitsetValue) Count() int {
	count := 0
	for _, bits := range b.lowBits {
		// Count set bits using Brian Kernighan's algorithm
		for bits != 0 {
			bits &= bits - 1
			count++
		}
	}
	return count + len(b.highChars)
}

// GetChars returns all characters in the bitset as a sorted slice.
func (b *BitsetValue) GetChars() []rune {
	chars := make([]rune, 0, b.Count())
	for i := 0; i < 256; i++ {
		if b.Test(rune(i)) {
			chars = append(chars, rune(i))
		}
	}
	for r := range b.highChars {
		chars = append(chars, r)
	}
	sort.Slice(chars, func(i, j int) bool {
		return chars[i] < chars[j]
	})
	return chars
}

// GetType implements core.Value.
func (b *BitsetValue) GetType() core.ValueType {
	return TypeBitset
}

// GetPayload implements core.Value.
func (b *BitsetValue) GetPayload() any {
	return b
}

// String implements core.Value.
func (b *BitsetValue) String() string {
	return b.Mold()
}

// Mold implements core.Value - returns a code-readable representation.
func (b *BitsetValue) Mold() string {
	if b.molded {
		return b.cachedMold
	}

	chars := b.GetChars()
	if len(chars) == 0 {
		b.cachedMold = "charset []"
		b.molded = true
		return b.cachedMold
	}

	// Try to compress ranges
	var parts []string
	i := 0
	for i < len(chars) {
		start := chars[i]
		end := start

		// Find consecutive characters
		for i+1 < len(chars) && chars[i+1] == chars[i]+1 {
			i++
			end = chars[i]
		}

		if end-start >= 2 {
			// Use range notation for 3+ consecutive chars
			parts = append(parts, fmt.Sprintf("[#\"%c\" - #\"%c\"]", start, end))
		} else {
			// Individual characters
			for r := start; r <= end; r++ {
				parts = append(parts, fmt.Sprintf("#\"%c\"", r))
			}
		}
		i++
	}

	b.cachedMold = fmt.Sprintf("charset [%s]", strings.Join(parts, " "))
	b.molded = true
	return b.cachedMold
}

// Form implements core.Value - returns a human-readable representation.
func (b *BitsetValue) Form() string {
	count := b.Count()
	if count == 0 {
		return "charset []"
	}
	if count <= 10 {
		return b.Mold()
	}
	// For large charsets, show count instead of all characters
	return fmt.Sprintf("charset [... %d characters ...]", count)
}

// Equals implements core.Value.
func (b *BitsetValue) Equals(other core.Value) bool {
	otherBS, ok := AsBitsetValue(other)
	if !ok {
		return false
	}

	// Compare low bits
	if b.lowBits != otherBS.lowBits {
		return false
	}

	// Compare high chars
	if len(b.highChars) != len(otherBS.highChars) {
		return false
	}
	for r := range b.highChars {
		if !otherBS.highChars[r] {
			return false
		}
	}

	return true
}

// Helper functions for type conversion

// NewBitsetVal creates a bitset value.
func NewBitsetVal(chars string) core.Value {
	return NewBitsetFromString(chars)
}

// AsBitsetValue checks if a value is a bitset and returns it.
func AsBitsetValue(v core.Value) (*BitsetValue, bool) {
	if v.GetType() != TypeBitset {
		return nil, false
	}
	bs, ok := v.(*BitsetValue)
	return bs, ok
}
