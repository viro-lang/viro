package value

import (
	"fmt"
	"sort"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// StringValue represents a UTF-8 character sequence.
// Stored as []rune internally for proper character series semantics
// (character-level operations, not byte-level).
//
// Design decision per research.md:
// - strings are character series, not byte series
// - first "hello" → 'h' (character), not byte 104
// - Multi-byte Unicode handled correctly
type StringValue struct {
	runes []rune // character sequence
	index int    // current position for series operations
}

// NewStringValue creates a StringValue from a Go string.
func NewStringValue(s string) *StringValue {
	return &StringValue{
		runes: []rune(s),
		index: 0,
	}
}

// String converts StringValue to Go string for display and I/O.
func (s *StringValue) String() string {
	return string(s.runes)
}

// Mold returns the mold-formatted string representation (with quotes).
func (s *StringValue) Mold() string {
	return fmt.Sprintf(`"%s"`, s.String())
}

// Form returns the form-formatted string representation (without quotes).
func (s *StringValue) Form() string {
	return s.String()
}

// EqualsString performs deep equality comparison with another StringValue.
func (s *StringValue) EqualsString(other *StringValue) bool {
	if len(s.runes) != len(other.runes) {
		return false
	}
	for i := range s.runes {
		if s.runes[i] != other.runes[i] {
			return false
		}
	}
	return true
}

func (s *StringValue) Equals(other core.Value) bool {
	if other.GetType() != TypeString {
		return false
	}
	otherStr, ok := other.(*StringValue)
	if !ok {
		return false
	}
	return s.EqualsString(otherStr)
}

func (s *StringValue) GetType() core.ValueType {
	return TypeString
}

func (s *StringValue) GetPayload() any {
	return s
}

// Series operations (contracts/series.md)

// First returns the first character (error if empty handled by caller).
func (s *StringValue) First() rune {
	return s.runes[0]
}

// Last returns the last character (error if empty handled by caller).
func (s *StringValue) Last() rune {
	return s.runes[len(s.runes)-1]
}

// At returns character at index (bounds checking by caller).
func (s *StringValue) At(index int) rune {
	return s.runes[index]
}

// Length returns character count.
func (s *StringValue) Length() int {
	return len(s.runes)
}

// Append adds a character or string to the end (in-place mutation).
func (s *StringValue) Append(val interface{}) {
	switch v := val.(type) {
	case rune:
		s.runes = append(s.runes, v)
	case *StringValue:
		s.runes = append(s.runes, v.runes...)
	case string:
		s.runes = append(s.runes, []rune(v)...)
	}
}

// Insert adds a character or string at current position (in-place mutation).
func (s *StringValue) Insert(val interface{}) {
	var toInsert []rune
	switch v := val.(type) {
	case rune:
		toInsert = []rune{v}
	case *StringValue:
		toInsert = v.runes
	case string:
		toInsert = []rune(v)
	}

	// Insert at current index
	s.runes = append(s.runes[:s.index], append(toInsert, s.runes[s.index:]...)...)
}

// Index returns current series position.
func (s *StringValue) Index() int {
	return s.index
}

// Runes returns the underlying rune slice of the string.
func (s *StringValue) Runes() []rune {
	return s.runes
}

// SetRunes sets the underlying rune slice of the string.
func (s *StringValue) SetRunes(r []rune) {
	s.runes = r
}

// Clone creates a shallow copy of the string.
// Runes are shared (not deep cloned).
func (s *StringValue) Clone() *StringValue {
	runesCopy := make([]rune, len(s.runes))
	copy(runesCopy, s.runes)
	return &StringValue{
		runes: runesCopy,
		index: s.index,
	}
}

// SetIndex updates current series position (bounds checking by caller).

func (s *StringValue) SetIndex(idx int) {
	s.index = idx
}

// Remove removes a specified number of characters from the current position (in-place mutation).

func (s *StringValue) Remove(count int) {
	if s.index+count <= len(s.runes) {
		s.runes = append(s.runes[:s.index], s.runes[s.index+count:]...)
	}
}

// SortString sorts the runes in the string in ascending order.
func SortString(s *StringValue) {
	sort.SliceStable(s.runes, func(i, j int) bool {
		return s.runes[i] < s.runes[j]
	})
}
