package value

import "sort"

// StringValue represents a UTF-8 character sequence.
// Stored as []rune internally for proper REBOL series semantics
// (character-level operations, not byte-level).
//
// Design decision per research.md:
// - REBOL strings are character series, not byte series
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

// Equals performs deep equality comparison.
func (s *StringValue) Equals(other *StringValue) bool {
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
