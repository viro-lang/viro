package value

import (
	"fmt"
	"sort"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type StringValue struct {
	runes []rune
	index int
}

func NewStringValue(s string) *StringValue {
	return &StringValue{
		runes: []rune(s),
		index: 0,
	}
}

func (s *StringValue) String() string {
	return string(s.runes)
}

func (s *StringValue) Mold() string {
	return fmt.Sprintf(`"%s"`, s.String())
}

func (s *StringValue) Form() string {
	return s.String()
}

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

func (s *StringValue) First() rune {
	return s.runes[0]
}

func (s *StringValue) Last() rune {
	return s.runes[len(s.runes)-1]
}

func (s *StringValue) At(index int) rune {
	return s.runes[index]
}

func (s *StringValue) ElementAt(index int) core.Value {
	return NewStrVal(string(s.At(index)))
}

func (s *StringValue) Length() int {
	return len(s.runes)
}

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

	s.runes = append(toInsert, s.runes...)
}

func (s *StringValue) GetIndex() int {
	return s.index
}

func (s *StringValue) Runes() []rune {
	return s.runes
}

func (s *StringValue) SetRunes(r []rune) {
	s.runes = r
}

func (s *StringValue) Clone() Series {
	runesCopy := make([]rune, len(s.runes))
	copy(runesCopy, s.runes)
	return &StringValue{
		runes: runesCopy,
		index: s.index,
	}
}

func (s *StringValue) SetIndex(idx int) {
	s.index = idx
}

func (s *StringValue) Remove(count int) {
	if s.index+count <= len(s.runes) {
		s.runes = append(s.runes[:s.index], s.runes[s.index+count:]...)
	}
}

func SortString(s *StringValue) {
	sort.SliceStable(s.runes, func(i, j int) bool {
		return s.runes[i] < s.runes[j]
	})
}
