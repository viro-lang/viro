package value

import (
	"errors"
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

func (s *StringValue) FirstValue() (core.Value, error) {
	if len(s.runes) == 0 {
		return NewNoneVal(), errors.New("empty series: first element")
	}
	if s.index >= len(s.runes) {
		return NewNoneVal(), fmt.Errorf("out of bounds: %d >= %d", s.index, len(s.runes))
	}
	return NewStrVal(string(s.runes[s.index])), nil
}

func (s *StringValue) LastValue() (core.Value, error) {
	if len(s.runes) == 0 {
		return NewNoneVal(), errors.New("empty series: last element")
	}
	return NewStrVal(string(s.Last())), nil
}

func (s *StringValue) AppendValue(val core.Value) error {
	switch val.GetType() {
	case TypeString:
		strVal, _ := AsStringValue(val)
		s.Append(strVal)
	default:
		return fmt.Errorf("type mismatch: expected string, got %s", TypeToString(val.GetType()))
	}
	return nil
}

func (s *StringValue) InsertValue(val core.Value) error {
	switch val.GetType() {
	case TypeString:
		strVal, _ := AsStringValue(val)
		s.SetIndex(0)
		s.Insert(strVal)
	default:
		return fmt.Errorf("type mismatch: expected string, got %s", TypeToString(val.GetType()))
	}
	return nil
}

func (s *StringValue) CopyPart(count int) (Series, error) {
	if count < 0 {
		return nil, fmt.Errorf("out of bounds: count %d < 0", count)
	}
	remaining := len(s.runes) - s.index
	if count > remaining {
		count = remaining
	}
	runesCopy := make([]rune, count)
	copy(runesCopy, s.runes[s.index:s.index+count])
	return NewStringValue(string(runesCopy)), nil
}

func (s *StringValue) RemoveCount(count int) error {
	if count < 0 {
		return fmt.Errorf("out of bounds: %d must be non-negative", count)
	}
	if s.index+count > len(s.runes) {
		return fmt.Errorf("out of bounds: index %d + count %d > length %d", s.index, count, len(s.runes))
	}
	s.Remove(count)
	return nil
}

func (s *StringValue) SkipBy(count int) {
	newIndex := s.index + count
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex > len(s.runes) {
		newIndex = len(s.runes)
	}
	s.SetIndex(newIndex)
}

func (s *StringValue) TakeCount(count int) Series {
	if count > s.Length()-s.index {
		count = s.Length() - s.index
	}
	end := s.index + count
	if end > len(s.runes) {
		end = len(s.runes)
	}
	runesCopy := make([]rune, count)
	copy(runesCopy, s.runes[s.index:end])
	return NewStringValue(string(runesCopy))
}

func (s *StringValue) ChangeValue(val core.Value) error {
	switch val.GetType() {
	case TypeString:
		strVal, _ := AsStringValue(val)
		runes := strVal.Runes()
		if len(runes) != 1 {
			return fmt.Errorf("type mismatch: expected single character string, got string of length %d", len(runes))
		}
		if s.index >= len(s.runes) {
			return fmt.Errorf("out of bounds: index %d >= length %d", s.index, len(s.runes))
		}
		s.runes[s.index] = runes[0]
	default:
		return fmt.Errorf("type mismatch: expected string, got %s", TypeToString(val.GetType()))
	}
	return nil
}

func (s *StringValue) ClearSeries() {
	s.runes = []rune{}
	s.index = 0
}

func SortString(s *StringValue) {
	sort.SliceStable(s.runes, func(i, j int) bool {
		return s.runes[i] < s.runes[j]
	})
}
