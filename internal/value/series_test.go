package value

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
)

func TestBinaryValue_SeriesInterface(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *BinaryValue
		testFunc func(t *testing.T, b *BinaryValue)
	}{
		{
			name: "FirstValue",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				val, err := b.FirstValue()
				if err != nil {
					t.Errorf("FirstValue() error = %v", err)
					return
				}
				if val.GetType() != TypeInteger {
					t.Errorf("FirstValue() type = %v, want %v", val.GetType(), TypeInteger)
				}
				intVal, _ := AsIntValue(val)
				if intVal != 1 {
					t.Errorf("FirstValue() = %v, want 1", intVal)
				}
			},
		},
		{
			name: "LastValue",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				val, err := b.LastValue()
				if err != nil {
					t.Errorf("LastValue() error = %v", err)
					return
				}
				if val.GetType() != TypeInteger {
					t.Errorf("LastValue() type = %v, want %v", val.GetType(), TypeInteger)
				}
				intVal, _ := AsIntValue(val)
				if intVal != 3 {
					t.Errorf("LastValue() = %v, want 3", intVal)
				}
			},
		},
		{
			name: "AppendValue integer",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				err := b.AppendValue(NewIntVal(3))
				if err != nil {
					t.Errorf("AppendValue() error = %v", err)
					return
				}
				if len(b.Bytes()) != 3 {
					t.Errorf("AppendValue() length = %v, want 3", len(b.Bytes()))
				}
				if b.Bytes()[2] != 3 {
					t.Errorf("AppendValue() last byte = %v, want 3", b.Bytes()[2])
				}
			},
		},
		{
			name: "AppendValue binary",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				err := b.AppendValue(NewBinaryValue([]byte{3, 4}))
				if err != nil {
					t.Errorf("AppendValue() error = %v", err)
					return
				}
				expected := []byte{1, 2, 3, 4}
				if len(b.Bytes()) != len(expected) {
					t.Errorf("AppendValue() length = %v, want %v", len(b.Bytes()), len(expected))
				}
				for i, v := range expected {
					if b.Bytes()[i] != v {
						t.Errorf("AppendValue() byte[%d] = %v, want %v", i, b.Bytes()[i], v)
					}
				}
			},
		},
		{
			name: "InsertValue",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{2, 3})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				err := b.InsertValue(NewIntVal(1))
				if err != nil {
					t.Errorf("InsertValue() error = %v", err)
					return
				}
				expected := []byte{1, 2, 3}
				if len(b.Bytes()) != len(expected) {
					t.Errorf("InsertValue() length = %v, want %v", len(b.Bytes()), len(expected))
				}
				for i, v := range expected {
					if b.Bytes()[i] != v {
						t.Errorf("InsertValue() byte[%d] = %v, want %v", i, b.Bytes()[i], v)
					}
				}
			},
		},
		{
			name: "CopyPart",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3, 4})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				b.SetIndex(1)
				copied, err := b.CopyPart(2)
				if err != nil {
					t.Errorf("CopyPart() error = %v", err)
					return
				}
				copiedBin, ok := copied.(*BinaryValue)
				if !ok {
					t.Errorf("CopyPart() returned wrong type")
					return
				}
				expected := []byte{2, 3}
				if len(copiedBin.Bytes()) != len(expected) {
					t.Errorf("CopyPart() length = %v, want %v", len(copiedBin.Bytes()), len(expected))
				}
				for i, v := range expected {
					if copiedBin.Bytes()[i] != v {
						t.Errorf("CopyPart() byte[%d] = %v, want %v", i, copiedBin.Bytes()[i], v)
					}
				}
			},
		},
		{
			name: "CopyPart from advanced index exceeding remaining",
			setup: func() *BinaryValue {
				b := NewBinaryValue([]byte{1, 2, 3, 4, 5})
				b.SetIndex(2) // Point to element 3 (index 2)
				return b
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				copied, err := b.CopyPart(5) // Request 5, only 3 remain
				if err != nil {
					t.Errorf("CopyPart() unexpected error: %v", err)
					return
				}
				copiedBin, ok := copied.(*BinaryValue)
				if !ok {
					t.Errorf("CopyPart() returned wrong type")
					return
				}
				if len(copiedBin.Bytes()) != 3 {
					t.Errorf("CopyPart() length = %d, want 3", len(copiedBin.Bytes()))
				}
				// Verify correct values and no trailing zeros
				expected := []byte{3, 4, 5}
				for i, v := range expected {
					if copiedBin.Bytes()[i] != v {
						t.Errorf("CopyPart() byte[%d] = %v, want %v", i, copiedBin.Bytes()[i], v)
					}
				}
			},
		},
		{
			name: "RemoveCount",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3, 4})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				err := b.RemoveCount(2)
				if err != nil {
					t.Errorf("RemoveCount() error = %v", err)
					return
				}
				expected := []byte{3, 4}
				if len(b.Bytes()) != len(expected) {
					t.Errorf("RemoveCount() length = %v, want %v", len(b.Bytes()), len(expected))
				}
				for i, v := range expected {
					if b.Bytes()[i] != v {
						t.Errorf("RemoveCount() byte[%d] = %v, want %v", i, b.Bytes()[i], v)
					}
				}
			},
		},
		{
			name: "RemoveCount negative",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3, 4})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				err := b.RemoveCount(-1)
				if err == nil {
					t.Errorf("RemoveCount() expected error for negative count")
					return
				}
				expectedErr := "out of bounds: -1 must be non-negative"
				if err.Error() != expectedErr {
					t.Errorf("RemoveCount() error = %v, want %v", err.Error(), expectedErr)
				}
			},
		},
		{
			name: "SkipBy",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3, 4})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				b.SkipBy(2)
				if b.GetIndex() != 2 {
					t.Errorf("SkipBy() index = %v, want 2", b.GetIndex())
				}
			},
		},
		{
			name: "TakeCount",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3, 4})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				b.SetIndex(1)
				taken := b.TakeCount(2)
				takenBin, ok := taken.(*BinaryValue)
				if !ok {
					t.Errorf("TakeCount() returned wrong type")
					return
				}
				expected := []byte{2, 3}
				if len(takenBin.Bytes()) != len(expected) {
					t.Errorf("TakeCount() length = %v, want %v", len(takenBin.Bytes()), len(expected))
				}
				for i, v := range expected {
					if takenBin.Bytes()[i] != v {
						t.Errorf("TakeCount() byte[%d] = %v, want %v", i, takenBin.Bytes()[i], v)
					}
				}
			},
		},
		{
			name: "ChangeValue",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				err := b.ChangeValue(NewIntVal(99))
				if err != nil {
					t.Errorf("ChangeValue() error = %v", err)
					return
				}
				if b.Bytes()[0] != 99 {
					t.Errorf("ChangeValue() byte[0] = %v, want 99", b.Bytes()[0])
				}
			},
		},
		{
			name: "ClearSeries",
			setup: func() *BinaryValue {
				return NewBinaryValue([]byte{1, 2, 3})
			},
			testFunc: func(t *testing.T, b *BinaryValue) {
				b.ClearSeries()
				if len(b.Bytes()) != 0 {
					t.Errorf("ClearSeries() length = %v, want 0", len(b.Bytes()))
				}
				if b.GetIndex() != 0 {
					t.Errorf("ClearSeries() index = %v, want 0", b.GetIndex())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.setup()
			tt.testFunc(t, b)
		})
	}
}

func TestStringValue_SeriesInterface(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *StringValue
		testFunc func(t *testing.T, s *StringValue)
	}{
		{
			name: "FirstValue",
			setup: func() *StringValue {
				return NewStringValue("abc")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				val, err := s.FirstValue()
				if err != nil {
					t.Errorf("FirstValue() error = %v", err)
					return
				}
				if val.GetType() != TypeString {
					t.Errorf("FirstValue() type = %v, want %v", val.GetType(), TypeString)
				}
				strVal, _ := AsStringValue(val)
				if strVal.String() != "a" {
					t.Errorf("FirstValue() = %v, want 'a'", strVal.String())
				}
			},
		},
		{
			name: "LastValue",
			setup: func() *StringValue {
				return NewStringValue("abc")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				val, err := s.LastValue()
				if err != nil {
					t.Errorf("LastValue() error = %v", err)
					return
				}
				if val.GetType() != TypeString {
					t.Errorf("LastValue() type = %v, want %v", val.GetType(), TypeString)
				}
				strVal, _ := AsStringValue(val)
				if strVal.String() != "c" {
					t.Errorf("LastValue() = %v, want 'c'", strVal.String())
				}
			},
		},
		{
			name: "AppendValue",
			setup: func() *StringValue {
				return NewStringValue("ab")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				err := s.AppendValue(NewStringValue("c"))
				if err != nil {
					t.Errorf("AppendValue() error = %v", err)
					return
				}
				if s.String() != "abc" {
					t.Errorf("AppendValue() = %v, want 'abc'", s.String())
				}
			},
		},
		{
			name: "InsertValue",
			setup: func() *StringValue {
				return NewStringValue("bc")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				err := s.InsertValue(NewStringValue("a"))
				if err != nil {
					t.Errorf("InsertValue() error = %v", err)
					return
				}
				if s.String() != "abc" {
					t.Errorf("InsertValue() = %v, want 'abc'", s.String())
				}
			},
		},
		{
			name: "CopyPart",
			setup: func() *StringValue {
				return NewStringValue("abcd")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				s.SetIndex(1)
				copied, err := s.CopyPart(2)
				if err != nil {
					t.Errorf("CopyPart() error = %v", err)
					return
				}
				copiedStr, ok := copied.(*StringValue)
				if !ok {
					t.Errorf("CopyPart() returned wrong type")
					return
				}
				if copiedStr.String() != "bc" {
					t.Errorf("CopyPart() = %v, want 'bc'", copiedStr.String())
				}
			},
		},
		{
			name: "CopyPart from advanced index exceeding remaining",
			setup: func() *StringValue {
				s := NewStringValue("hello")
				s.SetIndex(2) // Point to 'l' (index 2)
				return s
			},
			testFunc: func(t *testing.T, s *StringValue) {
				copied, err := s.CopyPart(5) // Request 5, only 3 remain
				if err != nil {
					t.Errorf("CopyPart() unexpected error: %v", err)
					return
				}
				copiedStr, ok := copied.(*StringValue)
				if !ok {
					t.Errorf("CopyPart() returned wrong type")
					return
				}
				if len(copiedStr.Runes()) != 3 {
					t.Errorf("CopyPart() length = %d, want 3", len(copiedStr.Runes()))
				}
				if copiedStr.String() != "llo" {
					t.Errorf("CopyPart() = %v, want 'llo'", copiedStr.String())
				}
				// Verify no NUL characters (trailing zeros in rune form)
				for i, r := range copiedStr.Runes() {
					if r == 0 {
						t.Errorf("CopyPart() rune[%d] is NUL character", i)
					}
				}
			},
		},
		{
			name: "RemoveCount",
			setup: func() *StringValue {
				return NewStringValue("abcd")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				err := s.RemoveCount(2)
				if err != nil {
					t.Errorf("RemoveCount() error = %v", err)
					return
				}
				if s.String() != "cd" {
					t.Errorf("RemoveCount() = %v, want 'cd'", s.String())
				}
			},
		},
		{
			name: "RemoveCount negative",
			setup: func() *StringValue {
				return NewStringValue("abcd")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				err := s.RemoveCount(-1)
				if err == nil {
					t.Errorf("RemoveCount() expected error for negative count")
					return
				}
				expectedErr := "out of bounds: -1 must be non-negative"
				if err.Error() != expectedErr {
					t.Errorf("RemoveCount() error = %v, want %v", err.Error(), expectedErr)
				}
			},
		},
		{
			name: "SkipBy",
			setup: func() *StringValue {
				return NewStringValue("abcd")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				s.SkipBy(2)
				if s.GetIndex() != 2 {
					t.Errorf("SkipBy() index = %v, want 2", s.GetIndex())
				}
			},
		},
		{
			name: "TakeCount",
			setup: func() *StringValue {
				return NewStringValue("abcd")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				s.SetIndex(1)
				taken := s.TakeCount(2)
				takenStr, ok := taken.(*StringValue)
				if !ok {
					t.Errorf("TakeCount() returned wrong type")
					return
				}
				if takenStr.String() != "bc" {
					t.Errorf("TakeCount() = %v, want 'bc'", takenStr.String())
				}
			},
		},
		{
			name: "ChangeValue",
			setup: func() *StringValue {
				return NewStringValue("abc")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				err := s.ChangeValue(NewStringValue("X"))
				if err != nil {
					t.Errorf("ChangeValue() error = %v", err)
					return
				}
				if s.String() != "Xbc" {
					t.Errorf("ChangeValue() = %v, want 'Xbc'", s.String())
				}
			},
		},
		{
			name: "ClearSeries",
			setup: func() *StringValue {
				return NewStringValue("abc")
			},
			testFunc: func(t *testing.T, s *StringValue) {
				s.ClearSeries()
				if s.String() != "" {
					t.Errorf("ClearSeries() = %v, want ''", s.String())
				}
				if s.GetIndex() != 0 {
					t.Errorf("ClearSeries() index = %v, want 0", s.GetIndex())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			tt.testFunc(t, s)
		})
	}
}

func TestBlockValue_SeriesInterface(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *BlockValue
		testFunc func(t *testing.T, b *BlockValue)
	}{
		{
			name: "FirstValue",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				val, err := b.FirstValue()
				if err != nil {
					t.Errorf("FirstValue() error = %v", err)
					return
				}
				if val.GetType() != TypeInteger {
					t.Errorf("FirstValue() type = %v, want %v", val.GetType(), TypeInteger)
				}
				intVal, _ := AsIntValue(val)
				if intVal != 1 {
					t.Errorf("FirstValue() = %v, want 1", intVal)
				}
			},
		},
		{
			name: "LastValue",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				val, err := b.LastValue()
				if err != nil {
					t.Errorf("LastValue() error = %v", err)
					return
				}
				if val.GetType() != TypeInteger {
					t.Errorf("LastValue() type = %v, want %v", val.GetType(), TypeInteger)
				}
				intVal, _ := AsIntValue(val)
				if intVal != 3 {
					t.Errorf("LastValue() = %v, want 3", intVal)
				}
			},
		},
		{
			name: "AppendValue",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				err := b.AppendValue(NewIntVal(3))
				if err != nil {
					t.Errorf("AppendValue() error = %v", err)
					return
				}
				if len(b.Elements) != 3 {
					t.Errorf("AppendValue() length = %v, want 3", len(b.Elements))
				}
				lastVal, _ := AsIntValue(b.Elements[2])
				if lastVal != 3 {
					t.Errorf("AppendValue() last element = %v, want 3", lastVal)
				}
			},
		},
		{
			name: "InsertValue",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(2), NewIntVal(3)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				err := b.InsertValue(NewIntVal(1))
				if err != nil {
					t.Errorf("InsertValue() error = %v", err)
					return
				}
				if len(b.Elements) != 3 {
					t.Errorf("InsertValue() length = %v, want 3", len(b.Elements))
				}
				firstVal, _ := AsIntValue(b.Elements[0])
				if firstVal != 1 {
					t.Errorf("InsertValue() first element = %v, want 1", firstVal)
				}
			},
		},
		{
			name: "CopyPart",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3), NewIntVal(4)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				b.SetIndex(1)
				copied, err := b.CopyPart(2)
				if err != nil {
					t.Errorf("CopyPart() error = %v", err)
					return
				}
				copiedBlock, ok := copied.(*BlockValue)
				if !ok {
					t.Errorf("CopyPart() returned wrong type")
					return
				}
				if len(copiedBlock.Elements) != 2 {
					t.Errorf("CopyPart() length = %v, want 2", len(copiedBlock.Elements))
				}
				val1, _ := AsIntValue(copiedBlock.Elements[0])
				val2, _ := AsIntValue(copiedBlock.Elements[1])
				if val1 != 2 || val2 != 3 {
					t.Errorf("CopyPart() elements = [%v, %v], want [2, 3]", val1, val2)
				}
			},
		},
		{
			name: "CopyPart from advanced index exceeding remaining",
			setup: func() *BlockValue {
				b := NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3), NewIntVal(4), NewIntVal(5)})
				b.SetIndex(2) // Point to element 3
				return b
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				copied, err := b.CopyPart(5) // Request 5, only 3 remain
				if err != nil {
					t.Errorf("CopyPart() unexpected error: %v", err)
					return
				}
				copiedBlock, ok := copied.(*BlockValue)
				if !ok {
					t.Errorf("CopyPart() returned wrong type")
					return
				}
				if len(copiedBlock.Elements) != 3 {
					t.Errorf("CopyPart() length = %d, want 3", len(copiedBlock.Elements))
				}
				// Verify no nil elements
				for i, elem := range copiedBlock.Elements {
					if elem == nil {
						t.Errorf("CopyPart() element[%d] is nil", i)
					}
				}
				// Verify correct values
				expected := []int64{3, 4, 5}
				for i, elem := range copiedBlock.Elements {
					intVal, ok := AsIntValue(elem)
					if !ok || intVal != expected[i] {
						t.Errorf("CopyPart() element[%d] = %v, want %d", i, elem, expected[i])
					}
				}
			},
		},
		{
			name: "RemoveCount",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3), NewIntVal(4)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				err := b.RemoveCount(2)
				if err != nil {
					t.Errorf("RemoveCount() error = %v", err)
					return
				}
				if len(b.Elements) != 2 {
					t.Errorf("RemoveCount() length = %v, want 2", len(b.Elements))
				}
				val1, _ := AsIntValue(b.Elements[0])
				val2, _ := AsIntValue(b.Elements[1])
				if val1 != 3 || val2 != 4 {
					t.Errorf("RemoveCount() elements = [%v, %v], want [3, 4]", val1, val2)
				}
			},
		},
		{
			name: "RemoveCount negative",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3), NewIntVal(4)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				err := b.RemoveCount(-1)
				if err == nil {
					t.Errorf("RemoveCount() expected error for negative count")
					return
				}
				expectedErr := "out of bounds: -1 must be non-negative"
				if err.Error() != expectedErr {
					t.Errorf("RemoveCount() error = %v, want %v", err.Error(), expectedErr)
				}
			},
		},
		{
			name: "SkipBy",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3), NewIntVal(4)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				b.SkipBy(2)
				if b.GetIndex() != 2 {
					t.Errorf("SkipBy() index = %v, want 2", b.GetIndex())
				}
			},
		},
		{
			name: "TakeCount",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3), NewIntVal(4)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				b.SetIndex(1)
				taken := b.TakeCount(2)
				takenBlock, ok := taken.(*BlockValue)
				if !ok {
					t.Errorf("TakeCount() returned wrong type")
					return
				}
				if len(takenBlock.Elements) != 2 {
					t.Errorf("TakeCount() length = %v, want 2", len(takenBlock.Elements))
				}
				val1, _ := AsIntValue(takenBlock.Elements[0])
				val2, _ := AsIntValue(takenBlock.Elements[1])
				if val1 != 2 || val2 != 3 {
					t.Errorf("TakeCount() elements = [%v, %v], want [2, 3]", val1, val2)
				}
			},
		},
		{
			name: "ChangeValue",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				err := b.ChangeValue(NewIntVal(99))
				if err != nil {
					t.Errorf("ChangeValue() error = %v", err)
					return
				}
				val, _ := AsIntValue(b.Elements[0])
				if val != 99 {
					t.Errorf("ChangeValue() element[0] = %v, want 99", val)
				}
			},
		},
		{
			name: "ClearSeries",
			setup: func() *BlockValue {
				return NewBlockValue([]core.Value{NewIntVal(1), NewIntVal(2), NewIntVal(3)})
			},
			testFunc: func(t *testing.T, b *BlockValue) {
				b.ClearSeries()
				if len(b.Elements) != 0 {
					t.Errorf("ClearSeries() length = %v, want 0", len(b.Elements))
				}
				if b.GetIndex() != 0 {
					t.Errorf("ClearSeries() index = %v, want 0", b.GetIndex())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.setup()
			tt.testFunc(t, b)
		})
	}
}
